package audio

// #include "headers.h"
import "C"
import (
	"errors"
)

var (
	alcDevice  *C.ALCdevice
	alcContext *C.ALCcontext

	listenerVolume    float32 = 100.0
	listenerPosition          = [3]float32{0, 0, 0}
	listenerDirection         = [3]float32{0, 0, -1}
	listenerUpVector          = [3]float32{0, 1, 0}
)

func initDevice() (err error) {
	alcDevice = C.alcOpenDevice(nil)
	if alcDevice == nil {
		return errors.New("failed to open audio device")
	}

	alcContext = C.alcCreateContext(alcDevice, nil)
	if alcContext == nil {
		return errors.New("failed to create audio context")
	}

	C.alcMakeContextCurrent(alcContext)

	orientation := []float32{
		listenerDirection[0], listenerDirection[1], listenerDirection[2],
		listenerUpVector[0], listenerUpVector[1], listenerUpVector[2],
	}

	C.alListenerf(C.AL_GAIN, C.float(listenerVolume*0.01))
	C.alListenerfv(C.AL_POSITION, ptrf(listenerPosition[:]))
	C.alListenerfv(C.AL_ORIENTATION, ptrf(orientation))

	return nil
}

func isExtensionSupported(name string) bool {
	cstr, free := c_str(name)
	defer free()

	if len(name) > 2 && name[:3] == "ALC" {
		// ALC extension
		return C.alcIsExtensionPresent(alcDevice, cstr) != C.AL_FALSE
	} else {
		// AL extension
		return C.alIsExtensionPresent(cstr) != C.AL_FALSE
	}
}

func getFormatFromChannelCount(channelCount int) C.ALenum {
	var format C.ALenum

	switch channelCount {
	case 1:
		format = C.AL_FORMAT_MONO16
	case 2:
		format = C.AL_FORMAT_STEREO16
	case 4:
		format = C.alGetEnumValue(c_str_const("AL_FORMAT_QUAD16\x00"))
	case 6:
		format = C.alGetEnumValue(c_str_const("AL_FORMAT_51CHN16\x00"))
	case 7:
		format = C.alGetEnumValue(c_str_const("AL_FORMAT_61CHN16\x00"))
	case 8:
		format = C.alGetEnumValue(c_str_const("AL_FORMAT_71CHN16\x00"))
	}

	// a bug on macOS
	if format == -1 {
		format = 0
	}

	return format
}

// SetGlobalVolume sets the global volume of the listener, from 0 to 100.
//
// The default is 100.
func SetGlobalVolume(volume float32) {
	C.alListenerf(C.AL_GAIN, (C.float)(volume*0.01))
	listenerVolume = volume
}

// GetGlobalVolume returns the global volume of the listener, from 0 to 100.
//
// The default is 100.
func GetGlobalVolume() float32 {
	return listenerVolume
}

// SetListenerPosition sets the position of the global listener.
//
// The default is [0, 0, 0].
func SetListenerPosition(pos [3]float32) {
	C.alListenerfv(C.AL_POSITION, ptrf(pos[:]))
	listenerPosition = pos
}

// GetListenerPosition returns the position of the global listener.
//
// The default is [0, 0, 0].
func GetListenerPosition() [3]float32 {
	return listenerPosition
}

// SetListenerDirection sets the direction the global listener is facing.
//
// The vector does not need to be normalized.
//
// the default is [0, 0, -1] (Z-Minus).
func SetListenerDirection(dir [3]float32) {
	orientation := []float32{
		dir[0], dir[1], dir[2],
		listenerUpVector[0], listenerUpVector[1], listenerUpVector[2],
	}
	C.alListenerfv(C.AL_ORIENTATION, ptrf(orientation))

	listenerDirection = dir
}

// GetListenerDirection sets the direction the global listener is facing.
//
// the default is [0, 0, -1] (Z-Minus).
func GetListenerDirection() [3]float32 {
	return listenerDirection
}

// SetListenerUpVector sets the vector the global listener is pointing up towards.
//
// The vector does not need to be normalized.
//
// the default is [0, 1, 0] (Y-Plus).
func SetListenerUpVector(up [3]float32) {
	orientation := []float32{
		listenerDirection[0], listenerDirection[1], listenerDirection[2],
		up[0], up[1], up[2],
	}
	C.alListenerfv(C.AL_ORIENTATION, ptrf(orientation))

	listenerUpVector = up
}

// GetListenerUpVector returns the vector the global listener is pointing up towards.
//
// the default is [0, 1, 0] (Y-Plus).
func GetListenerUpVector() [3]float32 {
	return listenerUpVector
}
