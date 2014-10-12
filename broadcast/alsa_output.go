package broadcast

import (
	"flag"
	metrics "github.com/tryphon/go-metrics"
	alsa "github.com/tryphon/alsa-go"
	"strings"
	"time"
)

type AlsaOutput struct {
	handle       alsa.Handle
	Device       string
	SampleRate   int
	SampleFormat SampleFormat
	Channels     int

	coder *InterleavedAudioCoder
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
	if output.Channels == 0 {
		output.Channels = 2
	}

	if output.SampleFormat != nil {
		output.handle.SampleFormat = ToAlsaSampleFormat(output.SampleFormat)
	}
	output.handle.SampleRate = output.SampleRate
	output.handle.Channels = output.Channels

	err = output.handle.ApplyHwParams()
	if err != nil {
		return err
	}

	output.SampleFormat = FromAlsaSampleFormat(output.handle.SampleFormat)
	output.coder = &InterleavedAudioCoder{SampleFormat: output.SampleFormat, ChannelCount: output.handle.Channels}

	Log.Debugf("Alsa SampleFormat: %v", output.SampleFormat.Name())
	Log.Debugf("Alsa SampleRate: %v", output.handle.SampleRate)

	return nil
}

func (alsa *AlsaOutput) AudioOut(audio *Audio) {
	delay := alsa.Delay()
	metrics.GetOrRegisterGauge("alsa.output.Delay", nil).Update(int64(delay))
	if delay < 0 {
		Log.Printf("Alsa delay is negative (%d), waiting for better conditions", delay)
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

	metrics.GetOrRegisterCounter("alsa.output.SampleCount", nil).Inc(wroteSamples)

	if alsaWriteLength != len(pcmBytes) {
		Log.Debugf("Did not write whole alsa buffer (Wrote %v, expected %v)", alsaWriteLength, len(pcmBytes))
	}
}

func (alsa *AlsaOutput) Delay() (delay int) {
	delay, _ = alsa.handle.Delay()
	return delay
}

type AlsaOutputConfig struct {
	Device       string
	SampleRate   int
	SampleFormat string
}

func (config *AlsaOutputConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.Device, strings.Join([]string{prefix, "device"}, "-"), "default", "The alsa device used to record sound")
	flags.IntVar(&config.SampleRate, strings.Join([]string{prefix, "sample-rate"}, "-"), 44100, "Sample rate")
	flags.StringVar(&config.SampleFormat, strings.Join([]string{prefix, "sample-format"}, "-"), "auto", "The sample format used to record sound (s16le, s32le, s32be)")
}

func (config *AlsaOutputConfig) Apply(alsaOutput *AlsaOutput) {
	alsaOutput.Device = config.Device
	alsaOutput.SampleRate = config.SampleRate
	alsaOutput.SampleFormat = ParseSampleFormat(config.SampleFormat)
}
