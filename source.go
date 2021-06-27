package audio

// #include "headers.h"
import "C"

// Enum object describing the play state of a sound.
type PlayStatus int8

const (
	Stopped PlayStatus = iota // The sound is stopped (i.e., not playing)
	Paused                    // The sound is paused
	Playing                   // The sound is currently playing
)

// Most of the comments below are directly copied from SFML.
type soundSource struct {
	source C.ALuint
}

func (s *soundSource) init() {
	if s.source == 0 {
		C.alGenSources(1, &s.source)
		C.alSourcei(s.source, C.AL_BUFFER, 0)
	}
}

func (s *soundSource) close() {
	if s.source != 0 {
		C.alDeleteSources(1, &s.source)
	}
}

// SetPitch sets the pitch of the sound.
//
// The pitch represents the perceived fundamental frequency
// of a sound; thus you can make a sound more acute or grave
// by changing its pitch. A side effect of changing the pitch
// is to modify the playing speed of the sound as well.
//
// The default value for the pitch is 1.
func (s *soundSource) SetPitch(pitch float32) {
	C.alSourcef(s.source, C.AL_PITCH, C.float(pitch))
}

// SetVolume sets the volume of the sound.
//
// The volume is a value between 0 (mute) and 100 (full volume).
//
// The default value for the volume is 100.
func (s *soundSource) SetVolume(volume float32) {
	C.alSourcef(s.source, C.AL_GAIN, C.float(volume*0.01))
}

// SetPosition sets the 3D position of the sound in the audio scene.
//
// Only sounds with one channel (mono sounds) can be
// spatialized.
//
// The default position of a sound is (0, 0, 0).
func (s *soundSource) SetPosition(pos [3]float32) {
	C.alSourcefv(s.source, C.AL_POSITION, ptrf(pos[:]))
}

// SetRelativeToListener makes the sound's position relative to the listener or absolute.
//
// Making a sound relative to the listener will ensure that it will always
// be played the same way regardless of the position of the listener.
// This can be useful for non-spatialized sounds, sounds that are
// produced by the listener, or sounds attached to it.
//
// The default value is false (position is absolute).
func (s *soundSource) SetRelativeToListener(relative bool) {
	if relative {
		C.alSourcei(s.source, C.AL_SOURCE_RELATIVE, 1)
	} else {
		C.alSourcei(s.source, C.AL_SOURCE_RELATIVE, 0)
	}
}

// SetMinDistance sets the minimum distance of the sound.
//
// The "minimum distance" of a sound is the maximum
// distance at which it is heard at its maximum volume. Further
// than the minimum distance, it will start to fade out according
// to its attenuation factor. A value of 0 ("inside the head
// of the listener") is an invalid value and is forbidden.
//
// The default value of the minimum distance is 1.
func (s *soundSource) SetMinDistance(distance float32) {
	C.alSourcef(s.source, C.AL_REFERENCE_DISTANCE, C.float(distance))
}

// SetAttenuation sets the attenuation factor of the sound.
//
// The attenuation is a multiplicative factor which makes
// the sound more or less loud according to its distance
// from the listener. An attenuation of 0 will produce a
// non-attenuated sound, i.e. its volume will always be the same
// whether it is heard from near or from far. On the other hand,
// an attenuation value such as 100 will make the sound fade out
// very quickly as it gets further from the listener.
//
// The default value of the attenuation is 1.
func (s *soundSource) SetAttenuation(attenuation float32) {
	C.alSourcef(s.source, C.AL_ROLLOFF_FACTOR, C.float(attenuation))
}

// Status returns the current status of the sound stream.
func (s *soundSource) Status() PlayStatus {
	var status C.ALint
	C.alGetSourcei(s.source, C.AL_SOURCE_STATE, &status)

	switch status {
	case C.AL_INITIAL, C.AL_STOPPED:
		return Stopped
	case C.AL_PAUSED:
		return Paused
	case C.AL_PLAYING:
		return Playing
	}

	return Stopped
}
