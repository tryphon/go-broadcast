package broadcast

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type AudioEncoder interface {
	Encode(audio *Audio) ([]byte, error)
}

type AudioDecoder interface {
	Decode([]byte) (*Audio, error)
}

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

type OpusAudioEncoder struct {
	opusEncoder *OpusEncoder
}

func (encoder *OpusAudioEncoder) Init() (error) {
	opusEncoder, err := OpusEncoderCreate()
	if err != nil {
		return err
	}
	encoder.opusEncoder = opusEncoder
	return nil
}

func (encoder *OpusAudioEncoder) Destroy() {
	encoder.opusEncoder.Destroy()
}

func (encoder *OpusAudioEncoder) Encode(audio *Audio) ([]byte, error) {
	opusBytes := make([]byte, 2048)

	encodedLength, err := encoder.opusEncoder.EncodeFloat(audio.InterleavedFloats(), audio.SampleCount(), opusBytes, 1280)
	if err != nil {
		return nil, err
	}
	return opusBytes[:encodedLength], nil
}

type OpusAudioDecoder struct {
	opusDecoder *OpusDecoder
}

func (decoder *OpusAudioDecoder) Init() (error) {
	opusDecoder, err := OpusDecoderCreate()
	if err != nil {
		return err
	}
	decoder.opusDecoder = opusDecoder
	return nil
}

func (decoder *OpusAudioDecoder) Destroy() {
	decoder.opusDecoder.Destroy()
}

func (decoder *OpusAudioDecoder) Decode(data []byte) (*Audio, error) {
	frameCount := 960
	samples := make([]float32, frameCount * 2)

	decodedFrameCount, err := decoder.opusDecoder.DecodeFloat(data, samples, frameCount)
	if err != nil {
		return nil, err
	}

	if int(decodedFrameCount) != frameCount {
		return nil, errors.New("Can't decode expected frameCount")
	}

	audio := NewAudio(frameCount, 2)
	audio.LoadInterleavedFloats(samples, frameCount, 2)
	return audio, nil
}
