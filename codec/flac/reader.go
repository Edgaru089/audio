package flac

// #include <stdint.h>
// #include <stddef.h>
// #include <FLAC/stream_decoder.h>
// #include "util.h"
import "C"
import (
	"errors"
	"io"
	"unsafe"

	"github.com/Edgaru089/audio"
)

type SoundFileReaderFLAC struct {
	decoder *C.FLAC__StreamDecoder

	file io.ReadSeeker
	info audio.SoundFileInfo

	readBuffer  []int16 // The main buffer to be written first.
	alreadyRead int     // The number of bytes already written into the buffer. Subsequent writing should happen at buffer[already].

	leftoverBuffer []int16 // Leftover samples from reading a frame that does not fit into the main buffer.

	err error
}

// Magic is the magic header of the FLAC format.
// It is at the very beginning of the file.
var Magic = []byte("fLaC")

// SoundFileCheckFLAC is the check function of the FLAC format.
var SoundFileCheckFLAC = audio.SoundFileCheckMagic(Magic, 0)

func init() {
	audio.RegisterSoundFileReader(
		SoundFileCheckFLAC,
		func() audio.SoundFileReader {
			return &SoundFileReaderFLAC{}
		},
	)
}

func (r *SoundFileReaderFLAC) Open(file io.ReadSeeker) (info audio.SoundFileInfo, err error) {
	r.decoder = C.FLAC__stream_decoder_new()
	if r.decoder == nil {
		err = errors.New("failed to open FLAC file (failed to allocate decoder)")
		return
	}

	r.file = file
	C.__GoAudioFLAC_C_InitStream(r.decoder, unsafe.Pointer(r))

	// read the header, FALSE if error
	if C.FLAC__stream_decoder_process_until_end_of_metadata(r.decoder) != 1 {
		r.Close()
		err = errors.New("failed to open FLAC file (failed to read metadata)")
	}

	return r.info, nil
}

func (r *SoundFileReaderFLAC) Info() audio.SoundFileInfo {
	return r.info
}

func (r *SoundFileReaderFLAC) Seek(sampleOffset int64) error {
	if r.decoder == nil {
		panic("flac: call Seek on nil Reader")
	}

	r.err = nil

	// clear the read buffer and the leftover
	// the seek operation will trigger a read of one frame
	r.readBuffer = nil
	r.alreadyRead = 0
	r.leftoverBuffer = r.leftoverBuffer[:0]

	// FLAC seeks expect absolute sample offset without channels
	if sampleOffset < r.info.SampleCount {
		C.FLAC__stream_decoder_seek_absolute(r.decoder, C.uint64_t(sampleOffset/int64(r.info.ChannelCount)))
		// the leftover buffer is now filled with samples from the read triggered
	} else {
		// seek to the end
		C.FLAC__stream_decoder_seek_absolute(r.decoder, C.uint64_t(r.info.SampleCount/int64(r.info.ChannelCount)-1))
		C.FLAC__stream_decoder_skip_single_frame(r.decoder)

		// clear the leftover buffer
		r.leftoverBuffer = r.leftoverBuffer[:0]
	}

	return r.err
}

func (r *SoundFileReaderFLAC) Read(data []int16) (samplesRead int64, err error) {
	if r.decoder == nil {
		panic("flac: call Read on nil Reader")
	}

	r.err = nil

	// if the leftover is not empty, use that first
	if len(r.leftoverBuffer) > 0 {
		if len(r.leftoverBuffer) > len(data) {
			// with leftover only we can fill the read request
			copy(data, r.leftoverBuffer)

			// move the remaining of the leftover to the beginning
			copy(r.leftoverBuffer, r.leftoverBuffer[len(data):])
			r.leftoverBuffer = r.leftoverBuffer[:len(r.leftoverBuffer)-len(data)]

			return int64(len(data)), nil
		} else {
			// copy all of leftover and then decode new frames
			copy(data, r.leftoverBuffer)
			r.alreadyRead = len(r.leftoverBuffer)
			r.leftoverBuffer = r.leftoverBuffer[:0]
		}
	}

	r.readBuffer = data

	for r.alreadyRead < len(r.readBuffer) {

		// it calls the write callback, everything happens there
		// this returns FALSE on fatal error (not including EOF)
		if C.FLAC__stream_decoder_process_single(r.decoder) != 1 {
			break
		}

		// break on EOF
		if C.FLAC__stream_decoder_get_state(r.decoder) == C.FLAC__STREAM_DECODER_END_OF_STREAM {
			break
		}
	}

	return int64(r.alreadyRead), r.err
}

func (r *SoundFileReaderFLAC) Close() error {
	if r.decoder != nil {
		C.FLAC__stream_decoder_finish(r.decoder)
		C.FLAC__stream_decoder_delete(r.decoder)
		r.decoder = nil
	}
	return nil
}
