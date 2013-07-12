package broadcast

import (
	"testing"
)

func TestAudioBuffer_NewAudioBuffer_default(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	if audioBuffer == nil {
		t.Errorf("Should return an AudioBuffer")
	}
}

func TestAudioBuffer_NewAudioBuffer_empty(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	if !audioBuffer.Empty() {
		t.Errorf("Should return an empty AudioBuffer")
	}
}

func TestAudioBuffer_AudioOut_sampleCount(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	audio := &Audio{sampleCount: 1024}

	audioBuffer.AudioOut(audio)
	if audioBuffer.SampleCount() != uint32(audio.SampleCount()) {
		t.Errorf("Wrong buffer SampleCount after adding Audio:\n got: %d\nwant: %d", audioBuffer.SampleCount(), audio.SampleCount())
	}
}

func TestAudioBuffer_AudioOut_nextFreeIndex(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	audio := &Audio{sampleCount: 1024}

	audioBuffer.AudioOut(audio)
	if audioBuffer.nextFreeIndex != 1 {
		t.Errorf("Wrong nextFreeIndex after adding Audio:\n got: %d\nwant: %d", audioBuffer.nextFreeIndex, 1)
	}
}

func TestAudioBuffer_AudioOut_nextFreeIndexLoop(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	audio := &Audio{sampleCount: 1024}

	for time := 0; uint32(time) < maxAudioBufferSize; time++ {
		audioBuffer.AudioOut(audio)
	}

	if audioBuffer.nextFreeIndex != 0 {
		t.Errorf("Wrong nextFreeIndex after adding %d Audios:\n got: %d\nwant: %d", maxAudioBufferSize, audioBuffer.nextFreeIndex, 0)
	}
}

func TestAudioBuffer_AudioOut_moveReader(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	audio := &Audio{sampleCount: 1024}

	for time := 0; uint32(time) < maxAudioBufferSize+1; time++ {
		audioBuffer.AudioOut(audio)
	}

	if audioBuffer.readIndex != 1 {
		t.Errorf("Wrong readIndex after adding %d Audios:\n got: %d\nwant: %d", maxAudioBufferSize+1, audioBuffer.readIndex, 1)
	}
}

func TestAudioBuffer_Read_return(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	audio := &Audio{sampleCount: 1024}

	audioBuffer.AudioOut(audio)
	readAudio := audioBuffer.Read()

	if readAudio != audio {
		t.Errorf("Wrong read Audio:\n got: %v\nwant: %v", readAudio, audio)
	}
}

func TestAudioBuffer_Read_SampleCount(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	audio := &Audio{sampleCount: 1024}

	audioBuffer.AudioOut(audio)
	audioBuffer.Read()

	if audioBuffer.SampleCount() != 0 {
		t.Errorf("Wrong SampleCount after read Audio:\n got: %d\nwant: %d", audioBuffer.SampleCount(), 0)
	}
}

func TestAudioBuffer_Read_readIndex(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	audio := &Audio{sampleCount: 1024}

	audioBuffer.AudioOut(audio)
	audioBuffer.Read()

	if audioBuffer.readIndex != 1 {
		t.Errorf("Wrong readIndex after read Audio:\n got: %d\nwant: %d", audioBuffer.readIndex, 1)
	}
}

func TestAudioBuffer_Read_readIndexLoop(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	audio := &Audio{sampleCount: 1024}

	for time := 0; uint32(time) < maxAudioBufferSize; time++ {
		audioBuffer.AudioOut(audio)
		audioBuffer.Read()
	}

	if audioBuffer.readIndex != 0 {
		t.Errorf("Wrong readIndex after read Audio:\n got: %d\nwant: %d", audioBuffer.readIndex, 0)
	}
}

func TestAudioBuffer_Read_empty(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}

	audio := &Audio{sampleCount: 1024}

	audioBuffer.AudioOut(audio)
	audioBuffer.Read()

	if audioBuffer.Read() != nil {
		t.Errorf("Read() returns Audio while empty")
	}
}

func TestAudioBuffer_Full(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	audio := &Audio{sampleCount: 1024}

	for time := 0; uint32(time) < maxAudioBufferSize; time++ {
		audioBuffer.AudioOut(audio)
	}

	if !audioBuffer.Full() {
		t.Errorf("AudioBuffer should be Full with %d Audios", maxAudioBufferSize)
	}
}
