package broadcast

import (
	"testing"
)

func TestRefillAudioBuffer_NewRefillAudioBuffer(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	refillAudioBuffer := &RefillAudioBuffer{Buffer: audioBuffer}

	if refillAudioBuffer.Buffer != audioBuffer {
		t.Errorf("Should use given AudioBuffer")
	}

	if refillAudioBuffer.readable {
		t.Errorf("Should set readable to false")
	}
}

func TestRefillAudioBuffer_Read(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	refillAudioBuffer := &RefillAudioBuffer{Buffer: audioBuffer}
	audio := NewAudio()

	refillAudioBuffer.readable = true
	refillAudioBuffer.AudioOut(audio)

	if refillAudioBuffer.Read() != audio {
		t.Errorf("Should read Audio in buffer")
	}
}

func TestRefillAudioBuffer_Read_WhileRefilling(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	refillAudioBuffer := &RefillAudioBuffer{Buffer: audioBuffer}
	refillAudioBuffer.readable = false

	if refillAudioBuffer.Read() != nil {
		t.Errorf("Should return nil when refilling")
	}
}

func TestRefillAudioBuffer_Read_WhenAudioBufferEmpty(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	refillAudioBuffer := &RefillAudioBuffer{Buffer: audioBuffer}

	if refillAudioBuffer.Read() != nil {
		t.Errorf("Should return nil when buffer is empty")
	}

	if refillAudioBuffer.readable {
		t.Errorf("Should set refill flag")
	}
}

func TestRefillAudioBuffer_Write(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	refillAudioBuffer := &RefillAudioBuffer{Buffer: audioBuffer}
	audio := NewAudio()

	refillAudioBuffer.AudioOut(audio)

	if audioBuffer.Read() != audio {
		t.Errorf("Should store Audio in buffer")
	}
}

func TestRefillAudioBuffer_Write_EndsRefill(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	refillAudioBuffer := &RefillAudioBuffer{Buffer: audioBuffer}
	audio := NewAudio()

	refillAudioBuffer.MinSampleCount = uint32(audio.SampleCount() - 1)
	refillAudioBuffer.readable = false

	refillAudioBuffer.AudioOut(audio)

	if ! refillAudioBuffer.readable {
		t.Errorf("Should reset readable when SampleCount > MinSampleCount")
	}
}
