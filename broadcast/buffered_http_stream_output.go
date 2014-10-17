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
		Log.Debugf("output.Identifier: %s", output.Metrics.prefix)
	}
	return output.Metrics
}

func (output *BufferedHttpStreamOutput) Init() error {
	if output.Identifier == "" {
		output.Identifier = output.defaultIdentifier()
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
		time.Sleep(500 * time.Millisecond)
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
	output.output.ChannelCount = int32(channelCount)
}

func (output *BufferedHttpStreamOutput) SetSampleRate(sampleRate int) {
	output.output.SampleRate = int32(sampleRate)
}

type BufferedHttpStreamOutputConfig struct {
	HttpStreamOutputConfig
	Identifier     string
	BufferDuration time.Duration
}

func (config *BufferedHttpStreamOutputConfig) Flags(flags *flag.FlagSet, prefix string) {
	config.HttpStreamOutputConfig.Flags(flags, prefix)

	flags.StringVar(&config.Identifier, strings.Join([]string{prefix, "id"}, "-"), "", "The identifier of this stream")
	flags.DurationVar(&config.BufferDuration, strings.Join([]string{prefix, "buffer-duration"}, "-"), 10*time.Second, "The maximum duration of saved sound")
}

func (config *BufferedHttpStreamOutputConfig) Apply(bufferedHttpStreamOutput *BufferedHttpStreamOutput) {
	config.HttpStreamOutputConfig.Apply(bufferedHttpStreamOutput.output)
	if config.Identifier != "" {
		bufferedHttpStreamOutput.Identifier = config.Identifier
	}
	bufferedHttpStreamOutput.unfillAudioBuffer.MaxSampleCount = uint32(float64(bufferedHttpStreamOutput.output.SampleRate) * config.BufferDuration.Seconds())
}

func (config *BufferedHttpStreamOutputConfig) Empty() bool {
	return config.Target == ""
}
