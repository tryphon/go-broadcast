package broadcast

import (
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

type AudioProvider interface {
	Read() *Audio
}

type AudioProviderFunc func() *Audio

func (f AudioProviderFunc) Read() *Audio {
	return f()
}

type Audio struct {
	samples      [][]float32
	channelCount int
	sampleCount  int

	timestamp time.Time
}

func NewAudio(sampleCount int, channelCount int) *Audio {
	audio := &Audio{
		sampleCount:  sampleCount,
		channelCount: channelCount,
		timestamp:    time.Now().UTC(),
		samples:      make([][]float32, channelCount),
	}

	for channel := 0; channel < channelCount; channel++ {
		audio.samples[channel] = make([]float32, sampleCount)
	}

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

type AudioProcessor func(int, int, float32) float32

func (audio *Audio) Process(processor AudioProcessor) {
	for channel := 0; channel < audio.channelCount; channel++ {
		for samplePosition := 0; samplePosition < audio.sampleCount; samplePosition++ {
			newSample := processor(channel, samplePosition, audio.Sample(channel, samplePosition))
			audio.SetSample(channel, samplePosition, newSample)
		}
	}
}

func (audio *Audio) SetSample(channel int, samplePosition int, sample float32) {
	audio.samples[channel][samplePosition] = sample
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
