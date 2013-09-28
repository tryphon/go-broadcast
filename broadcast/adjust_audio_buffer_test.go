package broadcast

import (
	"testing"
)

func TestAdjustAudioBuffer_adjustmentFactor_lowAdjust(t *testing.T) {
	limitSampleCount := uint32(0 * 44100)
	thresholdSampleCount := uint32(3 * 44100)

	buffer := &AdjustAudioBuffer{
		Buffer:               &MemoryAudioBuffer{},
		LimitSampleCount:     limitSampleCount,
		ThresholdSampleCount: thresholdSampleCount,
	}

	if buffer.adjustmentFactor() != 1 {
		t.Errorf("Wrong adjustmentFactor:\n got: %d\nwant: %d", buffer.adjustmentFactor(), 1)
	}
}

func TestAdjustAudioBuffer_adjustmentFactor_highAdjust(t *testing.T) {
	limitSampleCount := uint32(10 * 44100)
	thresholdSampleCount := uint32(7 * 44100)

	buffer := &AdjustAudioBuffer{
		Buffer:               &MemoryAudioBuffer{},
		LimitSampleCount:     limitSampleCount,
		ThresholdSampleCount: thresholdSampleCount,
	}

	if buffer.adjustmentFactor() != -1 {
		t.Errorf("Wrong adjustmentFactor:\n got: %d\nwant: %d", buffer.adjustmentFactor(), -1)
	}
}

func TestAdjustAudioBuffer_logAdjustment(t *testing.T) {
	buffer := &AdjustAudioBuffer{}

	audio := NewAudio(1024, 2)
	buffer.logAdjustment(audio)

	if buffer.adjustmentSampleCount != int64(audio.SampleCount()) {
		t.Errorf("logAdjustment should increase adjustmentSampleCount:\n got: %d\nwant: %d", buffer.adjustmentSampleCount, audio.SampleCount())
	}
}

func TestAdjustAudioBuffer_logAdjustment_nil(t *testing.T) {
	buffer := &AdjustAudioBuffer{}
	buffer.logAdjustment(nil)

	if buffer.adjustmentSampleCount != 0 {
		t.Errorf("logAdjustment should not increase adjustmentSampleCount when no audio is given:\n got: %d\nwant: %d", buffer.adjustmentSampleCount, 0)
	}
}
