package broadcast

import (
	"errors"
	"flag"
	"net"
	"net/http"
	"strings"
	"time"
)

type HttpStreamOutput struct {
	Target string
	Format AudioFormat

	Provider AudioProvider

	Description *StreamDescription

	encoder StreamEncoder

	client         http.Client
	connection     net.Conn
	connectedSince time.Time

	started bool

	Metrics *LocalMetrics
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
		output.metrics().Counter("http.output.Traffic").Inc(int64(wrote))

		output.updateDeadline()
	} else {
		Log.Printf("End of HTTP stream")
		output.Reset()
	}
	return wrote, err
}

func (output *HttpStreamOutput) metrics() *LocalMetrics {
	if output.Metrics == nil {
		output.Metrics = &LocalMetrics{}
	}
	return output.Metrics
}

func (output *HttpStreamOutput) createConnection() error {
	request, err := http.NewRequest("SOURCE", output.Target, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Content-type", output.Format.ContentType())
	request.Header.Add("User-Agent", "Go Broadcast v0")

	if output.Description != nil {
		for attribute, value := range output.Description.IcecastHeaders() {
			request.Header.Add(attribute, value)
		}
	}

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

	output.connectedSince = time.Now()

	encoder := NewStreamEncoder(output.Format, output)
	encoder.Init()

	output.encoder = encoder
	return nil
}

func (output *HttpStreamOutput) Start() {
	go output.Run()
}

func (output *HttpStreamOutput) Stop() {
	output.started = false
	Log.Printf("Stop")
}

func (output *HttpStreamOutput) Run() {
	output.started = true
	Log.Printf("Start")
	for output.started {
		if output.connection == nil {
			err := output.createConnection()

			if err != nil {
				Log.Printf("HTTP Error : %s", err.Error())
				time.Sleep(output.GetWaitOnError())
			}
		}

		if output.connection != nil && output.encoder != nil {
			audio := output.Provider.Read()
			output.metrics().Counter("http.Samples").Inc(int64(audio.SampleCount()))
			output.metrics().Gauge("http.ConnectionDuration").Update(int64(output.ConnectionDuration().Seconds()))

			output.encoder.AudioOut(audio)
		}
	}
	Log.Printf("Stopped")
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

func (output *HttpStreamOutput) ConnectionDuration() time.Duration {
	if output.connectedSince.IsZero() {
		return time.Duration(0)
	}

	return time.Now().Sub(output.connectedSince)
}

func (output *HttpStreamOutput) SampleRate() int {
	return output.Format.SampleRate
}

type HttpStreamOutputConfig struct {
	Target      string
	Format      string
	Description StreamDescription
}

func NewHttpStreamOutputConfig() HttpStreamOutputConfig {
	return HttpStreamOutputConfig{
		Target: "",
		Format: "ogg/vorbis:vbr(q=5):2:44100",
	}
}

func (config *HttpStreamOutputConfig) Flags(flags *flag.FlagSet, prefix string) {
	defaultConfig := NewHttpStreamOutputConfig()

	flags.StringVar(&config.Target, strings.Join([]string{prefix, "target"}, "-"), "", "The stream URL (ex: http://source:password@stream-in.tryphon.eu:8000/mystream.ogg)")
	flags.StringVar(&config.Format, strings.Join([]string{prefix, "format"}, "-"), defaultConfig.Format, "The stream format")
}

func (config *HttpStreamOutputConfig) Apply(httpStreamOutput *HttpStreamOutput) {
	httpStreamOutput.Target = config.Target
	httpStreamOutput.Format = ParseAudioFormat(config.Format)
	if !config.Description.IsEmpty() {
		httpStreamOutput.Description = &config.Description
	}
}
