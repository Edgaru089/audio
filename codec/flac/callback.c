
#include "callback.h"
#include <FLAC/format.h>


void __GoAudioFLAC_C_StreamMetadata(const FLAC__StreamDecoder* decoder, const FLAC__StreamMetadata* meta, void* clientData) {

	if (meta->type == FLAC__METADATA_TYPE_STREAMINFO) {
		__GoAudioFLAC_StreamMetadata(
			clientData,
			meta->data.stream_info.total_samples * meta->data.stream_info.channels,
			meta->data.stream_info.channels,
			meta->data.stream_info.sample_rate
		);
	}
}

