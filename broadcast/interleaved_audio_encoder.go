package broadcast

import (
	"bytes"
)

type InterleavedAudioEncoder struct {
	SampleFormat SampleFormat
	ChannelCount int
}

func (encoder *InterleavedAudioEncoder) Encode(audio *Audio) ([]byte, error) {
	buffer := &bytes.Buffer{}

	for samplePosition := 0; samplePosition < audio.sampleCount; samplePosition++ {
		for channel := 0; channel < audio.ChannelCount(); channel++ {
			sample := audio.Sample(channel, samplePosition)
			encoder.SampleFormat.Write(buffer, sample)
		}
	}

	return buffer.Bytes(), nil
}

func (encoder *InterleavedAudioEncoder) FrameSize() int {
	return encoder.SampleFormat.SampleSize() * encoder.ChannelCount
}

func (encoder *InterleavedAudioEncoder) Decode(data []byte) (*Audio, error) {
	buffer := bytes.NewBuffer(data)

	sampleCount := len(data) / encoder.FrameSize()
	audio := NewAudio(sampleCount, encoder.ChannelCount)

	for channel := 0; channel < encoder.ChannelCount; channel++ {
		audio.SetSamples(channel, make([]float32, sampleCount))
	}

	for samplePosition := 0; samplePosition < sampleCount; samplePosition++ {
		for channel := 0; channel < encoder.ChannelCount; channel++ {
			sample, err := encoder.SampleFormat.Read(buffer)
			if err != nil {
				return nil, err
			}
			audio.Samples(channel)[samplePosition] = sample
		}
	}

	return audio, nil
}
