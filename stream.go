package audio

import (
	"log"
	"sync"
	"time"
	"unsafe"
)

// #include "headers.h"
import "C"

const (
	SoundStreamBufferCount  = 3                     // number of audio buffers used by the stream thread
	SoundStreamRetries      = 2                     // number of retries (not counting first try) for GetData()
	SoundStreamPollInterval = 50 * time.Millisecond // interval between stream thread polling
)

// SoundStreamInterface wraps underlying streamed audio resource.
type SoundStreamInterface interface {
	// SoundGetData requests a new chunk of audio samples from the stream source.
	//
	// This function is called to provide the audio samples to play.
	// It is called continuously by the streaming loop, in a separate goroutine.
	//
	// The data inside the slice is copied, meaning that you can use the same buffer
	// over and over again.
	//
	// The source can choose to stop the streaming loop at any time, by
	// returning an empty slice.
	GetData() []int16

	// Seek changes the current playing position of the stream source.
	//
	// If the sound stream in question does not support seeking, this function
	// should do nothing.
	Seek(time time.Duration)

	// SeekSample changes the current playing position of the stream source.
	//
	// Unlike Seek, this function seeks by the sample offset.
	//
	// If the sound stream in question does not support seeking, this function
	// should do nothing.
	//SeekSample(offset int64)
}

// SoundStream implements a basis for streamed audio content.
type SoundStream struct {
	soundSource

	info   SoundFileInfo
	format C.ALenum
	iface  SoundStreamInterface

	// this group is mutex protected
	lock       sync.Mutex
	state      PlayStatus
	streaming  bool
	buffers    [SoundStreamBufferCount]C.ALuint // buffer handles
	stopped    chan struct{}
	seekOffset int64
}

// Init is called by derived classes to initialize the sound stream.
func (s *SoundStream) Init(iface SoundStreamInterface, info SoundFileInfo) {

	s.soundSource.init()
	s.format = getFormatFromChannelCount(info.ChannelCount)
	s.info = info
	s.iface = iface
}

// Play starts/resumes playing the sound stream.
//
// It restarts the stream from the beginning if it is already playing.
func (s *SoundStream) Play() {
	if s.source == 0 {
		panic("SoundStream: call of nil object on Play()")
	}

	s.lock.Lock()
	streaming := s.streaming
	state := s.state
	s.lock.Unlock()

	if streaming {

		if state == Paused {
			s.lock.Lock()
			s.state = Playing
			s.lock.Unlock()
			C.alSourcePlay(s.source)
			return
		} else if state == Playing {
			// stop the stream and start it again
			s.Stop()
		}
	}

	s.streaming = true
	s.state = Playing
	go s.streamData()
}

// Pause pauses the sound stream if playing.
func (s *SoundStream) Pause() {
	s.lock.Lock()
	if !s.streaming { // the goroutine is not running
		s.lock.Unlock()
		return
	}
	s.lock.Unlock()

	C.alSourcePause(s.source)
}

// Stop stops the sound streaming if playing.
func (s *SoundStream) Stop() {
	log.Print("stopping")

	// signal and wait for the thread to terminate
	var streaming bool
	s.lock.Lock()

	if s.streaming {
		streaming = true
		s.stopped = make(chan struct{})
	}
	s.streaming = false

	s.lock.Unlock()

	if streaming {
		log.Println("waiting for stop")
		<-s.stopped
		close(s.stopped)
		log.Println("waiting for stop: ok")
	}

	s.seekOffset = 0

	if s.iface != nil {
		s.iface.Seek(0)
	}

	log.Print("stopped")
}

// SampleCount returns the number of samples in the buffer.
//
// Two samples from two channels at the same timepoint count twice.
func (s *SoundStream) SampleCount() int64 {
	return s.info.SampleCount
}

// SampleRate returns the sample rate of the buffer, in samples per second.
func (s *SoundStream) SampleRate() int {
	return s.info.SampleRate
}

// ChannelCount returns the number of channels in the buffer.
func (s *SoundStream) ChannelCount() int {
	return s.info.ChannelCount
}

// Duration returns the duration of the sound in the buffer.
func (s *SoundStream) Duration() time.Duration {
	return time.Duration(float64(time.Second) / float64(s.info.ChannelCount) * float64(s.info.SampleCount) / float64(s.info.SampleRate))
}

// Status returns the current status of the sound stream.
func (s *SoundStream) Status() PlayStatus {
	status := s.soundSource.Status()

	// To compensate for the lag between play() and alSourceplay()
	if status == Stopped {
		s.lock.Lock()
		if s.streaming {
			status = s.state
		}
		s.lock.Unlock()
	}

	return status
}

// Close closes the SoundStream. It also ends the goroutine tied to it.
//
// The stream object should not be used again.
func (s *SoundStream) Close() {
	s.Stop()
}

// PlayingOffset returns the playing position of the sound in time.
func (s *SoundStream) PlayingOffset() time.Duration {
	if s.source == 0 {
		return 0
	}

	var secs C.ALfloat
	C.alGetSourcef(s.source, C.AL_SEC_OFFSET, &secs)

	return time.Duration(float64(time.Second) * (float64(secs) + float64(s.seekOffset)/float64(s.info.ChannelCount)/float64(s.info.SampleRate)))
}

