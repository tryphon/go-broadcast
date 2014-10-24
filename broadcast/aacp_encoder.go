package broadcast

/*
#cgo LDFLAGS: -laacplus
#include "aacplus.h"
*/
import "C"

import (
	"errors"
	"io"
	"runtime"
	"unsafe"
)

type AACPEncoder struct {
	BitRate      int
	ChannelCount int
	SampleRate   int

	Writer io.Writer

	handle       C.aacplusEncHandle
	encodedBytes []byte
	resizeAudio  *ResizeAudio
}

func (encoder *AACPEncoder) Init() error {
	if encoder.BitRate == 0 {
		encoder.BitRate = 48000
	}
	if encoder.ChannelCount == 0 {
		encoder.ChannelCount = 2
	}
	if encoder.SampleRate == 0 {
		encoder.SampleRate = 44100
	}

	var inputSamples C.ulong
	var maxOutputBytes C.ulong

	handle := C.aacplusEncOpen(
		C.ulong(encoder.SampleRate),
		C.uint(encoder.ChannelCount),
		&inputSamples,
		&maxOutputBytes)

	config := C.aacplusEncGetCurrentConfiguration(handle)
	config.bitRate = C.int(encoder.BitRate)
	config.bandWidth = C.int(0)
	config.outputFormat = C.int(1)
	config.nChannelsOut = C.int(encoder.ChannelCount)
	config.inputFormat = C.AACPLUS_INPUT_FLOAT

	if C.aacplusEncSetConfiguration(handle, config) == 0 {
		return errors.New("Can't configure AAC+ encoder : ")
	}

	encoder.encodedBytes = make([]byte, maxOutputBytes)
	encoder.resizeAudio = &ResizeAudio{
		SampleCount: int(inputSamples) / encoder.ChannelCount,
		Output:      AudioHandlerFunc(encoder.audioOut),
	}

	encoder.handle = handle
	runtime.SetFinalizer(encoder, finalizeAACPEncoder)

	return nil
}

func (encoder *AACPEncoder) AudioOut(audio *Audio) {
	if encoder.resizeAudio != nil {
		encoder.resizeAudio.AudioOut(audio)
	}
}

func (encoder *AACPEncoder) audioOut(audio *Audio) {
	interleavedFloats := audio.InterleavedFloats()
	encodedByteCount := C.aacplusEncEncode(encoder.handle,
		(*C.int32_t)(unsafe.Pointer(&interleavedFloats[0])),
		C.uint(len(interleavedFloats)),
		(*C.uchar)(unsafe.Pointer(&encoder.encodedBytes[0])),
		C.uint(len(encoder.encodedBytes)))

	if encodedByteCount > 0 {
		encoder.Writer.Write(encoder.encodedBytes[0:encodedByteCount])
	}
}

func (encoder *AACPEncoder) Close() {
	if encoder.handle != nil {
		C.aacplusEncClose(encoder.handle)
		encoder.handle = nil
	}
}

func finalizeAACPEncoder(encoder *AACPEncoder) {
	encoder.Close()
}
