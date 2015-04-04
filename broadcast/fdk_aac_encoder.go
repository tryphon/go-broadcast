package broadcast

/*
#include <fdk-aac/aacenc_lib.h>
#include <sys/types.h>

AACENC_ERROR aacEncEncode_s16le(HANDLE_AACENCODER handle, int16_t* convert_buf, int input_size, u_int8_t* outbuf, int out_size, int* numOutBytes) {
  AACENC_BufDesc in_buf = { 0 }, out_buf = { 0 };
  AACENC_InArgs in_args = { 0 };
  AACENC_OutArgs out_args = { 0 };
  int in_identifier = IN_AUDIO_DATA;
  int in_size, in_elem_size;
  int out_identifier = OUT_BITSTREAM_DATA;
  int out_elem_size;
  void *in_ptr, *out_ptr;
  AACENC_ERROR err;

  if (input_size <= 0) {
    in_args.numInSamples = -1;
  } else {
    in_ptr = convert_buf;
    in_size = input_size;
    in_elem_size = 2;

    in_args.numInSamples = input_size/2;
    in_buf.numBufs = 1;
    in_buf.bufs = &in_ptr;
    in_buf.bufferIdentifiers = &in_identifier;
    in_buf.bufSizes = &in_size;
    in_buf.bufElSizes = &in_elem_size;
  }

  out_ptr = outbuf;

  out_elem_size = 1;
  out_buf.numBufs = 1;
  out_buf.bufs = &out_ptr;
  out_buf.bufferIdentifiers = &out_identifier;
  out_buf.bufSizes = &out_size;
  out_buf.bufElSizes = &out_elem_size;

  if ((err = aacEncEncode(handle, &in_buf, &out_buf, &in_args, &out_args)) != AACENC_OK) {
    return err;
  };

  *numOutBytes = out_args.numOutBytes;
  return err;
}

#cgo LDFLAGS: -lfdk-aac
*/
import "C"

import (
	"errors"
	"io"
	"unsafe"
)

type FDKAACEncoder struct {
	ChannelCount int
	AOT          int
	SampleRate   int
	Mode         string
	Quality      float32
	BitRate      int
	Writer       io.Writer

	handle      C.HANDLE_AACENCODER
	frameLength int
	resizeAudio *ResizeAudio
	coder       *InterleavedAudioCoder
}

func (encoder *FDKAACEncoder) Init() error {
	if encoder.ChannelCount == 0 {
		encoder.ChannelCount = 2
	}
	if encoder.AOT == 0 {
		encoder.AOT = 2
	}
	if encoder.SampleRate == 0 {
		encoder.SampleRate = 44100
	}
	if encoder.Mode == "" {
		encoder.Mode = "cbr"
	}
	if encoder.BitRate == 0 {
		encoder.BitRate = 96000
	}

	if C.aacEncOpen(&encoder.handle, 0, C.UINT(encoder.ChannelCount)) != C.AACENC_OK {
		return errors.New("Can't open encoder")
	}

	// TODO AACENC_SBR_MODE

	if C.aacEncoder_SetParam(encoder.handle, C.AACENC_AOT, C.UINT(encoder.AOT)) != C.AACENC_OK {
		return errors.New("Unable to set the AOT")
	}

	if C.aacEncoder_SetParam(encoder.handle, C.AACENC_SAMPLERATE, C.UINT(encoder.SampleRate)) != C.AACENC_OK {
		return errors.New("Unable to set the samplerate")
	}

	if C.aacEncoder_SetParam(encoder.handle, C.AACENC_CHANNELMODE, C.UINT(encoder.ChannelMode())) != C.AACENC_OK {
		return errors.New("Unable to set the channel mode")
	}

	if C.aacEncoder_SetParam(encoder.handle, C.AACENC_CHANNELORDER, C.UINT(0)) != C.AACENC_OK {
		return errors.New("Unable to set the channel order")
	}

	if encoder.Mode == "vbr" {
		if C.aacEncoder_SetParam(encoder.handle, C.AACENC_BITRATEMODE, encoder.BitRateMode()) != C.AACENC_OK {
			return errors.New("Unable to set the VBR bitrate mode")
		}
	} else {
		if C.aacEncoder_SetParam(encoder.handle, C.AACENC_BITRATE, C.UINT(encoder.BitRate)) != C.AACENC_OK {
			return errors.New("Unable to set the bitrate")
		}
	}

	if C.aacEncoder_SetParam(encoder.handle, C.AACENC_TRANSMUX, 2) != C.AACENC_OK {
		return errors.New("Unable to set the ADTS transmux")
	}

	if C.aacEncoder_SetParam(encoder.handle, C.AACENC_AFTERBURNER, 1) != C.AACENC_OK {
		return errors.New("Unable to set the afterburner mode")
	}

	if C.aacEncEncode(encoder.handle, nil, nil, nil, nil) != C.AACENC_OK {
		return errors.New("Unable to initialize the encoder")
	}

	var info C.AACENC_InfoStruct
	if C.aacEncInfo(encoder.handle, &info) != C.AACENC_OK {
		return errors.New("Unable to retrieve encoder info")
	}

	encoder.resizeAudio = &ResizeAudio{
		SampleCount: int(info.frameLength),
		Output:      AudioHandlerFunc(encoder.audioOut),
	}

	encoder.coder = &InterleavedAudioCoder{
		SampleFormat: Sample16bLittleEndian,
		ChannelCount: encoder.ChannelCount,
	}

	return nil
}

func (encoder *FDKAACEncoder) BitRateMode() C.UINT {
	if encoder.Mode == "vbr" {
		return C.UINT(encoder.Quality * 7)
	} else {
		return 0
	}
}

func (encoder *FDKAACEncoder) ChannelMode() (channelMode C.CHANNEL_MODE) {
	switch encoder.ChannelCount {
	case 1:
		channelMode = C.MODE_1
	case 2:
		channelMode = C.MODE_2
	case 3:
		channelMode = C.MODE_1_2
	case 4:
		channelMode = C.MODE_1_2_1
	case 5:
		channelMode = C.MODE_1_2_2
	case 6:
		channelMode = C.MODE_1_2_2_1
	case 8:
		channelMode = C.MODE_7_1_REAR_SURROUND
	default:
		channelMode = C.MODE_INVALID
	}
	return
}

func (encoder *FDKAACEncoder) AudioOut(audio *Audio) {
	if encoder.resizeAudio != nil {
		encoder.resizeAudio.AudioOut(audio)
	}
}

func (encoder *FDKAACEncoder) audioOut(audio *Audio) {
	pcmBytes, err := encoder.coder.Encode(audio)
	if err != nil {
		Log.Debugf("Can't encode audio in bytes: %v", err.Error())
		return
	}

	encodedBytes := make([]byte, 20480)
	var encodedByteCount C.int

	encode_error := C.aacEncEncode_s16le(
		encoder.handle,
		(*C.int16_t)(unsafe.Pointer(&pcmBytes[0])),
		C.int(len(pcmBytes)),
		(*C.u_int8_t)(unsafe.Pointer(&encodedBytes[0])),
		C.int(len(encodedBytes)),
		&encodedByteCount,
	)

	if encode_error != C.AACENC_OK {
		if encode_error != C.AACENC_ENCODE_EOF {
			Log.Debugf("Encode returns error: %d", encode_error)
			return
		}
	} else {
		if encodedByteCount > 0 && encoder.Writer != nil {
			encoder.Writer.Write(encodedBytes[0:encodedByteCount])
		}
	}
}

func (encoder *FDKAACEncoder) Close() {
	C.aacEncClose(&encoder.handle)
}
