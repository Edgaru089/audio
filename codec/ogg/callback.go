package ogg

// #include <stdint.h>
// #include <stddef.h>
// #include <stdbool.h>
// #include <vorbis/vorbisfile.h>
import "C"
import (
	"io"
	"reflect"
	"unsafe"
)

//export __GoAudioOgg_Read
func __GoAudioOgg_Read(
	data uintptr,
	size, nmemb C.size_t,
	clientData uintptr,
) C.size_t {

	lock.RLock()
	reader := readers[int(clientData)]
	lock.RUnlock()

	// Let's construct a slice for simplcity
	bhead := &reflect.SliceHeader{
		Data: data,
		Len:  int(size * nmemb),
		Cap:  int(size * nmemb),
	}
	b := *((*[]byte)(unsafe.Pointer(bhead)))

	count, _ := reader.file.Read(b)
	if count > 0 {
		return C.size_t(count)
	} else {
		return 0
	}
}

//export __GoAudioOgg_Seek
func __GoAudioOgg_Seek(clientData uintptr, offset C.ogg_int64_t, whence C.int) C.int {
	lock.RLock()
	reader := readers[int(clientData)]
	lock.RUnlock()

	var goWhence int
	switch whence {
	case C.SEEK_SET:
		goWhence = io.SeekStart
	case C.SEEK_CUR:
		goWhence = io.SeekCurrent
	case C.SEEK_END:
		goWhence = io.SeekEnd
	}

	pos, err := reader.file.Seek(int64(offset), goWhence)
	if err != nil {
		return -1
	}
	return C.int(pos)
}

//export __GoAudioOgg_Tell
func __GoAudioOgg_Tell(clientData uintptr) C.long {
	lock.RLock()
	reader := readers[int(clientData)]
	lock.RUnlock()

	i, err := reader.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1
	}
	return C.long(i)
}
