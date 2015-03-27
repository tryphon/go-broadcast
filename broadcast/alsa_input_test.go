package broadcast

import (
	"flag"
	"strings"
	"testing"
	"time"
)

func TestAlsaInputConfig_Flags(t *testing.T) {
	config := AlsaInputConfig{}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	config.Flags(flags, "alsa")

	flags.Parse(strings.Split("-alsa-device=test -alsa-sample-format=s32le -alsa-buffer-duration=100ms -alsa-sample-rate=48000 -alsa-remix=8:9", " "))

	if config.Device != "test" {
		t.Errorf("Device should be 'device' flag value :\n got: %v\nwant: %v", config.Device, "test")
	}
	if config.SampleRate != 48000 {
		t.Errorf("SampleRate should be 'sample-rate' flag value :\n got: %v\nwant: %v", config.SampleRate, 48000)
	}
	if config.SampleFormat != "s32le" {
		t.Errorf("SampleFormat should be 'sample-format' flag value :\n got: %v\nwant: %v", config.SampleFormat, "s32le")
	}
	if config.BufferDuration != 100*time.Millisecond {
		t.Errorf("BufferDuration should be 'buffer-duration' flag value :\n got: %v\nwant: %v", config.BufferDuration, 100*time.Millisecond)
	}
	if config.Remix != "8:9" {
		t.Errorf("Remix should be 'remix' flag value :\n got: %v\nwant: %v", config.Remix, "8:9")
	}
}

func TestAlsaInputConfig_Apply(t *testing.T) {
	config := AlsaInputConfig{Device: "test", SampleRate: 48000, BufferDuration: 100 * time.Millisecond, SampleFormat: "s32le", Remix: "8:9"}
	alsaInput := &AlsaInput{}

	config.Apply(alsaInput)

	if alsaInput.Device != config.Device {
		t.Errorf("AlsaInput Device should be config Device :\n got: %v\nwant: %v", alsaInput.Device, config.Device)
	}
	if alsaInput.SampleRate != config.SampleRate {
		t.Errorf("AlsaInput SampleRate should be config SampleRate :\n got: %v\nwant: %v", alsaInput.SampleRate, config.SampleRate)
	}
	if alsaInput.SampleFormat != Sample32bLittleEndian {
		t.Errorf("AlsaInput SampleFormat should be parsed config SampleFormat :\n got: %v\nwant: %v", alsaInput.SampleFormat, Sample32bLittleEndian)
	}
	if alsaInput.BufferSampleCount != 4800 {
		t.Errorf("AlsaInput BufferSampleCount should be based on config SampleRate and BufferDuration :\n got: %v\nwant: %v", alsaInput.BufferSampleCount, 4800)
	}
	if alsaInput.remixer == nil {
		t.Fatalf("AlsaInput Remixer should be created if config Remix is defined")
	}
	if alsaInput.remixer.String() != "8:9" {
		t.Errorf("AlsaInput Remixer should be based on config Remix :\n got: %v\nwant: %v", alsaInput.remixer.String(), "8:9")
	}
}
