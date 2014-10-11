package broadcast

/*
#cgo LDFLAGS: -lmp3lame
#include "lame/lame.h"
*/
import "C"

import (
	"errors"
	metrics "github.com/tryphon/go-metrics"
	"io"
	"math"
	"runtime"
	"unsafe"
)

const (
	STEREO        = C.STEREO
	JOINT_STEREO  = C.JOINT_STEREO
	MONO          = C.MONO
	NOT_SET       = C.NOT_SET
	MAX_INDICATOR = C.MAX_INDICATOR
)

type LameEncoder struct {
	Quality      float32
	ChannelCount int
	SampleRate   int

	Writer io.Writer

	handle *C.lame_global_flags
}

func (encoder *LameEncoder) Init() error {
	if encoder.ChannelCount == 0 {
		encoder.ChannelCount = 2
	}
	if encoder.SampleRate == 0 {
		encoder.SampleRate = 44100
	}

	handle := C.lame_init()
	if handle == nil {
		return errors.New("Can't initialize lame")
	}

	C.lame_set_num_channels(handle, C.int(encoder.ChannelCount))
	C.lame_set_in_samplerate(handle, C.int(encoder.SampleRate))

	C.lame_set_quality(handle, C.int(encoder.LameQuality()))
	C.lame_set_mode(handle, (C.MPEG_mode)(encoder.LameMode()))

	initResults := C.lame_init_params(handle)
	if initResults == -1 {
		return errors.New("Can't setup lame")
	}

	encoder.handle = handle
	runtime.SetFinalizer(encoder, finalizeLameEncoder)

	return nil
}

func (encoder *LameEncoder) AudioOut(audio *Audio) {
	if encoder.Writer == nil {
		return
	}

	estimatedSize := int(1.25*float64(audio.SampleCount()) + 7200)

	encodedBytes := make([]byte, estimatedSize)
	encodedByteCount := C.int(C.lame_encode_buffer_ieee_float(
		encoder.handle,
		(*C.float)(unsafe.Pointer(&audio.Samples(0)[0])),
		(*C.float)(unsafe.Pointer(&audio.Samples(1)[0])),
		C.int(audio.SampleCount()),
		(*C.uchar)(unsafe.Pointer(&encodedBytes[0])),
		C.int(estimatedSize),
	))

	metrics.GetOrRegisterGauge("lame.frames", nil).Update(int64(C.long(C.lame_get_frameNum(encoder.handle))))

	encoder.Writer.Write(encodedBytes[0:encodedByteCount])
}

func (encoder *LameEncoder) Flush() {
	estimatedSize := 7200
	encodedBytes := make([]byte, estimatedSize)

	encodedByteCount := C.int(C.lame_encode_flush(
		encoder.handle,
		(*C.uchar)(unsafe.Pointer(&encodedBytes[0])),
		C.int(estimatedSize),
	))

	encoder.Writer.Write(encodedBytes[0:encodedByteCount])
}

func (encoder *LameEncoder) LameQuality() int {
	quality := (1.0 - (float64)(encoder.Quality)) * 10.0
	return (int)(math.Min(quality, 9))
}

func (encoder *LameEncoder) LameMode() int {
	if encoder.ChannelCount == 1 {
		return MONO
	} else {
		return JOINT_STEREO
	}
}

func (encoder *LameEncoder) Close() {
	if encoder.handle != nil {
		C.lame_close(encoder.handle)
		encoder.handle = nil
	}
}

func finalizeLameEncoder(encoder *LameEncoder) {
	encoder.Close()
}
