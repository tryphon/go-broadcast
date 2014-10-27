package broadcast

import (
	"errors"
	"flag"
	shoutcast "github.com/tryphon/go-shoutcast"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpStreamDialer interface {
	Connect(output *HttpStreamOutput) (net.Conn, error)
}

type Icecast2Dialer struct {
}

func (dialer *Icecast2Dialer) Connect(output *HttpStreamOutput) (net.Conn, error) {
	var connection net.Conn

	dialTimeout := func(network, addr string) (net.Conn, error) {
		newConnection, err := net.DialTimeout(network, addr, output.GetWriteTimeout())
		if err != nil {
			return nil, err
		}

		connection = newConnection
		return newConnection, nil
	}

	transport := http.Transport{Dial: dialTimeout}
	client := http.Client{Transport: &transport}

	request, err := http.NewRequest("SOURCE", output.Target, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-type", output.Format.ContentType())
	request.Header.Add("User-Agent", "Go Broadcast v0")

	if output.Description != nil {
		for attribute, value := range output.Description.IcecastHeaders() {
			Log.Debugf("IceCast header: %s=%s", attribute, value)
			request.Header.Add(attribute, value)
		}
	}

	// request.SetBasicAuth("source", password)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	Log.Debugf("HTTP Response : %s", response.Status)
	if response.Status != "200 OK" {
		err = errors.New("HTTP Error")
		return nil, err
	}

	return connection, nil
}

type ShoutcastDialer struct {
}

func (dialer *ShoutcastDialer) Client(output *HttpStreamOutput) (*shoutcast.Client, error) {
	targetURL, err := url.Parse(output.Target)
	if err != nil {
		return nil, err
	}
	password, ok := targetURL.User.Password()
	if !ok {
		return nil, errors.New("No specified password")
	}

	description := output.Description
	if description == nil {
		description = &StreamDescription{}
	}
	headers := description.ShoutcastHeaders()
	headers["content-type"] = output.Format.ContentType()

	Log.Debugf("ShoutCast headers: %v", headers)

	client := &shoutcast.Client{
		Host:     targetURL.Host,
		Password: password,
		Timeout:  output.GetWriteTimeout(),
		Headers:  headers,
	}
	return client, nil
}

func (dialer *ShoutcastDialer) Connect(output *HttpStreamOutput) (net.Conn, error) {
	client, err := dialer.Client(output)
	if err != nil {
		return nil, err
	}
	return client.Connect()
}

type HttpStreamOutput struct {
	Target     string
	Format     AudioFormat
	ServerType string

	Provider AudioProvider

	Description *StreamDescription

	encoder StreamEncoder

	dialer         HttpStreamDialer
	connection     net.Conn
	connectedSince time.Time

	started bool

	Metrics *LocalMetrics
}

func (output *HttpStreamOutput) Init() error {
	if output.ServerType == "shoutcast" {
		output.dialer = &ShoutcastDialer{}
	} else {
		output.dialer = &Icecast2Dialer{}
	}
	return nil
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
		Log.Printf("End of HTTP stream (%v)", err)
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

func (output *HttpStreamOutput) createConnection() (err error) {
	if output.dialer == nil {
		return errors.New("No selected Dialer")
	}
	output.connection, err = output.dialer.Connect(output)
	if err != nil {
		return err
	}

	if tcpConnection, ok := output.connection.(*net.TCPConn); ok {
		tcpConnection.SetNoDelay(true)
		tcpConnection.SetLinger(0)
	}

	output.updateDeadline()
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
				Log.Printf("Connection Error : %s", err.Error())
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
	ServerType  string
}

func NewHttpStreamOutputConfig() HttpStreamOutputConfig {
	return HttpStreamOutputConfig{
		Target:     "",
		Format:     "ogg/vorbis:vbr(q=5):2:44100",
		ServerType: "icecast2",
	}
}

func (config *HttpStreamOutputConfig) Flags(flags *flag.FlagSet, prefix string) {
	defaultConfig := NewHttpStreamOutputConfig()

	flags.StringVar(&config.Target, strings.Join([]string{prefix, "target"}, "-"), "", "The stream URL (ex: http://source:password@stream-in.tryphon.eu:8000/mystream.ogg)")
	flags.StringVar(&config.Format, strings.Join([]string{prefix, "format"}, "-"), defaultConfig.Format, "The stream format")
	flags.StringVar(&config.ServerType, strings.Join([]string{prefix, "servertype"}, "-"), defaultConfig.ServerType, "The type of stream server (icecast2 or shoutcast)")
}

func (config *HttpStreamOutputConfig) Apply(httpStreamOutput *HttpStreamOutput) {
	httpStreamOutput.Target = config.Target
	httpStreamOutput.Format = ParseAudioFormat(config.Format)
	httpStreamOutput.ServerType = config.ServerType
	if !config.Description.IsEmpty() {
		httpStreamOutput.Description = &config.Description
	} else {
		if httpStreamOutput.ServerType == "shoutcast" {
			httpStreamOutput.Description = &StreamDescription{}
		}
	}

	if httpStreamOutput.Description != nil {
		Log.Debugf("Define BitRate in description (%d)", httpStreamOutput.Format.BitRate)
		httpStreamOutput.Description.BitRate = httpStreamOutput.Format.BitRate
	}
}
