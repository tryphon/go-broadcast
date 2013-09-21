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