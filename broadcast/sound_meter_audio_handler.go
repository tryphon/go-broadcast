package broadcast

import (
	"fmt"
	"math"
)

type SoundMeterAudioHandler struct {
	Output AudioHandler
}

func (soundMeter *SoundMeterAudioHandler) AudioOut(audio *Audio) {
	var peak float64 = 0
	for channel := 0; channel < audio.ChannelCount(); channel++ {
		for _, sample := range audio.Samples(channel) {
			value := math.Abs(float64(sample))
			if value > peak {
				peak = value
			}
		}
	}

	fmt.Printf("Peak: %02.2f\n", 20*math.Log10(peak))

	soundMeter.Output.AudioOut(audio)
}
