package broadcast

import (
	"testing"
)

func TestUnfillAudioBuffer_NewUnfillAudioBuffer(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	unfillAudioBuffer := &UnfillAudioBuffer{Buffer: audioBuffer}

	if unfillAudioBuffer.Buffer != audioBuffer {
		t.Errorf("Should use given AudioBuffer")
	}

	if unfillAudioBuffer.full() {
		t.Errorf("Should not be full")
	}
}

func TestUnfillAudioBuffer_Read(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	unfillAudioBuffer := &UnfillAudioBuffer{Buffer: audioBuffer}
	audio := NewAudio()

	audioBuffer.AudioOut(audio)

	if unfillAudioBuffer.Read() != audio {
		t.Errorf("Should read Audio in buffer")
	}
}

func TestUnfillAudioBuffer_Write(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	unfillAudioBuffer := &UnfillAudioBuffer{Buffer: audioBuffer}
	audio := NewAudio()

	unfillAudioBuffer.AudioOut(audio)

	if audioBuffer.Read() != audio {
		t.Errorf("Should store Audio in buffer")
	}
}

func TestUnfillAudioBuffer_full(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	unfillAudioBuffer := &UnfillAudioBuffer{Buffer: audioBuffer}

	unfillAudioBuffer.MaxSampleCount = 0
	if unfillAudioBuffer.full() {
		t.Errorf("Should not be full when MaxSampleCount is zero")
	}

	unfillAudioBuffer.UnfillSampleCount = 0
	if unfillAudioBuffer.full() {
		t.Errorf("Should not be full when UnfillSampleCount is zero")
	}

	audio := NewAudio()
	audioBuffer.AudioOut(audio)
	unfillAudioBuffer.MaxSampleCount = 1
	unfillAudioBuffer.UnfillSampleCount = 1

	if ! unfillAudioBuffer.full() {
		t.Errorf("Should be full when buffer SampleCount() > MaxSampleCount")
	}

}

func TestUnfillAudioBuffer_unfill(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	unfillAudioBuffer := &UnfillAudioBuffer{Buffer: audioBuffer}
	audio := NewAudio()

	for times := 0; times < 5; times++ {
		audioBuffer.AudioOut(audio)
	}

	unfillAudioBuffer.MaxSampleCount = uint32(audio.SampleCount() * 3)
	unfillAudioBuffer.UnfillSampleCount = uint32(audio.SampleCount())

	unfillAudioBuffer.unfill()

	targetSampleCount := unfillAudioBuffer.MaxSampleCount - unfillAudioBuffer.UnfillSampleCount

	if audioBuffer.SampleCount() > targetSampleCount {
		t.Errorf("Should unfill buffer SampleCount() <= MaxSampleCount - UnfillSampleCount (%d >= %d)", audioBuffer.SampleCount(), targetSampleCount)
	}
}

func TestUnfillAudioBuffer_AudioOut(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	unfillAudioBuffer := &UnfillAudioBuffer{Buffer: audioBuffer}
	audio := NewAudio()

	unfillAudioBuffer.AudioOut(audio)

	if audioBuffer.Read() != audio {
		t.Errorf("Should store Audio in buffer")
	}
}

func TestUnfillAudioBuffer_AudioOut_WhenFull(t *testing.T) {
	audioBuffer := &MemoryAudioBuffer{}
	unfillAudioBuffer := &UnfillAudioBuffer{Buffer: audioBuffer}
	audio := NewAudio()

	unfillAudioBuffer.MaxSampleCount = uint32(audio.SampleCount() * 3) - 1
	unfillAudioBuffer.UnfillSampleCount = uint32(audio.SampleCount())

	for times := 0; times < 2; times++ {
		unfillAudioBuffer.AudioOut(audio)
	}

	if unfillAudioBuffer.full() {
		t.Errorf("Should not be full before MaxSampleCount is rechead")
	}

	unfillAudioBuffer.AudioOut(audio)

	if unfillAudioBuffer.full() {
		t.Errorf("Should not be full after unfill")
	}

	targetSampleCount := unfillAudioBuffer.MaxSampleCount - unfillAudioBuffer.UnfillSampleCount

	if audioBuffer.SampleCount() > targetSampleCount {
		t.Errorf("Should unfill buffer SampleCount() <= MaxSampleCount - UnfillSampleCount (%d >= %d)", audioBuffer.SampleCount(), targetSampleCount)
	}
}
