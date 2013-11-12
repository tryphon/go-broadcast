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

	flags.Parse(strings.Split("-alsa-device=test -alsa-sample-format=s32le -alsa-buffer-duration=100ms -alsa-sample-rate=48000", " "))

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
}

func TestAlsaInputConfig_Apply(t *testing.T) {
	config := AlsaInputConfig{Device: "test", SampleRate: 48000, BufferDuration: 100 * time.Millisecond, SampleFormat: "s32le"}
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
}

func TestAlsaOutputConfig_Flags(t *testing.T) {
	config := AlsaOutputConfig{}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	config.Flags(flags, "alsa")

	flags.Parse(strings.Split("-alsa-device=test -alsa-sample-format=s32le -alsa-sample-rate=48000", " "))

	if config.Device != "test" {
		t.Errorf("Device should be 'device' flag value :\n got: %v\nwant: %v", config.Device, "test")
	}
	if config.SampleRate != 48000 {
		t.Errorf("SampleRate should be 'sample-rate' flag value :\n got: %v\nwant: %v", config.SampleRate, 48000)
	}
	if config.SampleFormat != "s32le" {
		t.Errorf("SampleFormat should be 'sample-format' flag value :\n got: %v\nwant: %v", config.SampleFormat, "s32le")
	}
}

func TestAlsaOutputConfig_Apply(t *testing.T) {
	config := AlsaOutputConfig{Device: "test", SampleRate: 48000, SampleFormat: "s32le"}
	alsaOutput := &AlsaOutput{}

	config.Apply(alsaOutput)

	if alsaOutput.Device != config.Device {
		t.Errorf("AlsaOutput Device should be config Device :\n got: %v\nwant: %v", alsaOutput.Device, config.Device)
	}
	if alsaOutput.SampleRate != config.SampleRate {
		t.Errorf("AlsaOutput SampleRate should be config SampleRate :\n got: %v\nwant: %v", alsaOutput.SampleRate, config.SampleRate)
	}
	if alsaOutput.SampleFormat != Sample32bLittleEndian {
		t.Errorf("AlsaOutput SampleFormat should be parsed config SampleFormat :\n got: %v\nwant: %v", alsaOutput.SampleFormat, Sample32bLittleEndian)
	}
}

func TestUDPOutputConfig_Flags(t *testing.T) {
	config := UDPOutputConfig{}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	config.Flags(flags, "udp")

	flags.Parse(strings.Split("-udp-target=localhost:9000 -udp-opus-bitrate=512000", " "))
	if config.Target != "localhost:9000" {
		t.Errorf("Target should be 'target' flag value :\n got: %v\nwant: %v", config.Target, "localhost:9000")
	}
	if config.Opus.Bitrate != 512000 {
		t.Errorf("Opus.Bitrate should be 'opus-bitrate' flag value :\n got: %v\nwant: %v", config.Opus.Bitrate, 512000)
	}
}

func TestUDPOutputConfig_Apply(t *testing.T) {
	config := UDPOutputConfig{
		Target: "localhost:9000",
		Opus: OpusAudioEncoderConfig{
			Bitrate: 256000,
		},
	}
	udpOutput := &UDPOutput{}

	config.Apply(udpOutput)

	if udpOutput.Target != config.Target {
		t.Errorf("UDPOutput Target should be config Target :\n got: %v\nwant: %v", udpOutput.Target, config.Target)
	}
	if bitrate := udpOutput.Encoder.(*OpusAudioEncoder).Bitrate; bitrate != config.Opus.Bitrate {
		t.Errorf("UDPOutput Target should be config Target :\n got: %v\nwant: %v", bitrate, config.Opus.Bitrate)
	}
}

func TestUDPInputConfig_Flags(t *testing.T) {
	config := UDPInputConfig{}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	config.Flags(flags, "udp")

	flags.Parse(strings.Split("-udp-bind=localhost:9000", " "))
	if config.Bind != "localhost:9000" {
		t.Errorf("Bind should be 'bind' flag value :\n got: %v\nwant: %v", config.Bind, "localhost:9000")
	}
}

func TestUDPInputConfig_Apply(t *testing.T) {
	config := UDPInputConfig{
		Bind: "localhost:9000",
	}
	udpInput := &UDPInput{}

	config.Apply(udpInput)

	if udpInput.Bind != config.Bind {
		t.Errorf("UDPInput Bind should be config Bind :\n got: %v\nwant: %v", udpInput.Bind, config.Bind)
	}
}

func TestOpusAudioEncoderConfig_Flags(t *testing.T) {
	config := OpusAudioEncoderConfig{}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	config.Flags(flags, "opus")

	flags.Parse(strings.Split("-opus-bitrate=512000", " "))
	if config.Bitrate != 512000 {
		t.Errorf("Bitrate should be 'bitrate' flag value :\n got: %v\nwant: %v", config.Bitrate, 512000)
	}
}

func TestOpusAudioEncoderConfig_Apply(t *testing.T) {
	config := OpusAudioEncoderConfig{Bitrate: 256000}
	opusEncoder := &OpusAudioEncoder{}

	config.Apply(opusEncoder)

	if opusEncoder.Bitrate != config.Bitrate {
		t.Errorf("OpusAudioEncoder Bitrate should be config Bitrate :\n got: %v\nwant: %v", opusEncoder.Bitrate, config.Bitrate)
	}
}

func TestHttpServerConfig_Flags(t *testing.T) {
	config := HttpServerConfig{}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	config.Flags(flags, "http")

	flags.Parse(strings.Split("-http-bind=localhost:9000", " "))
	if config.Bind != "localhost:9000" {
		t.Errorf("Bind should be 'bind' flag value :\n got: %v\nwant: %v", config.Bind, "localhost:9000")
	}
}

func TestHttpServerConfig_Apply(t *testing.T) {
	config := HttpServerConfig{Bind: "0.0.0.0:9000"}
	httpEncoder := &HttpServer{}

	config.Apply(httpEncoder)

	if httpEncoder.Bind != config.Bind {
		t.Errorf("HttpServer Bind should be config Bind :\n got: %v\nwant: %v", httpEncoder.Bind, config.Bind)
	}
}
