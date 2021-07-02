// package wave implements a RIFF/Wave (.wav) audio decoder for github.com/Edgaru089/audio.

package wave

import (
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Edgaru089/audio"
)

const (
	RIFFMainChunkSize = 12     // The size of the main RIFF chunk, at the beginning of the file.
	RIFFHeader        = "RIFF" // The "Chunk ID" in the main RIFF chunk.
	RIFFFormat        = "WAVE" // The "Format" in the main RIFF chunk.
	WaveHeader        = "fmt " // The "Subchunk ID" in the "fmt" subchunk.
	WaveFormatPCM     = 1      // The "Audio Format" in the "fmt " subchunk we want, namely PCM.
	DataHeader        = "data" // The "Subchunk ID" in the "data" subchunk.
)

// SoundFileReaderWave is a decoder for the RIFF/WAVE audio format.
type SoundFileReaderWave struct {
	file io.ReadSeeker
	info audio.SoundFileInfo

	bitsPerSample          int
	dataOffset, dataLength int64
	readOffset             int64 // current read position, relative to dataOffset
}

// SoundFileCheckWave checks if a given file is in RIFF/WAVE audio format.
//
// It only checks the "RIFF" and "WAVE" magics in the main RIFF chunk,
// i.e., it only tells if the file is definitely not in another format.
func SoundFileCheckWave(file io.ReadSeeker) (ok bool) {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return false
	}

	buf := make([]byte, RIFFMainChunkSize)
	_, err = file.Read(buf)
	if err != nil {
		return false
	}

	return RIFFHeader == string(buf[0:4]) && RIFFFormat == string(buf[8:12])
}

func init() {
	audio.RegisterSoundFileReader(
		SoundFileCheckWave,
		func() audio.SoundFileReader {
			return &SoundFileReaderWave{}
		},
	)
}

// decodes little-endian data from the first BITS bits in the slice.
func decode(data []byte, bits int) (value int64) {

	switch bits {
	case 8:
		value = int64(data[0])
	case 16:
		value = int64(data[0]) | int64(data[1])<<8
	case 24:
		value = int64(data[0]) | int64(data[1])<<8 | int64(data[2])<<16
	case 32:
		value = int64(data[0]) | int64(data[1])<<8 | int64(data[2])<<16 | int64(data[3])<<24
	}

	return
}

// reads from file, then decode the first BITS bits.
func readcode(file io.ReadSeeker, bits int) (value int64, err error) {
	var buf [8]byte
	_, err = file.Read(buf[:bits/8])
	if err != nil {
		return
	}

	return decode(buf[:], bits), nil
}

func newerror(existing error, str string) error {
	if existing != nil {
		return existing
	}
	return errors.New(str)
}

