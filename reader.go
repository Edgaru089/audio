package audio

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

// SoundFileInfo contains the properities of a audio file.
type SoundFileInfo struct {
	SampleCount  int64 // Total numbers of samples in the file, 2 channels count as 2
	ChannelCount int   // Numbers of channels in the file
	SampleRate   int   // Sample rate of the file, in samples per second
}

func (s SoundFileInfo) String() string {
	if s.SampleRate == 0 {
		return fmt.Sprintf("[SampleCount=%d, SampleRate=%d, ChannelCount=%d]", s.SampleCount, s.SampleRate, s.ChannelCount)
	}
	return fmt.Sprintf(
		"[SampleCount=%d, SampleRate=%d (Duration=%v), ChannelCount=%d]",
		s.SampleCount,
		s.SampleRate,
		time.Duration(float64(time.Second)*float64(s.SampleCount)/float64(s.ChannelCount)/float64(s.SampleRate)),
		s.ChannelCount,
	)
}

// SoundFileCheck is called to tell if the given file can be handled by a codec.
//
// The file is seeked to the beginning when the function is called.
type SoundFileCheck func(file io.ReadSeeker) (ok bool)

// SoundFileCheckMagic returns a function that returns true if the file begins
// with the given magic string at the given offset.
func SoundFileCheckMagic(magic []byte, offset int64) SoundFileCheck {
	return func(file io.ReadSeeker) (ok bool) {
		_, err := file.Seek(offset, io.SeekStart)
		if err != nil {
			return false
		}

		buf := make([]byte, len(magic))
		_, err = file.Read(buf)
		if err != nil {
			return false
		}

		return bytes.Equal(buf, magic)
	}
}

// SoundFileReader is a interface to be implemented by sound file decoders.
//
// Sound file codecs also need to implement the SoundFileCheck function.
type SoundFileReader interface {

	// Open opens a file stream for future decoding.
	//
	// The file stream is already seeked to the beginning, i.e.,
	// have file.Seek(0, io.SeekStart) called.
	//
	// A single SoundFileReader instance should only call Open once.
	Open(file io.ReadSeeker) (SoundFileInfo, error)

	// Info returns the properities of the sound stream.
	Info() SoundFileInfo

	// Seek changes the read position to the given offset, relative to the beginning of the file.
	//
	// The sampleOffset can be computed from Time offset with the given formula:
	//     timeInSeconds * sampleRate * channelCount
	//
	// If sampleOffset is greater than the number of samples in the file,
	// this function must jump to the end of the file.
	Seek(sampleOffset int64) error

	// Read reads audio samples from the open file.
	//
	// The read data is written into the len() part of the data slice.
	//
	// The returned number of samples read may be smaller than len(data).
	// This should not be considered as an error only on EOF.
	//
	// On end-of-file, Read should return (0, io.EOF).
	Read(data []int16) (samplesRead int64, err error)

	// The reader needs to be closed after use.
	io.Closer
}

var (
	fileReaders []struct {
		alloc func() SoundFileReader
		check SoundFileCheck
	}
)

// RegisterSoundFileReader registers a new SoundFileReader.
//
// the allocator function allocates a new instance, it should look like
//     func () { return &Decoder{} }
func RegisterSoundFileReader(check SoundFileCheck, allocator func() SoundFileReader) {
	fileReaders = append(fileReaders, struct {
		alloc func() SoundFileReader
		check SoundFileCheck
	}{
		alloc: allocator,
		check: check,
	})
}

// NewSoundFileReader creates a new SoundFileReader from the registered codecs.
//
// It DOES NOT call reader.Open().
//
// It returns nil if no matching SoundFileReader is found.
func NewSoundFileReader(file io.ReadSeeker) SoundFileReader {

	for _, r := range fileReaders {
		file.Seek(0, io.SeekStart)
		if r.check(file) {
			file.Seek(0, io.SeekStart)
			return r.alloc()
		}
	}

	return nil
}
