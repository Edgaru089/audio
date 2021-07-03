//+build amd64 386

package wave

import (
	"reflect"
	"unsafe"
)

// reads 16-bit raw PCM data from the 16-bit raw PCM file
//
// on little-endian systems, this requires only a simple memory copy
func (r *SoundFileReaderWave) read16(data []int16) (samplesRead int64, err error) {

	canread := r.dataLength - r.readOffset

	var toreadBytes int64
	if int64(len(data))*2 < canread {
		toreadBytes = int64(len(data)) * 2
	} else {
		toreadBytes = canread
	}

	// dirty hacks, but lighting fast!
	dataSlice := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	buildSlice := reflect.SliceHeader{
		Data: dataSlice.Data,
		Len:  dataSlice.Len * 2,
		Cap:  dataSlice.Cap * 2,
	}

	readlen, err := r.file.Read((*(*[]byte)(unsafe.Pointer(&buildSlice)))[:toreadBytes])

	r.readOffset += int64(readlen)

	if readlen/2 == len(data) {
		return int64(readlen / 2), nil
	} else {
		return int64(readlen / 2), err
	}
}