func (r *SoundFileReaderWave) Open(file io.ReadSeeker) (info audio.SoundFileInfo, err error) {
	// skip the main chunk
	file.Seek(RIFFMainChunkSize, io.SeekStart)

	// scan all subchunks
	for {
		var nerr error
		var chunkSize, chunkOffset int64

		// read the chunk header - chunk ID and chunkSize
		var name [4]byte
		_, nerr = file.Read(name[:])
		if nerr != nil {
			goto endloop
		}

		// read the chunk size (4 bytes)
		chunkSize, nerr = readcode(file, 32)
		if nerr != nil {
			goto endloop
		}

		// the chunk data offset
		chunkOffset, nerr = file.Seek(0, io.SeekCurrent)
		if nerr != nil {
			goto endloop
		}

		log.Printf(`Wave: Subchunk Name="%s", Offset=%d, Length=%d`, name, chunkOffset, chunkSize)

		switch string(name[:]) {
		case WaveHeader:
			// the "fmt " chunk

			// AudioFormat (2 bytes)
			audioFormat, nerr := readcode(file, 16)
			if nerr != nil || audioFormat != WaveFormatPCM {
				err = newerror(nerr, "Wave: Audio format error (not PCM)")
				break
			}

			// NumChannels (2 bytes)
			numChannels, nerr := readcode(file, 16)
			if nerr != nil {
				goto endloop
			}

			// SampleRate (4 bytes)
			sampleRate, nerr := readcode(file, 32)
			if nerr != nil {
				goto endloop
			}

			// skip ByteRate(4 bytes) and BlockAlign(2 bytes)
			// this should do the trick
			_, nerr = readcode(file, 48)
			if nerr != nil {
				goto endloop
			}

			// BitsPerSample (2 bytes)
			// should be 8 or 16 only
			bitsPerSample, nerr := readcode(file, 16)
			if nerr != nil || (bitsPerSample != 8 && bitsPerSample != 16) {
				err = newerror(nerr, "Wave: Audio format error (not 8 or 16 bits per sample)")
				break
			}

			info.ChannelCount = int(numChannels)
			info.SampleRate = int(sampleRate)
			r.bitsPerSample = int(bitsPerSample)

		case DataHeader:
			// the "data" chunk
			// skip the data
			_, nerr = file.Seek(chunkSize, io.SeekCurrent)
			if nerr != nil {
				goto endloop
			}

			r.dataOffset = chunkOffset
			r.dataLength = chunkSize
			info.SampleCount = chunkSize / (int64(r.bitsPerSample / 8))
		}

		// for whatever chunk, seek to the next chunk position
		file.Seek(chunkOffset+chunkSize, io.SeekStart)

		continue

	endloop: // it's driving me nuts the restrictions, and, as a common knowledge, only those who're nuts use goto
		log.Print("lunatic asylum")
		err = nerr
		break
	}

	if err != io.EOF && err != nil {
		return
	}

	if info.ChannelCount == 0 || r.dataOffset == 0 {
		return info, errors.New("Wave: Audio format error (no FMT or DATA subchunk)")
	}

	log.Printf("Wave: Asylum: info=%v, bits=%d", info, r.bitsPerSample)

	// seek to the beginning of the data
	file.Seek(r.dataOffset, io.SeekStart)
	r.readOffset = 0

	r.info = info
	r.file = file
	return info, nil
}

func (r *SoundFileReaderWave) Info() audio.SoundFileInfo {
	return r.info
}

func (r *SoundFileReaderWave) Seek(sampleOffset int64) error {
	r.readOffset = sampleOffset * int64(r.bitsPerSample/8)
	_, err := r.file.Seek(r.dataOffset+r.readOffset, io.SeekStart)
	return err
}

// reads from file BITS bits, then decode the first 16 bits.
func readcode16(file io.ReadSeeker, bits int) (val16 int16, err error) {
	var buf [8]byte
	_, err = file.Read(buf[:bits/8])
	if err != nil {
		return
	}

	var value int64
	switch bits {
	case 8:
		value = int64(buf[0]) << 8
	case 16, 24, 32:
		value = int64(buf[0]) | int64(buf[1])<<8
	}

	return int16(value), nil
}

func (r *SoundFileReaderWave) Read(data []int16) (samplesRead int64, err error) {

	log.Printf("Wave: Reading: len(data)=%d, readOffset=%d (dataLength=%d)", len(data), r.readOffset, r.dataLength)

	t := time.Now()

	for samplesRead < int64(len(data)) && r.readOffset < r.dataLength {

		data[samplesRead], err = readcode16(r.file, r.bitsPerSample)
		if err != nil {
			return
		}

		samplesRead++
		r.readOffset += int64(r.bitsPerSample / 8)

		if time.Since(t) > time.Millisecond*10 {
			fmt.Printf("\rWave: Read over: samplesRead=%d, readOffset=%d (dataLength=%d)", samplesRead, r.readOffset, r.dataLength)
			t = time.Now()
		}
	}

	log.Printf("Wave: Read over: samplesRead=%d, readOffset=%d (dataLength=%d)", samplesRead, r.readOffset, r.dataLength)

	if samplesRead == int64(len(data)) {
		// data is full, return success regardless of r.readOffset or EOF
		return
	}

	// data is not full but EOF
	return samplesRead, io.EOF
}

func (r *SoundFileReaderWave) Close() error {
	return nil
}
