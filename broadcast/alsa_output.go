package broadcast

import (
	alsa "github.com/tryphon/alsa-go"
	"time"
)

type AlsaOutput struct {
	handle       alsa.Handle
	Device       string
	SampleRate   int
	SampleFormat SampleFormat

	sampleCount int64
	coder       *InterleavedAudioEncoder
}

func (output *AlsaOutput) Init() error {
	if output.Device == "" {
		output.Device = "default"
	}

	err := output.handle.Open(output.Device, alsa.StreamTypePlayback, alsa.ModeBlock)
	if err != nil {
		return err
	}

	if output.SampleRate == 0 {
		output.SampleRate = 44100
	}

	if output.SampleFormat != nil {
		output.handle.SampleFormat = ToAlsaSampleFormat(output.SampleFormat)
	}
	output.handle.SampleRate = output.SampleRate
	output.handle.Channels = 2

	err = output.handle.ApplyHwParams()
	if err != nil {
		return err
	}

	output.SampleFormat = FromAlsaSampleFormat(output.handle.SampleFormat)
	output.coder = &InterleavedAudioEncoder{SampleFormat: output.SampleFormat, ChannelCount: output.handle.Channels}

	Log.Debugf("Alsa SampleFormat: %v", output.SampleFormat.Name())

	return nil
}

func (alsa *AlsaOutput) AudioOut(audio *Audio) {
	if alsa.Delay() < 0 {
		Log.Printf("Alsa delay is negative (%d), waiting for better conditions", alsa.Delay())
		time.Sleep(time.Second)
	}
	alsa.handle.AvailUpdate()

	pcmBytes, err := alsa.coder.Encode(audio)
	if err != nil {
		Log.Debugf("Can't encode audio in bytes: %v", err.Error())
		return
	}

	alsaWriteLength, err := alsa.handle.Write(pcmBytes)
	if err != nil {
		Log.Debugf("Can't write alsa buffer: %v", err.Error())
		return
	}

	wroteSamples := int64(alsaWriteLength / len(pcmBytes) * audio.sampleCount)
	alsa.sampleCount += wroteSamples

	if alsaWriteLength != len(pcmBytes) {
		Log.Debugf("Did not write whole alsa buffer (Wrote %v, expected %v)", alsaWriteLength, len(pcmBytes))
	}
}

func (alsa *AlsaOutput) SampleCount() int64 {
	return alsa.sampleCount
}

func (alsa *AlsaOutput) Delay() (delay int) {
	delay, _ = alsa.handle.Delay()
	return delay
}
