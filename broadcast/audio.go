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
	integerValue := int16(sample * 32768)
	return byte(integerValue), byte(integerValue >> 8)
}

func (audio *Audio) PcmBytes() []byte {
	pcmSampleSize := 4
	pcmBytesLength := audio.sampleCount * pcmSampleSize
	pcmBytes := make([]byte, pcmBytesLength)

	for samplePosition := 0; samplePosition < audio.sampleCount; samplePosition++ {
		pcmBytes[samplePosition*pcmSampleSize], pcmBytes[samplePosition*pcmSampleSize+1] = floatSamplesToBytes(audio.samples[0][samplePosition])
		pcmBytes[samplePosition*pcmSampleSize+2], pcmBytes[samplePosition*pcmSampleSize+3] = floatSamplesToBytes(audio.samples[1][samplePosition])
	}

	return pcmBytes
}
