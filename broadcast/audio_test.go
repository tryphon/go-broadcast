package broadcast

import (
	"testing"
)

func TestAudio_SampleCount(t *testing.T) {
	audio := Audio{sampleCount: 1024}

	if audio.SampleCount() != 1024 {
		t.Errorf("Wrong SampleCount() value:\n got: %d\nwant: %d", audio.SampleCount(), 1024)
	}

}

func TestAudio_InterleavedFloats(t *testing.T) {
	audio := NewAudio(1024, 4)

	// Fill channel 0 with 0, channel 1 with 1, channel 2 with 2, ...
	for channel := 0; channel < audio.ChannelCount(); channel++ {
		samples := make([]float32, audio.SampleCount())
		for samplePosition := 0; samplePosition < audio.SampleCount(); samplePosition++ {
			samples[samplePosition] = float32(channel)
		}
		audio.SetSamples(channel, samples)
	}

	for position, float := range audio.InterleavedFloats() {
		expectedFloat := float32(position % audio.ChannelCount())
		if float != expectedFloat {
			t.Errorf("#sample:%d Wrong float sample value:\n got: %d\nwant: %d", position, float, expectedFloat)
		}
	}
}
