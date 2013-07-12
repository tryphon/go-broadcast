package broadcast

import (
	"sync"
)

type MutexAudioBuffer struct {
	Buffer AudioBuffer

	mutex sync.Mutex
}

func (pseudoBuffer *MutexAudioBuffer) AudioOut(audio *Audio) {
	pseudoBuffer.mutex.Lock()
	defer pseudoBuffer.mutex.Unlock()

	pseudoBuffer.Buffer.AudioOut(audio)
}

func (pseudoBuffer *MutexAudioBuffer) Read() (audio *Audio) {
	pseudoBuffer.mutex.Lock()
	defer pseudoBuffer.mutex.Unlock()

	return pseudoBuffer.Buffer.Read()
}

func (pseudoBuffer *MutexAudioBuffer) SampleCount() uint32 {
	return pseudoBuffer.Buffer.SampleCount()
}
