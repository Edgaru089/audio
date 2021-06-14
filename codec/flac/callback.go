package flac

// #include <stdint.h>
// #include <stddef.h>
// #include <stdbool.h>
// #include <FLAC/stream_decoder.h>
//
// #include "util.h"
import "C"
import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/Edgaru089/audio"
)

//export __GoAudioFLAC_StreamRead
func __GoAudioFLAC_StreamRead(
	streamDecoder uintptr,
	buffer *byte,
	size *C.size_t,
	clientData unsafe.Pointer,
) C.FLAC__StreamDecoderReadStatus {

	reader := (*SoundFileReaderFLAC)(clientData)

	// Let's construct a slice for simplcity
	bhead := &reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(buffer)),
		Len:  int(*size),
		Cap:  int(*size),
	}
	b := *((*[]byte)(unsafe.Pointer(bhead)))

	count, err := reader.file.Read(b)
	if count > 0 {
		(*size) = C.size_t(count)
		return C.FLAC__STREAM_DECODER_READ_STATUS_CONTINUE
	} else if err == io.EOF {
		return C.FLAC__STREAM_DECODER_READ_STATUS_END_OF_STREAM
	} else {
		return C.FLAC__STREAM_DECODER_READ_STATUS_ABORT
	}
}

//export __GoAudioFLAC_StreamSeek
func __GoAudioFLAC_StreamSeek(streamDecoder uintptr, offset int64, clientData unsafe.Pointer) C.FLAC__StreamDecoderSeekStatus {
	reader := (*SoundFileReaderFLAC)(clientData)
	_, err := reader.file.Seek(offset, io.SeekStart)
	if err != nil {
		return C.FLAC__STREAM_DECODER_SEEK_STATUS_ERROR
	}
	return C.FLAC__STREAM_DECODER_SEEK_STATUS_OK
}

//export __GoAudioFLAC_StreamTell
func __GoAudioFLAC_StreamTell(streamDecoder uintptr, offset *uint64, clientData unsafe.Pointer) C.FLAC__StreamDecoderTellStatus {
	reader := (*SoundFileReaderFLAC)(clientData)
	i, err := reader.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return C.FLAC__STREAM_DECODER_TELL_STATUS_ERROR
	}
	(*offset) = uint64(i)
	return C.FLAC__STREAM_DECODER_TELL_STATUS_OK
}

//export __GoAudioFLAC_StreamLength
func __GoAudioFLAC_StreamLength(streamDecoder uintptr, length *uint64, clientData unsafe.Pointer) C.FLAC__StreamDecoderLengthStatus {
	reader := (*SoundFileReaderFLAC)(clientData)

	// previous position
	prev, err := reader.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return C.FLAC__STREAM_DECODER_LENGTH_STATUS_ERROR
	}

	// seek to end
	l, err := reader.file.Seek(0, io.SeekEnd)
	if err != nil {
		return C.FLAC__STREAM_DECODER_LENGTH_STATUS_ERROR
	}

	(*length) = uint64(l)

	_, err = reader.file.Seek(prev, io.SeekStart)
	if err != nil {
		return C.FLAC__STREAM_DECODER_LENGTH_STATUS_ERROR
	}

	return C.FLAC__STREAM_DECODER_LENGTH_STATUS_OK
}

//export __GoAudioFLAC_StreamEOF
func __GoAudioFLAC_StreamEOF(streamDecoder uintptr, clientData unsafe.Pointer) C.FLAC__bool {
	reader := (*SoundFileReaderFLAC)(clientData)

	// previous position
	prev, err := reader.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 1
	}

	// seek to end
	l, err := reader.file.Seek(0, io.SeekEnd)
	if err != nil {
		return 1
	}

	// restote prev position
	_, err = reader.file.Seek(prev, io.SeekStart)
	if err != nil {
		return 1
	}

	if l == prev {
		return 1
	} else {
		return 0
	}
}

// int32_t* buffer[], returns buffer[i][j]
func readBuffer(buffer uintptr, i, j int) int32 {
	// offset buffer by i
	bufferaddr := buffer + uintptr(i)*unsafe.Sizeof(uintptr(0))
	// dereference at bufferaddr
	nextdim := *((*uintptr)(unsafe.Pointer(bufferaddr)))

	// offset nextdim by j
	bufferaddr = nextdim + uintptr(j)*unsafe.Sizeof(int32(0))
	// dereference
	return *((*int32)(unsafe.Pointer(bufferaddr)))
}

//export __GoAudioFLAC_StreamWrite
func __GoAudioFLAC_StreamWrite(
	streamDecoder uintptr,
	frame *C.FLAC__Frame,
	buffer uintptr,
	clientData unsafe.Pointer,
) C.FLAC__StreamDecoderWriteStatus {
	reader := (*SoundFileReaderFLAC)(clientData)

	// sample count in this frame
	_ = int(frame.header.blocksize * frame.header.channels)

	for i := 0; i < int(frame.header.blocksize); i++ {
		for j := 0; j < int(frame.header.channels); j++ {
			var sample int32 = 0
			// let's do some dithering!!!
			switch frame.header.bits_per_sample {
			case 8:
				sample = readBuffer(buffer, j, i) << 8
			case 16:
				sample = readBuffer(buffer, j, i) << 8
			case 24:
				sample = readBuffer(buffer, j, i) << 8
			case 32:
				sample = readBuffer(buffer, j, i) << 8
			default:
				panic(fmt.Sprint("__GoAudioFLAC_StreamWrite: unsupported bits per sample: ", frame.header.bits_per_sample))
			}

			if reader.readBuffer != nil && reader.alreadyRead < len(reader.readBuffer) {
				reader.readBuffer[reader.alreadyRead] = int16(sample)
				reader.alreadyRead++
			} else {
				reader.leftoverBuffer = append(reader.leftoverBuffer, int16(sample))
			}
		}
	}

	return 0
}

// meta.type is to be read, but type is a keyword in Go?!
// We need to wrap it around
//export __GoAudioFLAC_StreamMetadata
func __GoAudioFLAC_StreamMetadata(clientData unsafe.Pointer, sampleCount int64, channelCount, sampleRate int32) {
	reader := (*SoundFileReaderFLAC)(clientData)

	reader.info = audio.SoundFileInfo{
		SampleCount:  sampleCount,
		ChannelCount: int(channelCount),
		SampleRate:   int(sampleCount),
	}
}

//export __GoAudioFLAC_StreamError
func __GoAudioFLAC_StreamError(streamDecoder uintptr, status C.FLAC__StreamDecoderErrorStatus, clientData unsafe.Pointer) {
	reader := (*SoundFileReaderFLAC)(clientData)

	reader.err = fmt.Errorf("flac decode error: %s", C.GoString(C.__GoAudioFLAC_C_StreamDecoderErrorStatusString(status)))

}
