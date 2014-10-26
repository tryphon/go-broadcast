package broadcast

/*
#cgo LDFLAGS: -lfaac
#include "faac.h"
*/
import "C"

import (
	faac "github.com/tryphon/go-faac"
	"io"
)

type AACEncoder struct {
	BitRate      int
	ChannelCount int
	SampleRate   int

	Writer io.Writer

	faacEncoder  *faac.Encoder
	encodedBytes []byte
	resizeAudio  *ResizeAudio
}

func (encoder *AACEncoder) Init() error {
	if encoder.BitRate == 0 {
		encoder.BitRate = 48000
	}
	if encoder.ChannelCount == 0 {
		encoder.ChannelCount = 2
	}
	if encoder.SampleRate == 0 {
		encoder.SampleRate = 44100
	}

	faacEncoder := faac.Open(encoder.SampleRate, encoder.ChannelCount)

	config := faac.EncoderConfiguration{
		// Not available for the moemnt
		// QuantizerQuality: int(encoder.Quality * 1000),

		// faac bitrate is per channel
		BitRate: encoder.BitRate / encoder.ChannelCount,

		InputFormat: faac.InputFloat,
	}
	err := faacEncoder.SetConfiguration(&config)
	if err != nil {
		return err
	}

	encoder.encodedBytes = faacEncoder.OutputBuffer()
	encoder.resizeAudio = &ResizeAudio{
		SampleCount: faacEncoder.InputSamples() / encoder.ChannelCount,
		Output:      AudioHandlerFunc(encoder.audioOut),
	}

	encoder.faacEncoder = faacEncoder

	return nil
}

func (encoder *AACEncoder) AudioOut(audio *Audio) {
	if encoder.resizeAudio != nil {
		encoder.resizeAudio.AudioOut(audio)
	}
}

func (encoder *AACEncoder) audioOut(audio *Audio) {
	interleavedFloats := audio.InterleavedFloats()
	// faac requires values between +/- 32768
	for index, float := range interleavedFloats {
		interleavedFloats[index] = float * 32767.0
	}

	encodedByteCount := encoder.faacEncoder.EncodeFloats(
		interleavedFloats,
		encoder.encodedBytes)

	if encodedByteCount > 0 {
		encoder.Writer.Write(encoder.encodedBytes[0:encodedByteCount])
	}
}

func (encoder *AACEncoder) Close() {
	if encoder.faacEncoder != nil {
		encoder.faacEncoder.Close()
		encoder.faacEncoder = nil
	}
}
