package broadcast

import (
	"errors"
	"net"
	"net/http"
	"time"
)

type HttpStreamOutput struct {
	Target string
	// FIXME
	Quality      float32
	ChannelCount int32
	SampleRate   int32

	Provider AudioProvider

	encoder *OggEncoder

	client     http.Client
	connection net.Conn
}

func (output *HttpStreamOutput) Init() error {
	transport := http.Transport{
		Dial: output.dialTimeout,
	}

	output.client = http.Client{
		Transport: &transport,
	}

	return nil
}

func (output *HttpStreamOutput) dialTimeout(network, addr string) (net.Conn, error) {
	connection, err := net.DialTimeout(network, addr, output.GetReadTimeout())
	if err != nil {
		output.connection = nil
		return nil, err
	}

	if tcpConnection, ok := connection.(*net.TCPConn); ok {
		tcpConnection.SetNoDelay(true)
		tcpConnection.SetLinger(0)
	}

	output.connection = connection
	output.updateDeadline()

	return connection, nil
}

func (output *HttpStreamOutput) updateDeadline() {
	output.connection.SetWriteDeadline(time.Now().Add(output.GetReadTimeout()))
}

func (output *HttpStreamOutput) Write(buffer []byte) (int, error) {
	if output.connection == nil {
		return len(buffer), nil
	}

	wrote, err := output.connection.Write(buffer)
	if err == nil {
		output.updateDeadline()
	} else {
		Log.Printf("End of HTTP stream")
		output.Reset()
	}
	return wrote, err
}

func (output *HttpStreamOutput) createConnection() error {
	request, err := http.NewRequest("SOURCE", output.Target, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Content-type", "application/ogg")
	request.Header.Add("User-Agent", "Go Broadcast v0")

	// request.SetBasicAuth("source", password)

	response, err := output.client.Do(request)
	if err != nil {
		return err
	}

	Log.Debugf("HTTP Response : %s", response.Status)
	if response.Status != "200 OK" {
		err = errors.New("HTTP Error")
		return err
	}

	encoder := OggEncoder{
		Encoder: VorbisEncoder{
			Quality:      output.Quality,
			ChannelCount: output.ChannelCount,
			SampleRate:   output.SampleRate,
		},
		Writer: output,
	}
	encoder.Encoder.PacketHandler = &encoder
	encoder.Init()

	output.encoder = &encoder
	return nil
}

func (output *HttpStreamOutput) Run() {
	for {
		if output.connection == nil {
			err := output.createConnection()

			if err != nil {
				Log.Printf("HTTP Error : %s", err.Error())
				time.Sleep(output.GetWaitOnError())
			}
		}

		if output.connection != nil {
			audio := output.Provider.Read()
			output.encoder.AudioOut(audio)
		}
	}
}

func (output *HttpStreamOutput) GetReadTimeout() time.Duration {
	return 10 * time.Second
}

func (output *HttpStreamOutput) GetWaitOnError() time.Duration {
	return 10 * time.Second
}

func (output *HttpStreamOutput) Reset() {
	if output.connection != nil {
		output.connection.Close()
		output.connection = nil
	}

	// input.oggDecoder.Reset()
	// input.vorbisDecoder.Reset()
}
