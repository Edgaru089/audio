## Package audio

The audio package implements audio decoding and playback in a manner mirroring the Audio module of SFML.

SFML stands for Simple and Fast Multimedia Library, a audio/graphics/windowing/networking library in C++. It is licensed under the Zlib/png license. I like it personally, being both fast, reliable and elegant in design.

Under the hood, the audio package wraps OpenAL for playback, and libFLAC, libVorbis and libOgg for decoding of FLAC and Ogg/Vorbis audio files.

```go
import "github.com/Edgaru089/audio"
import _ "github.com/Edgaru089/audio/codec/flac"


file, err := os.Open("Alstroemeria Records - Bad Apple!!.flac")

b := audio.NewSoundBuffer() // a large buffer keeping all the audio samples in memory
err := b.Load(file)

s := audio.NewSound() // the player of the sound buffer
s.SetBuffer(b)
s.Play()

m := audio.NewMusic() // a streaming audio object, keeping only a small piece of samples
err = m.Open(file)
m.Play()
```

### Linking on Windows

In the extlib folder there are headers and library for mingw in 32 and 64 bits. The OpenAL part is linked dynamically and the file openal32.dll must be copied with the executable. libFLAC, libVorbis and libOgg are linked statically.

Things should work just fine.

