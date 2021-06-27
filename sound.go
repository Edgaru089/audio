package audio

// #include "headers.h"
import "C"
import (
	"runtime"
	"time"
)

type Sound struct {
	soundSource
	buffer *SoundBuffer
}

// NewSound creates a new empty Sound instance.
func NewSound() *Sound {
	s := &Sound{}
	s.soundSource.init()

	runtime.SetFinalizer(
		s,
		func(sound *Sound) {
			C.alSourcei(sound.source, C.AL_BUFFER, 0)
			sound.soundSource.close()
		},
	)

	return s
}

// SetBuffer sets the underlying buffer of the sound.
func (s *Sound) SetBuffer(buf *SoundBuffer) {
	s.buffer = buf
	C.alSourcei(s.source, C.AL_BUFFER, C.ALint(buf.buffer))
}

// Buffer returns the underlying buffer of the sound.
func (s *Sound) Buffer() *SoundBuffer {
	return s.buffer
}

// Play starts or resumes playing the sound.
func (s *Sound) Play() {
	C.alSourcePlay(s.source)
}

// Pause pauses the sound.
func (s *Sound) Pause() {
	C.alSourcePause(s.source)
}

// Stop stops playing the sound.
func (s *Sound) Stop() {
	C.alSourceStop(s.source)
}

// PlayingOffset returns the playing position of the sound in time.
func (s *Sound) PlayingOffset() time.Duration {
	var secs C.ALfloat
	C.alGetSourcef(s.source, C.AL_SEC_OFFSET, &secs)

	return time.Duration(float64(time.Second) * float64(secs))
}

// SetPlayingOffset changes the playing position of the sound.
//
// It can be called when the sound is playing or paused.
// Calling on a stopped sound has no effect.
func (s *Sound) SetPlayingOffset(offset time.Duration) {
	C.alSourcef(s.source, C.AL_SEC_OFFSET, C.float(offset.Seconds()))
}
