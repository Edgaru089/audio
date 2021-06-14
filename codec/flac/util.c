
#include "util.h"
#include "callback.h"


void __GoAudioFLAC_C_InitStream(FLAC__StreamDecoder* decoder, void* clientData) {
	FLAC__stream_decoder_init_stream(
		decoder,
		&__GoAudioFLAC_StreamRead,
		&__GoAudioFLAC_StreamSeek,
		&__GoAudioFLAC_StreamTell,
		&__GoAudioFLAC_StreamLength,
		&__GoAudioFLAC_StreamEOF,
		&__GoAudioFLAC_StreamWrite,
		&__GoAudioFLAC_C_StreamMetadata,
		&__GoAudioFLAC_StreamError,
		clientData
	);
}

const char * __GoAudioFLAC_C_StreamDecoderErrorStatusString(FLAC__StreamDecoderErrorStatus status) {
	return FLAC__StreamDecoderErrorStatusString[status];
}

FLAC__int32 __GoAudioFLAC_C_IndexBuffer(void * buffer, int64_t i, int64_t j) {
	return ((FLAC__int32**)buffer)[i][j];
}

