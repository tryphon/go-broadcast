package broadcast

import (
	"bytes"
	"encoding/binary"
	"net"
)

type UDPInput struct {
	Bind string

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

	buffer := bytes.NewBuffer(input.buffer[:readLength])

	var sampleCount int16
	binary.Read(buffer, binary.LittleEndian, &sampleCount)

	channelCount := 2

	audio := NewAudio(int(sampleCount), channelCount)
	for channel := 0; channel < channelCount; channel++ {
		audio.SetSamples(channel, make([]float32, sampleCount))

		for samplePosition := 0; samplePosition < audio.SampleCount(); samplePosition++ {
			var sample float32
			binary.Read(buffer, binary.LittleEndian, &sample)
			audio.Samples(channel)[samplePosition] = sample
		}
	}

	input.audioHandler.AudioOut(audio)

	return nil
}

func (input *UDPInput) Run() {
	for {
		input.Read()
	}
}
