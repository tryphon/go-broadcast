package broadcast

import (
	"time"
)

type FixedRateAudioHandler struct {
	SampleRate uint
	Output     AudioHandler
	Tolerance  float64

	expectedNextTime time.Time
}

func (rated *FixedRateAudioHandler) audioDuration(audio *Audio) time.Duration {
	sampleRatio := int(float64(audio.SampleCount()) / float64(rated.SampleRate) * 1000.0)
	return time.Duration(sampleRatio) * time.Millisecond
}

func (rated *FixedRateAudioHandler) fixedDuration(audio *Audio) time.Duration {
	return time.Duration(rated.audioDuration(audio).Seconds()*1000*(1-rated.Tolerance)) * time.Millisecond
}

func (rated *FixedRateAudioHandler) AudioOut(audio *Audio) {
	if !rated.expectedNextTime.IsZero() {
		requiredDelay := rated.expectedNextTime.Sub(time.Now())
		if requiredDelay > 0 {
			Log.Debugf("Required delay: %v", requiredDelay)
			time.Sleep(requiredDelay)
		}
	}

	rated.expectedNextTime = time.Now().Add(rated.fixedDuration(audio))
	rated.Output.AudioOut(audio)
}
