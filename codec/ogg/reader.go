package ogg

// #include <stdint.h>
// #include <stddef.h>
// #include <stdlib.h>
// #include <vorbis/vorbisfile.h>
//
// OggVorbis_File* __GoAudioOgg_C_OpenCallbacks(void* clientData);
// const char* __GoAudioOgg_C_GetError();
import "C"
import (
	"errors"
	"io"
	"log"
	"sync"
	"unsafe"

	"github.com/Edgaru089/audio"
)

type SoundFileReaderOgg struct {
	id int

	vorbis *C.OggVorbis_File

	file io.ReadSeeker
	info audio.SoundFileInfo
}

// Magic is the magic header of the Ogg format.
// It is at the very beginning of the file.
var Magic = []byte("OggS")

// SoundFileCheckOgg is the check function of the Ogg format.
var SoundFileCheckOgg = audio.SoundFileCheckMagic(Magic, 0)

var (
	readers map[int]*SoundFileReaderOgg
	rid     int = 1
	lock    sync.RWMutex
)

func init() {
	readers = make(map[int]*SoundFileReaderOgg)

	audio.RegisterSoundFileReader(
		SoundFileCheckOgg,
		func() audio.SoundFileReader {
			lock.Lock()
			defer lock.Unlock()
			reader := &SoundFileReaderOgg{id: rid}
			readers[rid] = reader
			rid++
			return reader
		},
	)
}

func (r *SoundFileReaderOgg) Open(file io.ReadSeeker) (info audio.SoundFileInfo, err error) {
	r.file = file
	r.vorbis = C.__GoAudioOgg_C_OpenCallbacks(unsafe.Pointer(uintptr(r.id)))
	if r.vorbis == nil {
		return audio.SoundFileInfo{}, errors.New("ogg: failed to open Vorbis callbacked struct: " + C.GoString(C.__GoAudioOgg_C_GetError()))
	}

	vinfo := C.ov_info(r.vorbis, -1)
	r.info.SampleCount = int64(C.ov_pcm_total(r.vorbis, -1) * C.ogg_int64_t(vinfo.channels))
	r.info.ChannelCount = int(vinfo.channels)
	r.info.SampleRate = int(vinfo.rate)

	return r.info, nil
}

func (r *SoundFileReaderOgg) Info() audio.SoundFileInfo {
	return r.info
}

func (r *SoundFileReaderOgg) Seek(sampleOffset int64) error {
	if r.vorbis == nil {
		panic("ogg: call Seek on nil Reader")
	}

	stat := C.ov_pcm_seek(r.vorbis, C.ogg_int64_t(sampleOffset/int64(r.info.ChannelCount)))

	switch stat {
	case 0:
		return nil
	case C.OV_ENOSEEK:
		return errors.New("ogg: seek error: bitstream is not seekable")
	case C.OV_EINVAL:
		return errors.New("ogg: seek error: invalid argument value; possibly called with an OggVorbis_File structure not open")
	case C.OV_EREAD:
		return errors.New("ogg: seek error: read from media returned error")
	case C.OV_EFAULT:
		return errors.New("ogg: seek error: internal logic fault; indicates a bug or heap/stack corruption")
	case C.OV_EBADLINK:
		return errors.New("ogg: seek error: invalid stream section supplied to libvorbisfile, or the requested link is corrupt")
	default:
		return errors.New("ogg: seek error: unknown error")
	}
}

func (r *SoundFileReaderOgg) Read(data []int16) (samplesRead int64, err error) {
	if r.vorbis == nil {
		panic("ogg: call Read on nil Reader")
	}

	maxcount := int64(len(data))

	log.Printf("Ogg: READ: maxcount=%d", maxcount)

	for samplesRead < maxcount {

		bytesToRead := (maxcount - samplesRead) * int64(unsafe.Sizeof(int16(0)))
		bytesRead := int64(C.ov_read(r.vorbis, (*C.char)(unsafe.Pointer(&data[samplesRead])), C.int(bytesToRead), 0, 2, 1, nil))
		//log.Printf("Ogg: READcycle: (samplesRead=%d, maxcount=%d) BytesToRead=%d, BytesRead=%d", samplesRead, maxcount, bytesToRead, bytesRead)
		if bytesRead > 0 {
			samples := bytesRead / int64(unsafe.Sizeof(int16(0)))
			samplesRead += samples
		} else if bytesRead == 0 {
			return samplesRead, io.EOF
		} else {
			// error condition: https://xiph.org/vorbis/doc/vorbisfile/ov_read.html
			switch bytesRead {
			case C.OV_HOLE:
				return samplesRead, errors.New("ogg: Read: there was an interruption in the data (garbage between pages, loss of sync followed by recapture, or a corrupt page)")
			case C.OV_EBADLINK:
				return samplesRead, errors.New("ogg: Read: an invalid stream section was supplied to libvorbisfile, or the requested link is corrupt.")
			case C.OV_EINVAL:
				return samplesRead, errors.New("ogg: Read: initial file headers couldn't be read or are corrupt, or the initial open call for vf failed.")
			}
		}
	}

	return
}

func (r *SoundFileReaderOgg) Close() error {
	if r.vorbis != nil {
		C.ov_clear(r.vorbis)
		C.free(unsafe.Pointer(r.vorbis))
		r.vorbis = nil
	}

	lock.Lock()
	defer lock.Unlock()
	delete(readers, r.id)
	return nil
}
