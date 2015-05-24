package broadcast

/*
#cgo LDFLAGS: -lmad
#include "mad.h"

// no typedef in mad.h
typedef struct mad_stream mad_stream_t;
typedef struct mad_frame mad_frame_t;
typedef struct mad_synth mad_synth_t;
typedef struct mad_pcm mad_pcm_t;

int mad_decoder_pending_buffer_length(struct mad_stream *stream) {
  return stream->bufend - stream->next_frame;
}
*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

type MadDecoder struct {
	audioHandler  AudioHandler
	buffer        bytes.Buffer
	readBuffer    []byte
	pendingBuffer []byte

	stream C.mad_stream_t
	frame  C.mad_frame_t
	synth  C.mad_synth_t
}

type MadError int

const (
	MadErrorBufferLength MadError = C.MAD_ERROR_BUFLEN
)

func (err MadError) IsRecoverable() bool {
	return err&0xff00 != 0
}

func (err MadError) String() string {
	return fmt.Sprintf("MadError: %#x", (int)(err))
}

func (err MadError) Error() error {
	return errors.New(err.String())
}

func (decoder *MadDecoder) SetAudioHandler(audioHandler AudioHandler) {
	decoder.audioHandler = audioHandler
}

func (decoder *MadDecoder) Init() error {
	C.mad_stream_init(&decoder.stream)
	C.mad_frame_init(&decoder.frame)
	C.mad_synth_init(&decoder.synth)

	return nil
}

func (decoder *MadDecoder) Reset() {
	C.mad_frame_finish(&decoder.frame)
	C.mad_stream_finish(&decoder.stream)
}

func (decoder *MadDecoder) Read(reader io.Reader) error {
	if decoder.readBuffer == nil {
		decoder.readBuffer = make([]byte, 1024)
	}

	readCount, err := reader.Read(decoder.readBuffer)
	// Log.Debugf("Read: %d bytes, err: %v", readCount, err)
	if err != nil {
		return err
	}

	if readCount == 0 {
		return nil
	}

	buffer := decoder.readBuffer[0:readCount]
	if decoder.pendingBuffer != nil {
		buffer = append(decoder.pendingBuffer, buffer...)
	}

	// Log.Debugf("Decode %d bytes", len(buffer))

	C.mad_stream_buffer(&decoder.stream, (*C.uchar)(unsafe.Pointer(&buffer[0])), (C.ulong)(len(buffer)))

	for {
		if C.mad_frame_decode(&decoder.frame, &decoder.stream) != 0 {
			if decoder.streamError() == MadErrorBufferLength {
				break
			}

			Log.Debugf("Stream error: %v", decoder.streamError())
			if streamError := decoder.streamError(); !streamError.IsRecoverable() {
				return streamError.Error()
			}
		}

		C.mad_synth_frame(&decoder.synth, &decoder.frame)

		decoder.output()
	}

	decoder.pendingBuffer = decoder.streamPendingBuffer()

	return nil
}

func (decoder *MadDecoder) streamError() MadError {
	return (MadError)(decoder.stream.error)
}

func (decoder *MadDecoder) streamPendingBuffer() []byte {
	pendingLength := C.mad_decoder_pending_buffer_length(&decoder.stream)
	// Log.Debugf("Pending buffer length: %d", pendingLength)

	// C.GoBytes seems to be copy data
	return C.GoBytes(unsafe.Pointer(decoder.stream.next_frame), pendingLength)
}

const MadFixedOne float32 = 1 << 28

func (decoder *MadDecoder) output() {
	if decoder.audioHandler == nil {
		return
	}

	pcm := decoder.synth.pcm

	audio := NewAudio((int)(pcm.length), (int)(pcm.channels))

	audio.Process(func(channel int, samplePosition int, _ float32) float32 {
		madSample := pcm.samples[channel][samplePosition]
		floatSample := (float32)(madSample) / MadFixedOne
		return floatSample
	})

	// Log.Debugf("Output audio: %d", audio.SampleCount())
	decoder.audioHandler.AudioOut(audio)
}
