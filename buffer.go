package audio

// #include "headers.h"
import "C"
import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"time"
	"unsafe"
)

// SoundBuffer holds sound sample data, together
// with a OpenAL resource tag pointing to the buffer.
//
// SoundBuffers load all the sound data uncompressed into memory,
// so they tend to occupy a lot of room.
type SoundBuffer struct {
	buffer   C.ALuint // OpenAL buffer handle
	samples  []int16
	info     SoundFileInfo
	duration time.Duration
}

func NewSoundBuffer() *SoundBuffer {
	b := &SoundBuffer{}
	C.alGenBuffers(1, &b.buffer)

	runtime.SetFinalizer(b, func(buf *SoundBuffer) {
		C.alDeleteBuffers(1, &buf.buffer)
	})

	return b
}

// Samples returns the internal samples buffer.
func (b *SoundBuffer) Samples() []int16 {
	return b.samples
}

// SampleCount returns the number of samples in the buffer.
//
// Two samples from two channels at the same timepoint count twice.
func (b *SoundBuffer) SampleCount() int64 {
	return b.info.SampleCount
}

// SampleRate returns the sample rate of the buffer, in samples per second.
func (b *SoundBuffer) SampleRate() int {
	return b.info.SampleRate
}

// ChannelCount returns the number of channels in the buffer.
func (b *SoundBuffer) ChannelCount() int {
	return b.info.ChannelCount
}

// Duration returns the duration of the sound in the buffer.
func (b *SoundBuffer) Duration() time.Duration {
	return b.duration
}

// Load loads the sound buffer with the given file.
//
// Please, please don't load it again if a Sound is already using it.
// This may cause weird issues I cannot diagnose.
func (b *SoundBuffer) Load(file io.ReadSeeker) (err error) {
	reader := NewSoundFileReader(file)
	if reader == nil {
		return errors.New("SoundBuffer: cannot load: unknown format")
	}

	b.info, err = reader.Open(file)
	if err != nil {
		return fmt.Errorf("SoundBuffer: cannot open stream: %s", err.Error())
	}

	// FIXME: SoundBuffer internal buffer reallocated on every Load
	b.samples = make([]int16, b.info.SampleCount)
	_, err = reader.Read(b.samples)
	if err != nil {
		return err
	}

	return b.update()
}

// update updates the OpenAL state of the buffer after samples change.
func (b *SoundBuffer) update() error {
	if len(b.samples) == 0 {
		return errors.New("SoundBuffer: OpenAL update on empty samples")
	}

	format := getFormatFromChannelCount(b.info.ChannelCount)
	if format == 0 {
		return fmt.Errorf("SoundBuffer: failed to load (unsupported number of channels: %d)", b.info.ChannelCount)
	}

	C.alBufferData(
		b.buffer,
		format,
		unsafe.Pointer(&b.samples[0]),
		C.ALsizei(uintptr(len(b.samples))*unsafe.Sizeof(int16(0))),
		C.ALsizei(b.info.SampleRate),
	)

	// calculate the duration here
	b.duration = time.Duration(float64(time.Second) * float64(b.info.SampleCount) / float64(b.info.ChannelCount) / float64(b.info.SampleRate))

	return nil
}