// SetPlayingOffset changes the playing position of the sound.
//
// It can be called when the sound is playing or paused.
// Calling on a stopped sound has no effect.
func (s *SoundStream) SetPlayingOffset(offset time.Duration) {
	log.Print("seeking")

	// stop the streaming, seek, and then start streaming again

	oldstatus := s.Status()
	if oldstatus == Stopped {
		return
	}

	log.Printf("Old status %d (Stopped=0, Paused=1, Playing=2)", oldstatus)

	s.Stop()

	s.iface.Seek(offset)

	s.streaming = true
	s.state = oldstatus
	s.seekOffset = int64(offset.Seconds() * float64(s.info.ChannelCount) * float64(s.info.SampleRate))
	go s.streamData()
	log.Print("seek ok")
}

func (s *SoundStream) streamData() {
	log.Print("log: running")

	var wantstop bool

	// return if the thread is launched stopped
	s.lock.Lock()
	if s.state == Stopped {
		s.lock.Unlock()
		return
	}
	s.lock.Unlock()

	// create the buffers
	C.alGenBuffers(SoundStreamBufferCount, &s.buffers[0])

	// fill the queue
	wantstop = s.fillQueue()

	// play the sound
	C.alSourcePlay(s.source)

	// check if the thread is launched paused
	s.lock.Lock()
	if s.state == Paused {
		C.alSourcePause(s.source)
	}
	s.lock.Unlock()

	for {
		s.lock.Lock()
		if !s.streaming {
			log.Print("breaking loop")
			s.lock.Unlock()
			break
		}
		s.lock.Unlock()

		// interrupted
		if s.Status() == Stopped {
			if !wantstop {
				log.Print("loop: stopped: !wantstop, continuing")
				// just continue
				C.alSourcePlay(s.source)
			} else {
				log.Print("loop: stopped: wantstop, ending")
				// end streaming
				s.lock.Lock()
				s.streaming = false
				s.lock.Unlock()
				break
			}
		}

		var numProcessed C.ALint
		C.alGetSourcei(s.source, C.AL_BUFFERS_PROCESSED, &numProcessed)

		for i := 0; i < int(numProcessed); i++ {
			// pop the first (processed) buffer from the queue
			var buffer C.ALuint
			C.alSourceUnqueueBuffers(s.source, 1, &buffer)

			// find its number
			var buffernum int
			for i := 0; i < SoundStreamBufferCount; i++ {
				if s.buffers[i] == buffer {
					buffernum = i
					break
				}
			}

			// add the processed sample size to the offset
			var size, bits C.ALint
			C.alGetBufferi(buffer, C.AL_SIZE, &size)
			C.alGetBufferi(buffer, C.AL_BITS, &bits)
			s.lock.Lock()
			s.seekOffset += int64(size / (bits / 8))
			s.lock.Unlock()

			// fill and push the buffer again
			if !wantstop {
				if s.fillAndPushBuffer(buffernum) {
					wantstop = true
				}
			}

		}

		// sleep for a while
		if s.Status() != Stopped {
			time.Sleep(SoundStreamPollInterval)
		}
	}

	// stop playback
	C.alSourceStop(s.source)

	// pop anything left in the queue
	s.clearQueue()

	// delete the buffers
	C.alSourcei(s.source, C.AL_BUFFER, 0)
	C.alDeleteBuffers(SoundStreamBufferCount, &s.buffers[0])

	log.Print("cleanup ok")

	// signal stopped
	var hasStopped bool
	hasStopped = true
	s.lock.Lock()
	if s.stopped != nil {
		hasStopped = true
	}
	s.lock.Unlock()

	if hasStopped {
		log.Println("thread signaling")
		s.stopped <- struct{}{}
		log.Println("thread signaled")
	}
	log.Println("thread exiting")
}

// returns true if the new buffer reaches end of file
func (s *SoundStream) fillAndPushBuffer(num int) bool {
	var wantstop bool

	var data []int16
	for retries := 0; retries <= SoundStreamRetries; retries++ {
		data = s.iface.GetData()
		if len(data) > 0 {
			// got data; stop trying
			break
		}
	}

	if len(data) > 0 {
		C.alBufferData(
			s.buffers[num],
			s.format,
			unsafe.Pointer(&data[0]),
			C.ALsizei(uintptr(len(data))*unsafe.Sizeof(int16(0))),
			C.ALsizei(s.info.SampleRate),
		)
		C.alSourceQueueBuffers(s.source, 1, &s.buffers[num])
	} else {
		wantstop = true
	}

	log.Printf("fillAndPush: #%d, len=%d", num, len(data))

	if wantstop {
		log.Print("fillAndPush: wantStop")
	}

	return wantstop
}

// returns true if the queue reaches end of file
func (s *SoundStream) fillQueue() bool {
	var wantstop bool
	for i := 0; i < SoundStreamBufferCount; i++ {
		if s.fillAndPushBuffer(i) {
			wantstop = true
			break
		}
	}

	return wantstop
}

func (s *SoundStream) clearQueue() {

	var n C.ALint
	C.alGetSourcei(s.source, C.AL_BUFFERS_QUEUED, &n)

	// dequeue all of them
	var buffer C.ALuint
	for i := 0; i < int(n); i++ {
		C.alSourceUnqueueBuffers(s.source, 1, &buffer)
	}

}
