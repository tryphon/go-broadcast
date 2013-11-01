package broadcast

import (
	"bytes"
	"encoding/binary"
)

type RawAudioEncoder struct {
}

func (encoder *RawAudioEncoder) Encode(audio *Audio) ([]byte, error) {
	buffer := &bytes.Buffer{}
	binary.Write(buffer, binary.LittleEndian, int16(audio.SampleCount()))

	for channel := 0; channel < audio.ChannelCount(); channel++ {
		for samplePosition := 0; samplePosition < audio.SampleCount(); samplePosition++ {
			binary.Write(buffer, binary.LittleEndian, audio.Samples(channel)[samplePosition])
		}
	}

	return buffer.Bytes(), nil
}

type RawAudioDecoder struct {
}

func (decoder *RawAudioDecoder) Decode(data []byte) (*Audio, error) {
	buffer := bytes.NewBuffer(data)

	var sampleCount int16
	binary.Read(buffer, binary.LittleEndian, &sampleCount)

	channelCount := 2

	audio := NewAudio(int(sampleCount), channelCount)
	for channel := 0; channel < channelCount; channel++ {
		audio.SetSamples(channel, make([]float32, sampleCount))

		for samplePosition := 0; samplePosition < audio.SampleCount(); samplePosition++ {
			var sample float32
			binary.Read(buffer, binary.LittleEndian, &sample)
			audio.Samples(channel)[samplePosition] = sample
		}
	}

	return audio, nil
}
