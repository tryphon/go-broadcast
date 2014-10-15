package broadcast

import (
	"errors"
	"flag"
	metrics "github.com/tryphon/go-metrics"
	"net"
	"net/http"
	"strings"
	"time"
)

type HttpStreamOutput struct {
	Target string
	// FIXME
	Quality      float32
	ChannelCount int32
	SampleRate   int32
	Format       string

	Provider AudioProvider

	encoder StreamEncoder

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
	connection, err := net.DialTimeout(network, addr, output.GetWriteTimeout())
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
	output.connection.SetWriteDeadline(time.Now().Add(output.GetWriteTimeout()))
}

func (output *HttpStreamOutput) Write(buffer []byte) (int, error) {
	if output.connection == nil {
		return len(buffer), nil
	}

	wrote, err := output.connection.Write(buffer)
	if err == nil {
		metrics.GetOrRegisterCounter("http.output.Traffic", nil).Inc(int64(wrote))
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

	request.Header.Add("Content-type", output.contentType())
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

	encoder := output.newEncoder()
	encoder.Init()

	output.encoder = encoder
	return nil
}

func (output *HttpStreamOutput) contentType() string {
	if output.Format == "mp3" {
		return "audio/mpeg"
	} else {
		return "application/ogg"
	}
}

func (output *HttpStreamOutput) newEncoder() StreamEncoder {
	if output.Format == "mp3" {
		encoder := LameEncoder{
			SampleRate:   int(output.SampleRate),
			ChannelCount: int(output.ChannelCount),
			Quality:      output.Quality,
			Writer:       output,
		}
		return &encoder
	} else {
		encoder := OggEncoder{
			Encoder: VorbisEncoder{
				Quality:      output.Quality,
				ChannelCount: output.ChannelCount,
				SampleRate:   output.SampleRate,
			},
			Writer: output,
		}
		encoder.Encoder.PacketHandler = &encoder
		return &encoder
	}
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

		if output.connection != nil && output.encoder != nil {
			audio := output.Provider.Read()
			output.encoder.AudioOut(audio)
		}
	}
}

func (output *HttpStreamOutput) GetWriteTimeout() time.Duration {
	return 30 * time.Second
}

func (output *HttpStreamOutput) GetWaitOnError() time.Duration {
	return 10 * time.Second
}

func (output *HttpStreamOutput) Reset() {
	if output.connection != nil {
		output.connection.Close()
		output.connection = nil
		output.encoder = nil
	}
}

type HttpStreamOutputConfig struct {
	Target string
	// FIXME
	Quality int
	Format  string
}

func (config *HttpStreamOutputConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.Target, strings.Join([]string{prefix, "target"}, "-"), "", "The stream URL (ex: http://source:password@stream-in.tryphon.eu:8000/mystream.ogg)")
	flags.IntVar(&config.Quality, strings.Join([]string{prefix, "quality"}, "-"), 5, "The stream quality")
	flags.StringVar(&config.Format, strings.Join([]string{prefix, "format"}, "-"), "ogg/vorbis", "The stream format")
}

func (config *HttpStreamOutputConfig) Apply(httpStreamOutput *HttpStreamOutput) {
	httpStreamOutput.Target = config.Target
	httpStreamOutput.Quality = float32(config.Quality / 10.0)
	httpStreamOutput.Format = config.Format
}
