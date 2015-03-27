package broadcast

import (
	"flag"
	alsa "github.com/tryphon/alsa-go"
	metrics "github.com/tryphon/go-metrics"
	"strings"
	"time"
)

type AlsaInput struct {
	handle            alsa.Handle
	Device            string
	SampleRate        int
	BufferSampleCount int
	SampleFormat      SampleFormat
	Channels          int
	remixer           *Remixer

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
	if input.Channels == 0 {
		input.Channels = 2
	}

	if input.SampleFormat != nil {
		input.handle.SampleFormat = ToAlsaSampleFormat(input.SampleFormat)
	}
	input.handle.SampleRate = input.SampleRate
	input.handle.Channels = input.Channels

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

	if input.remixer != nil {
		Log.Debugf("Alsa remixer: '%s'", input.remixer.String())
		input.remixer.Output = input.audioHandler
		input.audioHandler = input.remixer
	}

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

		input.audioOut(audio)
	}

	return nil
}

func (input *AlsaInput) audioOut(audio *Audio) {
	metrics.GetOrRegisterCounter("alsa.input.Samples", nil).Inc(int64(audio.SampleCount()))
	input.audioHandler.AudioOut(audio)
}

func (input *AlsaInput) ChannelCount() int {
	if input.remixer == nil {
		return input.Channels
	} else {
		return input.remixer.OutputChannelCount()
	}
}

func (input *AlsaInput) Run() {
	for {
		input.Read()
	}
}

type AlsaInputConfig struct {
	Device         string
	SampleRate     int
	BufferDuration time.Duration
	SampleFormat   string
	Channels       int
	Remix          string
}

func (config *AlsaInputConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.Device, strings.Join([]string{prefix, "device"}, "-"), "default", "The alsa device used to record sound")
	flags.IntVar(&config.SampleRate, strings.Join([]string{prefix, "sample-rate"}, "-"), 44100, "Sample rate")
	flags.DurationVar(&config.BufferDuration, strings.Join([]string{prefix, "buffer-duration"}, "-"), 250*time.Millisecond, "The alsa buffer duration")
	flags.StringVar(&config.SampleFormat, strings.Join([]string{prefix, "sample-format"}, "-"), "auto", "The sample format used to record sound (s16le, s32le, s32be)")
	flags.IntVar(&config.Channels, strings.Join([]string{prefix, "channels"}, "-"), 2, "The channels count to be used on alsa device")
	flags.StringVar(&config.Remix, strings.Join([]string{prefix, "remix"}, "-"), "", "The remix applied on input audio channels")
}

func (config *AlsaInputConfig) Apply(alsaInput *AlsaInput) {
	alsaInput.Device = config.Device
	alsaInput.SampleRate = config.SampleRate

	bufferSampleCount := int(float64(config.SampleRate) * config.BufferDuration.Seconds())
	alsaInput.BufferSampleCount = bufferSampleCount
	alsaInput.SampleFormat = ParseSampleFormat(config.SampleFormat)
	alsaInput.Channels = config.Channels

	if config.Remix != "" {
		alsaInput.remixer = NewRemixer(config.Remix)
	}
}
