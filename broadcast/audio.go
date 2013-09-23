package broadcast

import (
	"bytes"
	"encoding/binary"
	"github.com/grd/vorbis"
	"time"
	"io"
	alsa "github.com/tryphon/alsa-go"
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

func (audio *Audio) LoadPcmBytes(pcmBytes []byte, sampleCount int, channelCount int, sampleFormat alsa.SampleFormat) {
	audio.channelCount = channelCount
	audio.sampleCount = sampleCount

	audio.samples = make([][]float32, channelCount)
	for channel := 0; channel < channelCount; channel++ {
		audio.samples[channel] = make([]float32, sampleCount)
	}

	buffer := bytes.NewBuffer(pcmBytes)

	var err error

	switch sampleFormat {
	case alsa.SampleFormatS16LE: err = audio.readPcm16BitsBuffer(buffer, sampleCount, channelCount)
	case alsa.SampleFormatS32LE: err = audio.readPcm32BitsBuffer(buffer, sampleCount, channelCount)
	default: Log.Printf("Unsupported sample format: %s", sampleFormat)
	}

	if err != nil {
		Log.Printf("Can't read correctly pcm buffer: %s", err.Error())
	}
}

func (audio *Audio) readPcm16BitsBuffer(buffer io.Reader, sampleCount int, channelCount int) error {
	for samplePosition := 0; samplePosition < sampleCount; samplePosition++ {
		for channel := 0; channel < channelCount; channel++ {
			var pcmSample int16
			err := binary.Read(buffer, binary.LittleEndian, &pcmSample)
			if err != nil {
				return err
			}
			audio.samples[channel][samplePosition] = pcmSample16BitsToFloat(pcmSample)
		}
	}
	return nil
}

func pcmSample16BitsToFloat(pcmSample int16) float32 {
	if pcmSample != -32768 {
		return float32(pcmSample) / 32767
	} else {
		return -1
	}
}

func (audio *Audio) readPcm32BitsBuffer(buffer io.Reader, sampleCount int, channelCount int) error {
	for samplePosition := 0; samplePosition < sampleCount; samplePosition++ {
		for channel := 0; channel < channelCount; channel++ {
			var pcmSample int32
			err := binary.Read(buffer, binary.LittleEndian, &pcmSample)
			if err != nil {
				return err
			}
			audio.samples[channel][samplePosition] = pcmSample32BitsToFloat(pcmSample)
		}
	}
	return nil
}

func pcmSample32BitsToFloat(pcmSample int32) float32 {
	if pcmSample != -2147483648 {
		return float32(pcmSample) / 2147483648
	} else {
		return -1
	}
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
