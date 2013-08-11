package broadcast

import (
	"bytes"
	"encoding/binary"
	"net"
)

type UDPOutput struct {
	Target     string
	connection net.Conn
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
	return nil
}

func (output *UDPOutput) AudioOut(audio *Audio) {
	buffer := &bytes.Buffer{}
	binary.Write(buffer, binary.LittleEndian, int16(audio.SampleCount()))

	for channel := 0; channel < audio.ChannelCount(); channel++ {
		for samplePosition := 0; samplePosition < audio.SampleCount(); samplePosition++ {
			binary.Write(buffer, binary.LittleEndian, audio.Samples(channel)[samplePosition])
		}
	}

	_, err := buffer.WriteTo(output.connection)
	if err != nil {
		Log.Printf("Can't write data in UDP socket: %s", err.Error())
	}
}
