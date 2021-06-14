
#include <stdint.h>
#include <stddef.h>
#include <FLAC/stream_decoder.h>


FLAC__StreamDecoderReadStatus __GoAudioFLAC_StreamRead(const FLAC__StreamDecoder*, FLAC__byte buffer[], size_t *size, void* clientData);
FLAC__StreamDecoderSeekStatus __GoAudioFLAC_StreamSeek(const FLAC__StreamDecoder*, FLAC__uint64 offset, void* clientData);
FLAC__StreamDecoderTellStatus __GoAudioFLAC_StreamTell(const FLAC__StreamDecoder*, FLAC__uint64* offset, void* clientData);
FLAC__StreamDecoderLengthStatus __GoAudioFLAC_StreamLength(const FLAC__StreamDecoder*, FLAC__uint64* length, void* clientData);

FLAC__bool __GoAudioFLAC_StreamEOF(const FLAC__StreamDecoder*, void* clientData);

FLAC__StreamDecoderWriteStatus __GoAudioFLAC_StreamWrite(
	const FLAC__StreamDecoder*,
	const FLAC__Frame* frame,
	const FLAC__int32* const buffer[],
	void* clientData
);

void __GoAudioFLAC_StreamMetadata(void* clientData, int64_t sampleCount, int32_t channelCount, int32_t sampleRate);
void __GoAudioFLAC_C_StreamMetadata(const FLAC__StreamDecoder*, const FLAC__StreamMetadata*, void* clientData);

void __GoAudioFLAC_StreamError(const FLAC__StreamDecoder*, FLAC__StreamDecoderErrorStatus status, void* clientData);

