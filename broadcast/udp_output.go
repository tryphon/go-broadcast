package broadcast

import (
	"net"
)

type UDPOutput struct {
	Target     string
	Encoder    AudioEncoder
	PacketSampleCount int

	connection net.Conn
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
		opusEncoder := &OpusAudioEncoder{}
		err = opusEncoder.Init()
		if err != nil {
			return err
		}

		output.Encoder = opusEncoder
	}

	if output.PacketSampleCount == 0 {
		output.PacketSampleCount = 960
	}

	audioHandler := AudioHandlerFunc(func(audio *Audio) {
		output.audioOut(audio)
	})
	output.resizeAudio = &ResizeAudio{
		Output: audioHandler,
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

	_, err = output.connection.Write(bytes)
	if err != nil {
		Log.Printf("Can't write data in UDP socket: %s", err.Error())
	}
}
