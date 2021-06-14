### Package audio

The audio package implements audio decoding and playback in a manner mirroring the Audio module of SFML.

SFML stands for Simple and Fast Multimedia Library, a audio/graphics/windowing/networking library in C++. It is licensed under the Zlib/png license. I like it personally, being both fast, reliable and elegant in design.

Under the hood, the audio package wraps OpenAL for playback, and libflac and libvorbis for decoding of FLAC and Ogg/Vorbis audio files.
