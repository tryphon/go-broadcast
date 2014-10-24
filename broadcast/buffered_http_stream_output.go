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

	unfillAudioBuffer *UnfillAudioBuffer
	memoryAudioBuffer *MemoryAudioBuffer

	Metrics *LocalMetrics
	config  *BufferedHttpStreamOutputConfig
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

func (output *BufferedHttpStreamOutput) Start() {
	output.output.Start()
}

func (output *BufferedHttpStreamOutput) Stop() {
	output.output.Stop()
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

func (output *BufferedHttpStreamOutput) Init() error {
	if output.Identifier == "" {
		output.Identifier = output.defaultIdentifier()
		if output.config != nil {
			output.config.Identifier = output.Identifier
		}
	}

	output.memoryAudioBuffer.Metrics = output.metrics()
	output.output.Metrics = output.metrics()

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
		audio = output.buffer.Read()
	}
	return
}

func (output *BufferedHttpStreamOutput) AudioOut(audio *Audio) {
	output.buffer.AudioOut(audio)
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
