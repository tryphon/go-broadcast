package broadcast

import (
	"flag"
	metrics "github.com/tryphon/go-metrics"
	"net"
	"strings"
)

type UDPOutput struct {
	Target            string
	Encoder           AudioEncoder
	PacketSampleCount int

	connection  net.Conn
	resizeAudio *ResizeAudio
}

func (output *UDPOutput) Init() (err error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", output.Target)
	if err != nil {
		return err
	}

	connection, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	output.connection = connection

	if output.Encoder == nil {
		output.Encoder = &OpusAudioEncoder{}
	}

	err = output.Encoder.Init()
	if err != nil {
		return err
	}

	if output.PacketSampleCount == 0 {
		output.PacketSampleCount = 960
	}

	audioHandler := AudioHandlerFunc(func(audio *Audio) {
		output.audioOut(audio)
	})
	output.resizeAudio = &ResizeAudio{
		Output:       audioHandler,
		SampleCount:  output.PacketSampleCount,
		ChannelCount: 2,
	}

	return nil
}

func (output *UDPOutput) AudioOut(audio *Audio) {
	output.resizeAudio.AudioOut(audio)
}

func (output *UDPOutput) audioOut(audio *Audio) {
	bytes, err := output.Encoder.Encode(audio)
	if err != nil {
		Log.Printf("Can't encode audio: %s", err.Error())
	}

	metrics.GetOrRegisterCounter("udp.output.PacketCount", nil).Inc(1)

	wroteLength, err := output.connection.Write(bytes)
	if err != nil {
		Log.Printf("Can't write data in UDP socket: %s", err.Error())
	}

	metrics.GetOrRegisterCounter("udp.output.Traffic", nil).Inc(int64(wroteLength))
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
