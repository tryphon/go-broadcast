package broadcast

/*
#cgo LDFLAGS: -lopus
#include <opus/opus.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

type OpusEncoder struct {
	handle *C.OpusEncoder
}

const (
	OPUS_APPLICATION_AUDIO int = C.OPUS_APPLICATION_AUDIO

	OPUS_OK int = C.OPUS_OK
)

func OpusEncoderCreate() (*OpusEncoder, error) {
	encoder := &OpusEncoder{}

	var cError C.int
	handle := C.opus_encoder_create(48000, 2, C.int(OPUS_APPLICATION_AUDIO), &cError)

	if int(cError) != OPUS_OK {
		return nil, errors.New(fmt.Sprintf("Can't create Opus encoder: %d", int(cError)))
	}

	encoder.handle = handle
	return encoder, nil
}

func (encoder *OpusEncoder) EncodeFloat(pcmFloats []float32, frameSize int, data []byte, maxDataSize int32) (int32, error) {
	cLength := C.opus_encode_float(encoder.handle, (*C.float)(unsafe.Pointer(&pcmFloats[0])), C.int(frameSize), (*C.uchar)(unsafe.Pointer(&data[0])), C.opus_int32(maxDataSize))
	if cLength > 0 {
		return int32(cLength), nil
	} else {
		return 0, errors.New(fmt.Sprintf("Can't encode: %d", int(cLength)))
	}
}

func (encoder *OpusEncoder) Destroy() {
	C.opus_encoder_destroy(encoder.handle)
}

type OpusDecoder struct {
	handle *C.OpusDecoder
}

func OpusDecoderCreate() (*OpusDecoder, error) {
	decoder := &OpusDecoder{}

	var cError C.int
	handle := C.opus_decoder_create(48000, 2, &cError)

	if int(cError) != OPUS_OK {
		return nil, errors.New(fmt.Sprintf("Can't create Opus decoder: %d", int(cError)))
	}

	decoder.handle = handle
	return decoder, nil
}

func (decoder *OpusDecoder) DecodeFloat(data []byte, pcmFloats []float32, frameSize int) (int32, error) {
	cLength := C.opus_decode_float(decoder.handle, (*C.uchar)(unsafe.Pointer(&data[0])), C.opus_int32(len(data)), (*C.float)(unsafe.Pointer(&pcmFloats[0])), C.int(frameSize), 0)
	if cLength > 0 {
		return int32(cLength), nil
	} else {
		return 0, errors.New(fmt.Sprintf("Can't decode: %d", int(cLength)))
	}
}

func (decoder *OpusDecoder) Destroy() {
	C.opus_decoder_destroy(decoder.handle)
}
