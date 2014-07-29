package broadcast
import (
	"bytes"
)

type InterleavedAudioCoder struct {
	SampleFormat SampleFormat
	ChannelCount int
}

func (encoder *InterleavedAudioCoder) Encode(audio *Audio) ([]byte, error) {
	buffer := &bytes.Buffer{}
	channelCount := encoder.ChannelCount
	if channelCount == 0 {
		channelCount = audio.ChannelCount()
	}

	for samplePosition := 0; samplePosition < audio.sampleCount; samplePosition++ {
		for channel := 0; channel < channelCount; channel++ {
			sample := audio.Sample(channel, samplePosition)
			encoder.SampleFormat.Write(buffer, sample)
		}
	}

	return buffer.Bytes(), nil
}

func (encoder *InterleavedAudioCoder) FrameSize() int {
	return encoder.SampleFormat.SampleSize() * encoder.ChannelCount
}

func (encoder *InterleavedAudioCoder) Decode(data []byte) (*Audio, error) {
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
