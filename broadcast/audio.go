package broadcast

import (
	vorbis "github.com/tryphon/go-vorbis"
	"time"
)

type AudioHandler interface {
	AudioOut(audio *Audio)
}

// audioHandler := broadcast.AudioHandlerFunc(func(audio *broadcast.Audio) {
//   ...
// })
type AudioHandlerFunc func(audio *Audio)

func (f AudioHandlerFunc) AudioOut(audio *Audio) {
	f(audio)
}

type Audio struct {
	samples      [][]float32
	channelCount int
	sampleCount  int

	timestamp time.Time
}

func NewAudio(sampleCount int, channelCount int) *Audio {
	audio := &Audio{sampleCount: sampleCount, channelCount: channelCount}
	audio.samples = make([][]float32, channelCount)
	audio.timestamp = time.Now().UTC()
	return audio
}

func (audio *Audio) Samples(channel int) []float32 {
	return audio.samples[channel]
}

func (audio *Audio) Sample(channel int, samplePosition int) float32 {
	if audio.samples[channel] != nil {
		return audio.samples[channel][samplePosition]
	} else {
		return 0
	}
}

func (audio *Audio) SetSamples(channel int, samples []float32) {
	audio.samples[channel] = samples
}

func (audio *Audio) SampleCount() int {
	return audio.sampleCount
}

func (audio *Audio) ChannelCount() int {
	return audio.channelCount
}

func (audio *Audio) Timestamp() time.Time {
	return audio.timestamp
}

func (audio *Audio) SetTimestamp(timestamp time.Time) {
	audio.timestamp = timestamp
}

func (audio *Audio) LoadPcmFloats(pcmArray ***float32, sampleCount int, channelCount int) {
	audio.samples = make([][]float32, 2)
	audio.channelCount = channelCount
	audio.sampleCount = sampleCount

	// OPTIMISE see vorbis.AnalysisBuffer
	for channel := 0; channel < channelCount; channel++ {
		audio.samples[channel] = make([]float32, sampleCount)
		for samplePosition := 0; samplePosition < sampleCount; samplePosition++ {
			audio.samples[channel][samplePosition] = vorbis.PcmArrayHelper(*pcmArray, channel, samplePosition)
		}
	}
}

func (audio *Audio) InterleavedFloats() []float32 {
	floatCount := audio.channelCount * audio.sampleCount
	floats := make([]float32, floatCount)

	floatPosition := 0

	for samplePosition := 0; samplePosition < audio.sampleCount; samplePosition++ {
		for channel := 0; channel < audio.channelCount; channel++ {
			if audio.samples[channel] != nil {
				floats[floatPosition] = audio.samples[channel][samplePosition]
			}
			floatPosition += 1
		}
	}

	return floats
}

func (audio *Audio) LoadInterleavedFloats(samples []float32, sampleCount int, channelCount int) {
	audio.samples = make([][]float32, channelCount)
	audio.channelCount = channelCount
	audio.sampleCount = sampleCount

	for channel := 0; channel < channelCount; channel++ {
		audio.samples[channel] = make([]float32, sampleCount)
	}

	floatPosition := 0

	for samplePosition := 0; samplePosition < sampleCount; samplePosition++ {
		for channel := 0; channel < channelCount; channel++ {
			audio.samples[channel][samplePosition] = samples[floatPosition]
			floatPosition += 1
		}
	}
}
