package broadcast

import (
	alsa "github.com/tryphon/alsa-go"
	metrics "github.com/tryphon/go-metrics"
)

type AlsaInput struct {
	handle            alsa.Handle
	Device            string
	SampleRate        int
	BufferSampleCount int
	SampleFormat      SampleFormat

	audioHandler AudioHandler

	bufferLength int
	buffer       []byte

	decoder *InterleavedAudioCoder
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

	if input.SampleFormat != nil {
		input.handle.SampleFormat = ToAlsaSampleFormat(input.SampleFormat)
	}
	input.handle.SampleRate = input.SampleRate
	input.handle.Channels = 2

	err = input.handle.ApplyHwParams()
	if err != nil {
		return err
	}

	input.SampleFormat = FromAlsaSampleFormat(input.handle.SampleFormat)
	input.decoder = &InterleavedAudioCoder{SampleFormat: input.SampleFormat, ChannelCount: input.handle.Channels}

	Log.Debugf("Alsa SampleFormat: %v", input.SampleFormat.Name())
	Log.Debugf("Alsa SampleRate: %v", input.handle.SampleRate)

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
		Log.Printf("Read error : %v\n", err)
		return err
	}
	if readBytes != input.bufferLength {
		Log.Printf("Did not read whole buffer (Read %v, expected %v)\n", readBytes, input.bufferLength)
	}

	if readBytes > 0 && input.audioHandler != nil {
		audio, err := input.decoder.Decode(input.buffer[:readBytes])
		if err != nil {
			return err
		}

		metrics.GetOrRegisterCounter("alsa.input.SampleCount", nil).Inc(int64(audio.SampleCount()))

		input.audioHandler.AudioOut(audio)
	}

	return nil
}

func (input *AlsaInput) Run() {
	for {
		input.Read()
	}
}
