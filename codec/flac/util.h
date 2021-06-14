
#include <stdint.h>
#include <stddef.h>
#include <FLAC/stream_decoder.h>


void __GoAudioFLAC_C_InitStream(FLAC__StreamDecoder* decoder, void* clientData);

const char * __GoAudioFLAC_C_StreamDecoderErrorStatusString(FLAC__StreamDecoderErrorStatus status);

FLAC__int32 __GoAudioFLAC_C_IndexBuffer(void * buffer, int64_t i, int64_t j);

