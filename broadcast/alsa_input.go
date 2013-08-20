package broadcast

import (
	"fmt"
	alsa "github.com/tryphon/alsa-go"
	"os"
)

type AlsaInput struct {
	handle alsa.Handle
	Device      string
	SampleRate int

	audioHandler AudioHandler

	bufferLength int
	buffer       []byte
}

func (input *AlsaInput) Init() (err error) {
	if (input.Device == "") {
		input.Device = "default"
	}

	err = input.handle.Open(input.Device, alsa.StreamTypeCapture, alsa.ModeBlock)
	if err != nil {
		return err
	}

	if (input.SampleRate == 0) {
		input.SampleRate = 44100
	}

	input.handle.SampleFormat = alsa.SampleFormatS16LE
	input.handle.SampleRate = input.SampleRate
	input.handle.Channels = 2

	input.bufferLength = 4096
	input.buffer = make([]byte, input.bufferLength)

	err = input.handle.ApplyHwParams()
	return err
}

func (input *AlsaInput) SetAudioHandler(audioHandler AudioHandler) {
	input.audioHandler = audioHandler
}

func (input *AlsaInput) Read() (err error) {
	readBytes, err := input.handle.Read(input.buffer)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Write error : %v\n", err)
		return err
	}
	if readBytes != input.bufferLength {
		fmt.Fprintf(os.Stderr, "Did not read whole buffer (Read %v, expected %v)\n", readBytes, input.bufferLength)
	}

	audio := Audio{}
	audio.LoadPcmBytes(input.buffer, readBytes/input.handle.FrameSize(), input.handle.Channels)

	if input.audioHandler != nil {
		input.audioHandler.AudioOut(&audio)
	}

	return nil
}

func (input *AlsaInput) Run() {
	for {
		input.Read()
	}
}
