package broadcast

import (
	"testing"
)

func TestAmplifier_AudioOut(t *testing.T) {
	audio := NewAudio(1024, 2)
	audio.Process(func(_ int, _ int, _ float32) float32 {
		return 1
	})

	amplifier := Amplifier{Amplification: 1}
	// Should not without Output
	amplifier.AudioOut(audio)

	amplifier.Output = AudioHandlerFunc(func(audio *Audio) {
		audio.Process(func(_ int, _ int, sample float32) float32 {
			if sample != 2 {
				t.Errorf("Wrong amplified sample :\n got: %v\nwant: %v", sample, 2)
			}

			return sample
		})
	})
	amplifier.AudioOut(audio)
}

func TestAmplifier_amplify(t *testing.T) {
	var conditions = []struct {
		sample          float32
		amplification   float32
		amplifiedSample float32
	}{
		{1, 0, 1},
		{0, 10000000, 0},
		{1, -1, 0},
		{1, 1, 2},
		{-1, 1000, -1001},
	}

	for _, condition := range conditions {
		amplifier := Amplifier{Amplification: condition.amplification}
		if amplifier.amplify(0, 0, condition.sample) != condition.amplifiedSample {
			t.Errorf("Wrong amplified sampled :\n got: %v\nwant: %v", amplifier.amplify(0, 0, condition.sample), condition.amplifiedSample)
		}
	}
}
