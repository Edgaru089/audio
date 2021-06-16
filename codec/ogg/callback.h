
#include <stdint.h>
#include <stddef.h>
#include <vorbis/vorbisfile.h>


size_t __GoAudioOgg_Read(void* ptr, size_t size, size_t nmemb, void* clientData);
int    __GoAudioOgg_Seek(void* clientData, ogg_int64_t offset, int whence);
long   __GoAudioOgg_Tell(void* clientData);

OggVorbis_File* __GoAudioOgg_C_OpenCallbacks(void* clientData);

