package broadcast

import (
	"flag"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"time"
)

type BufferedHttpStreamOutput struct {
	Identifier string

	output *HttpStreamOutput
	buffer AudioBuffer

	started bool

	unfillAudioBuffer *UnfillAudioBuffer
	memoryAudioBuffer *MemoryAudioBuffer

	Metrics  *LocalMetrics
	EventLog *LocalEventLog

	config *BufferedHttpStreamOutputConfig

	efficiencyMeter IoEfficiencyMeter
}

func NewBufferedHttpStreamOutput() *BufferedHttpStreamOutput {
	output := BufferedHttpStreamOutput{}

	output.memoryAudioBuffer = &MemoryAudioBuffer{}
	output.unfillAudioBuffer = &UnfillAudioBuffer{
		Buffer:            output.memoryAudioBuffer,
		MaxSampleCount:    5 * 44100,
		UnfillSampleCount: 1024,
	}
	output.buffer = &MutexAudioBuffer{
		Buffer: output.unfillAudioBuffer,
	}
	output.output = &HttpStreamOutput{}

	return &output
}

func (output *BufferedHttpStreamOutput) defaultIdentifier() string {
	return strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(output.output.Target))), 16)
}

func (output *BufferedHttpStreamOutput) metrics() *LocalMetrics {
	if output.Metrics == nil {
		output.Metrics = &LocalMetrics{prefix: fmt.Sprintf("stream-%s", output.Identifier)}
	}
	return output.Metrics
}

func (output *BufferedHttpStreamOutput) eventLog() *LocalEventLog {
	if output.EventLog == nil {
		output.EventLog = &LocalEventLog{Source: fmt.Sprintf("stream-%s", output.Identifier)}
	}
	return output.EventLog
}

func (output *BufferedHttpStreamOutput) Start() {
	output.output.Start()
	output.started = true
}

func (output *BufferedHttpStreamOutput) Stop() {
	output.output.Stop()
	output.started = false
}

func (output *BufferedHttpStreamOutput) Setup(config *BufferedHttpStreamOutputConfig) {
	config.HttpStreamOutputConfig.Apply(output.output)
	if config.Identifier != "" {
		output.Identifier = config.Identifier
	}
	output.unfillAudioBuffer.MaxSampleCount = uint32(float64(output.output.SampleRate()) * config.BufferDuration.Seconds())

	output.config = config
}

func (output *BufferedHttpStreamOutput) Config() BufferedHttpStreamOutputConfig {
	return *output.config
}

func (output *BufferedHttpStreamOutput) Status() BufferedHttpStreamOutputStatus {
	status := BufferedHttpStreamOutputStatus{
		BufferedHttpStreamOutputConfig: output.Config(),
		AdminStatus:                    output.AdminStatus(),
		OperationalStatus:              output.OperationalStatus(),
		ConnectionDuration:             output.output.ConnectionDuration(),
		Efficiency:                     output.efficiencyMeter.Efficiency(),
		Events:                         output.eventLog().Events(),
	}
	if efficiencyHistory := output.efficiencyMeter.History(); !efficiencyHistory.IsEmpty() {
		status.EfficiencyHistory = *efficiencyHistory
	}
	return status
}

func (output *BufferedHttpStreamOutput) AdminStatus() string {
	return output.output.AdminStatus()
}

func (output *BufferedHttpStreamOutput) OperationalStatus() string {
	return output.output.OperationalStatus()
}

func (output *BufferedHttpStreamOutput) Init() error {
	if output.Identifier == "" {
		output.Identifier = output.defaultIdentifier()
		if output.config != nil {
			output.config.Identifier = output.Identifier
		}
	}

	output.memoryAudioBuffer.Metrics = output.metrics()
	output.output.Metrics = output.metrics()
	output.efficiencyMeter.Metrics = output.metrics()

	output.output.EventLog = output.eventLog()
	output.efficiencyMeter.EventLog = output.eventLog()

	output.output.Provider = output

	err := output.output.Init()
	if err != nil {
		return err
	}

	return nil
}

func (output *BufferedHttpStreamOutput) Read() (audio *Audio) {
	audio = output.buffer.Read()
	for audio == nil {
		time.Sleep(100 * time.Millisecond)
		if !output.started {
			return nil
		}

		audio = output.buffer.Read()
	}

	output.efficiencyMeter.Output(int64(audio.SampleCount()))
	return
}

func (output *BufferedHttpStreamOutput) AudioOut(audio *Audio) {
	output.buffer.AudioOut(audio)
	output.efficiencyMeter.Input(int64(audio.SampleCount()))
}

func (output *BufferedHttpStreamOutput) Run() {
	output.output.Run()
}

func (output *BufferedHttpStreamOutput) SetChannelCount(channelCount int) {
	output.output.Format.ChannelCount = channelCount
}

func (output *BufferedHttpStreamOutput) SetSampleRate(sampleRate int) {
	output.output.Format.SampleRate = sampleRate
}

type BufferedHttpStreamOutputConfig struct {
	HttpStreamOutputConfig
	Identifier     string
	BufferDuration time.Duration
}

type BufferedHttpStreamOutputStatus struct {
	BufferedHttpStreamOutputConfig
	AdminStatus        string // "started" / "stopped"
	OperationalStatus  string // "connected" / "disconnected"
	ConnectionDuration time.Duration
	Efficiency         float64
	EfficiencyHistory  IoEfficiencyMeterHistory
	Events             []*Event
}

func NewBufferedHttpStreamOutputConfig() BufferedHttpStreamOutputConfig {
	return BufferedHttpStreamOutputConfig{
		HttpStreamOutputConfig: NewHttpStreamOutputConfig(),
		BufferDuration:         10 * time.Second,
	}
}

func (config *BufferedHttpStreamOutputConfig) Flags(flags *flag.FlagSet, prefix string) {
	config.HttpStreamOutputConfig.Flags(flags, prefix)

	defaultConfig := NewBufferedHttpStreamOutputConfig()
	flags.StringVar(&config.Identifier, strings.Join([]string{prefix, "id"}, "-"), "", "The identifier of this stream")
	flags.DurationVar(&config.BufferDuration, strings.Join([]string{prefix, "buffer-duration"}, "-"), defaultConfig.BufferDuration, "The maximum duration of saved sound")
}

func (config *BufferedHttpStreamOutputConfig) Apply(bufferedHttpStreamOutput *BufferedHttpStreamOutput) {
	bufferedHttpStreamOutput.Setup(config)
}

func (config *BufferedHttpStreamOutputConfig) Empty() bool {
	return config.Target == ""
}
