package broadcast

import (
	"errors"
	"flag"
	"strings"
)

type OpusAudioEncoder struct {
	Bitrate     int
	opusEncoder *OpusEncoder
}

func (encoder *OpusAudioEncoder) Init() error {
	if encoder.Bitrate == 0 {
		encoder.Bitrate = 256000
	}

	opusEncoder, err := OpusEncoderCreate(encoder.Bitrate)
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

	encodedLength, err := encoder.opusEncoder.EncodeFloat(audio.InterleavedFloats(), audio.SampleCount(), opusBytes, int32(len(opusBytes)))
	if err != nil {
		return nil, err
	}
	return opusBytes[:encodedLength], nil
}

type OpusAudioDecoder struct {
	opusDecoder *OpusDecoder
}

func (decoder *OpusAudioDecoder) Init() error {
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
	samples := make([]float32, frameCount*2)

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

type OpusAudioEncoderConfig struct {
	Bitrate int
}

func (config *OpusAudioEncoderConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.IntVar(&config.Bitrate, strings.Join([]string{prefix, "bitrate"}, "-"), 256000, "The Opus stream bitrate")
}

func (config *OpusAudioEncoderConfig) Apply(opusEncoder *OpusAudioEncoder) {
	opusEncoder.Bitrate = config.Bitrate
}
