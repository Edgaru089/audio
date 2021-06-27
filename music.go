package audio

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

const (
	MusicBufferLength = time.Second // the length of the internal buffer of Music
)

// musicStream satisfies SoundStreamInterface
type musicStream struct {
	music *Music
}

func (m musicStream) GetData() []int16 {
	m.music.lock.Lock()
	defer m.music.lock.Unlock()

	read, _ := m.music.file.Read(m.music.buffer)
	return m.music.buffer[:read]
}

func (m musicStream) Seek(offset time.Duration) {
	m.music.lock.Lock()
	defer m.music.lock.Unlock()
	m.music.file.Seek(int64(offset.Seconds() * float64(m.music.info.SampleRate) * float64(m.music.info.ChannelCount)))
}

// Music is a streamed sound played from a InputSoundFile.
type Music struct {
	SoundStream

	file SoundFileReader

	lock   sync.Mutex
	buffer []int16
}

func NewMusic() (m *Music) {
	m = &Music{}
	return
}

func (m *Music) Open(file io.ReadSeeker) (err error) {
	m.file = NewSoundFileReader(file)
	if m.file == nil {
		return errors.New("Music: cannot open: unknown format")
	}

	m.info, err = m.file.Open(file)
	if err != nil {
		return fmt.Errorf("Music: cannot open stream: %s", err.Error())
	}

	m.init()

	return nil
}

// init is called when the music file has changed
func (m *Music) init() {

	m.SoundStream.Init(
		musicStream{m},
		m.info,
	)

	//if m.buffer == nil {
	// allocate a second worth of buffer
	m.buffer = make([]int16, int(float64(m.info.SampleRate*m.info.ChannelCount)*MusicBufferLength.Seconds()))
	//}
}

func (m *Music) Close() {
	m.SoundStream.Close()
}
