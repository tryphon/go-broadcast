package broadcast

import (
	"github.com/grd/vorbis"
)

type AudioHandler interface {
	AudioOut(audio *Audio)
}

type AudioHandlerFunc func(audio *Audio)

func (f AudioHandlerFunc) AudioOut(audio *Audio) {
	f(audio)
}

type Audio struct {
	samples      [][]float32
	channelCount int
	sampleCount  int
}

func (audio *Audio) SampleCount() int {
	return audio.sampleCount
}

func (audio *Audio) LoadPcmArray(pcmArray ***float32, sampleCount int, channelCount int) {
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

func floatSamplesToBytes(sample float32) (byte, byte) {
	integerValue := int16(sample * 32767)
	return byte(integerValue), byte(integerValue >> 8)
}

func (audio *Audio) PcmBytes() []byte {
	pcmSampleSize := 2
	pcmSampleSetSize := audio.channelCount * pcmSampleSize
	pcmBytesLength := audio.sampleCount * pcmSampleSetSize
	pcmBytes := make([]byte, pcmBytesLength)

	if audio.samples != nil {
		for samplePosition := 0; samplePosition < audio.sampleCount; samplePosition++ {
			if audio.samples[0] != nil {
				pcmPosition := samplePosition * pcmSampleSetSize
				pcmBytes[pcmPosition], pcmBytes[pcmPosition+1] = floatSamplesToBytes(audio.samples[0][samplePosition])
			}
			if audio.samples[1] != nil {
				pcmPosition := (samplePosition * pcmSampleSetSize) + pcmSampleSize
				pcmBytes[pcmPosition], pcmBytes[pcmPosition+1] = floatSamplesToBytes(audio.samples[1][samplePosition])
			}
		}
	}

	return pcmBytes
}
