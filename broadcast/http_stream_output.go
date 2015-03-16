package broadcast

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"
)

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

	Metrics  *LocalMetrics
	EventLog *LocalEventLog
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

func (output *HttpStreamOutput) eventLog() *LocalEventLog {
	if output.EventLog == nil {
		output.EventLog = &LocalEventLog{}
	}
	return output.EventLog
}

func (output *HttpStreamOutput) createConnection() (err error) {
	if output.dialer == nil {
		return errors.New("No selected Dialer")
	}
	output.metrics().Counter("http.AttemptedConnections").Inc(1)
	output.connection, err = output.dialer.Connect(output)
	if err != nil {
		output.eventLog().NewEvent(fmt.Sprintf("Can't connect : %s", err))
		return err
	}

	if tcpConnection, ok := output.connection.(*net.TCPConn); ok {
		tcpConnection.SetNoDelay(true)
		tcpConnection.SetLinger(0)
	}

	output.updateDeadline()
	output.connectedSince = time.Now()

	output.eventLog().NewEvent("Connected")

	encoder := NewStreamEncoder(output.Format, output)
	encoder.Init()

	Log.Debugf("New encoder")
	output.encoder = encoder
	return nil
}

func (output *HttpStreamOutput) Start() {
	output.eventLog().NewEvent("Start")
	go output.Run()
}

func (output *HttpStreamOutput) Stop() {
	output.started = false
	output.eventLog().NewEvent("Stop")
	output.Reset()
	output.eventLog().NewEvent("Stopped")
}

func (output *HttpStreamOutput) AdminStatus() string {
	if output.started {
		return "started"
	} else {
		return "stopped"
	}
}

func (output *HttpStreamOutput) IsConnected() bool {
	return output.connection != nil
}

func (output *HttpStreamOutput) OperationalStatus() string {
	if output.IsConnected() {
		return "connected"
	} else {
		return "disconnected"
	}
}

func (output *HttpStreamOutput) Run() {
	output.started = true
	output.eventLog().NewEvent("Started")

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
			// audio can be nil when stopped
			if audio != nil && output.encoder != nil {
				output.metrics().Counter("http.Samples").Inc(int64(audio.SampleCount()))
				output.metrics().Gauge("http.ConnectionDuration").Update(int64(output.ConnectionDuration().Seconds()))
				output.encoder.AudioOut(audio)
			}
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

		output.eventLog().NewEvent("Disconnected")
		output.connectedSince = time.Time{}
		output.metrics().Gauge("http.ConnectionDuration").Update(0)
	}

	if output.encoder != nil {
		Log.Debugf("Close encoder")
		output.encoder.Close()
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
