
#include "callback.h"
#include <stdlib.h>
#include <stdio.h>
#include <vorbis/vorbisfile.h>


static ov_callbacks __GoAudioOgg_C_Callbacks = {
	(size_t (*)(void *, size_t, size_t, void *)) __GoAudioOgg_Read,
	(int (*)(void *, ogg_int64_t, int))          __GoAudioOgg_Seek,
	(int (*)(void *))                            NULL,
	(long (*)(void *))                           __GoAudioOgg_Tell
};
static const char* __GoAudioOgg_C_Error;

const char* __GoAudioOgg_C_GetError() {
	return __GoAudioOgg_C_Error;
}

// this function allocates the OggVorbis_File struct using malloc.
// it must be freed.
OggVorbis_File* __GoAudioOgg_C_OpenCallbacks(void* clientData) {
	OggVorbis_File* file = malloc(sizeof(OggVorbis_File));

	int status = ov_open_callbacks(clientData, file, NULL, 0, __GoAudioOgg_C_Callbacks);
	if (status < 0) {

		switch (status) {
		case OV_EREAD:
			__GoAudioOgg_C_Error = "a read from media returned an error"; break;
		case OV_ENOTVORBIS:
			__GoAudioOgg_C_Error = "bitstream does not contain any Vorbis data"; break;
		case OV_EVERSION:
			__GoAudioOgg_C_Error = "Vorbis version mismatch"; break;
		case OV_EBADHEADER:
			__GoAudioOgg_C_Error = "invalid Vorbis bitstream header"; break;
		case OV_EFAULT:
			__GoAudioOgg_C_Error = "internal logic fault; indicates a bug or heap/stack corruption"; break;
		}

		free(file);
		return NULL;
	}

	return file;
}

