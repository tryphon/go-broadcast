package broadcast

import (
	"flag"
	"strings"
	"time"
)

type UDPClientConfig struct {
	Alsa AlsaInputConfig
	Udp  UDPOutputConfig
	Http HttpServerConfig
}

func (config *UDPClientConfig) Flags(flags *flag.FlagSet) {
	config.Alsa.Flags(flags, "alsa")
	config.Udp.Flags(flags, "udp")
	config.Http.Flags(flags, "http")
}

func (config *UDPClientConfig) Apply(alsaInput *AlsaInput, udpOutput *UDPOutput, httpServer *HttpServer) {
	config.Alsa.Apply(alsaInput)
	config.Udp.Apply(udpOutput)
	config.Http.Apply(httpServer)
}

type AlsaInputConfig struct {
	Device         string
	SampleRate     int
	BufferDuration time.Duration
	SampleFormat   string
}

func (config *AlsaInputConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.Device, strings.Join([]string{prefix, "device"}, "-"), "default", "The alsa device used to record sound")
	flags.IntVar(&config.SampleRate, strings.Join([]string{prefix, "sample-rate"}, "-"), 44100, "Sample rate")
	flags.DurationVar(&config.BufferDuration, strings.Join([]string{prefix, "buffer-duration"}, "-"), 250*time.Millisecond, "The alsa buffer duration")
	flags.StringVar(&config.SampleFormat, strings.Join([]string{prefix, "sample-format"}, "-"), "auto", "The sample format used to record sound (s16le, s32le, s32be)")
}

func (config *AlsaInputConfig) Apply(alsaInput *AlsaInput) {
	alsaInput.Device = config.Device
	alsaInput.SampleRate = config.SampleRate

	bufferSampleCount := int(float64(config.SampleRate) * config.BufferDuration.Seconds())
	alsaInput.BufferSampleCount = bufferSampleCount
	alsaInput.SampleFormat = ParseSampleFormat(config.SampleFormat)
}

type UDPOutputConfig struct {
	Target string
	Opus   OpusAudioEncoderConfig
}

func (config *UDPOutputConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.Target, strings.Join([]string{prefix, "target"}, "-"), "", "The host:port where UDP stream is sent")
	config.Opus.Flags(flags, strings.Join([]string{prefix, "opus"}, "-"))
}

func (config *UDPOutputConfig) Apply(udpOutput *UDPOutput) {
	udpOutput.Target = config.Target

	if udpOutput.Encoder == nil {
		udpOutput.Encoder = &OpusAudioEncoder{}
	}
	config.Opus.Apply(udpOutput.Encoder.(*OpusAudioEncoder))
}

type OpusAudioEncoderConfig struct {
	Bitrate int
}

func (config *OpusAudioEncoderConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.IntVar(&config.Bitrate, strings.Join([]string{prefix, "bitrate"}, "-"), 256000, "The Opus stream bitrate")
}

func (config *OpusAudioEncoderConfig) Apply(opusEncoder *OpusAudioEncoder) {
	opusEncoder.Bitrate = config.Bitrate
}

type HttpServerConfig struct {
	Bind string
}

func (config *HttpServerConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.Bind, strings.Join([]string{prefix, "bind"}, "-"), "", "'[address]:port' where the HTTP server is bind")
}

func (config *HttpServerConfig) Apply(httpServer *HttpServer) {
	httpServer.Bind = config.Bind
}
