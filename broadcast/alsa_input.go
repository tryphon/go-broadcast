package broadcast

import (
	alsa "github.com/tryphon/alsa-go"
)

type AlsaInput struct {
	handle            alsa.Handle
	Device            string
	SampleRate        int
	BufferSampleCount int

	audioHandler AudioHandler

	bufferLength int
	buffer       []byte
}

func (input *AlsaInput) Init() (err error) {
	if input.Device == "" {
		input.Device = "default"
	}

	err = input.handle.Open(input.Device, alsa.StreamTypeCapture, alsa.ModeBlock)
	if err != nil {
		return err
	}

	if input.SampleRate == 0 {
		input.SampleRate = 44100
	}

	input.handle.SampleFormat = alsa.SampleFormatS16LE
	input.handle.SampleRate = input.SampleRate
	input.handle.Channels = 2

	err = input.handle.ApplyHwParams()
	if err != nil {
		return err
	}

	if input.BufferSampleCount == 0 {
		input.BufferSampleCount = 1024
	}

	input.bufferLength = input.BufferSampleCount * input.handle.FrameSize()
	input.buffer = make([]byte, input.bufferLength)

	return nil
}

func (input *AlsaInput) SetAudioHandler(audioHandler AudioHandler) {
	input.audioHandler = audioHandler
}

func (input *AlsaInput) Read() (err error) {
	readBytes, err := input.handle.Read(input.buffer)

	if err != nil {
		Log.Printf("Write error : %v\n", err)
		return err
	}
	if readBytes != input.bufferLength {
		Log.Printf("Did not read whole buffer (Read %v, expected %v)\n", readBytes, input.bufferLength)
	}

	if readBytes > 0 {
		audio := Audio{}
		audio.LoadPcmBytes(input.buffer, readBytes/input.handle.FrameSize(), input.handle.Channels)

		if input.audioHandler != nil {
			input.audioHandler.AudioOut(&audio)
		}
	}

	return nil
}

func (input *AlsaInput) Run() {
	for {
		input.Read()
	}
}
