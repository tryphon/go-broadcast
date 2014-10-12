package broadcast

import (
	"flag"
	metrics "github.com/rcrowley/go-metrics"
	"net"
	"strings"
)

type UDPInput struct {
	Bind    string
	Decoder AudioDecoder

	connection   *net.UDPConn
	audioHandler AudioHandler

	bufferLength int
	buffer       []byte
}

func (input *UDPInput) Init() (err error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", input.Bind)
	if err != nil {
		return err
	}

	connection, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	input.connection = connection
	input.bufferLength = 4096
	input.buffer = make([]byte, input.bufferLength)

	opusDecoder := &OpusAudioDecoder{}
	err = opusDecoder.Init()
	if err != nil {
		return err
	}

	input.Decoder = opusDecoder

	return nil
}

func (input *UDPInput) SetAudioHandler(audioHandler AudioHandler) {
	input.audioHandler = audioHandler
}

func (input *UDPInput) Read() (err error) {
	readLength, _, err := input.connection.ReadFromUDP(input.buffer)

	if err != nil {
		Log.Printf("Can't read data in UDP socket: %s", err.Error())
		return err
	}

	metrics.GetOrRegisterCounter("udp.input.PacketCount", nil).Inc(1)
	metrics.GetOrRegisterCounter("udp.input.Traffic", nil).Inc(int64(readLength))

	audio, err := input.Decoder.Decode(input.buffer[:readLength])
	if err != nil {
		Log.Printf("Can't decode data from UDP socket: %s", err.Error())
		return err
	}
	input.audioHandler.AudioOut(audio)

	return nil
}

func (input *UDPInput) Run() {
	for {
		input.Read()
	}
}

type UDPInputConfig struct {
	Bind string
}

func (config *UDPInputConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.Bind, strings.Join([]string{prefix, "bind"}, "-"), ":9090", "The [address]:port where UDP stream is received")
}

func (config *UDPInputConfig) Apply(udpInput *UDPInput) {
	udpInput.Bind = config.Bind
}
