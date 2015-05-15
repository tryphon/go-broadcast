package broadcast

import (
	"flag"
	"strings"
	"time"
)

type BufferedHttpStreamInput struct {
	http   HttpInput
	buffer *BufferHttpStreamInput
}

func NewBufferedHttpStreamInput() *BufferedHttpStreamInput {
	return &BufferedHttpStreamInput{
		buffer: NewBufferHttpStreamInput(),
	}
}

func (input *BufferedHttpStreamInput) Setup(config *BufferedHttpStreamInputConfig) {
	input.http.Setup(&config.HttpStreamInputConfig)
	input.buffer.Setup(&config.Buffer)
}

func (input *BufferedHttpStreamInput) Init() error {
	input.http.SetAudioHandler(
		&ResizeAudio{
			Output:      input.buffer,
			SampleCount: 1024,
		},
	)

	input.http.Init()

	return nil
}

func (input *BufferedHttpStreamInput) Run() {
	input.http.Run()
}

func (input *BufferedHttpStreamInput) Read() *Audio {
	return input.buffer.Read()
}

func (input *BufferedHttpStreamInput) SetChannelCount(channelCount int) {
	input.buffer.SetChannelCount(channelCount)
}

func (input *BufferedHttpStreamInput) SetSampleRate(sampleRate int) {
	input.buffer.SetSampleRate(sampleRate)
}

type BufferedHttpStreamInputConfig struct {
	HttpStreamInputConfig
	Buffer BufferHttpStreamInputConfig
}

func (config *BufferedHttpStreamInputConfig) Flags(flags *flag.FlagSet, prefix string) {
	config.HttpStreamInputConfig.Flags(flags, prefix)
	config.Buffer.Flags(flags, strings.Join([]string{prefix, "buffer"}, "-"))
}

func (config *BufferedHttpStreamInputConfig) Apply(input *BufferedHttpStreamInput) {
	input.Setup(config)
}

type BufferHttpStreamInput struct {
	sampleRate int

	lowAdjustBuffer *AdjustAudioBuffer
	lowRefillBuffer *RefillAudioBuffer

	highAdjustBuffer *AdjustAudioBuffer
	highUnfillBuffer *UnfillAudioBuffer

	mutexBuffer *MutexAudioBuffer

	config *BufferHttpStreamInputConfig
}

func NewBufferHttpStreamInput() *BufferHttpStreamInput {
	buffer := &BufferHttpStreamInput{}

	buffer.lowAdjustBuffer = &AdjustAudioBuffer{
		Buffer: &MemoryAudioBuffer{},
	}
	buffer.lowRefillBuffer = &RefillAudioBuffer{
		Buffer: buffer.lowAdjustBuffer,
	}
	buffer.highAdjustBuffer = &AdjustAudioBuffer{
		Buffer: buffer.lowRefillBuffer,
	}
	buffer.highUnfillBuffer = &UnfillAudioBuffer{
		Buffer: buffer.highAdjustBuffer,
	}
	buffer.mutexBuffer = &MutexAudioBuffer{
		Buffer: buffer.highUnfillBuffer,
	}

	buffer.Setup(NewBufferHttpStreamInputConfig())

	return buffer
}

func (buffer *BufferHttpStreamInput) SampleRate() int {
	if buffer.sampleRate == 0 {
		buffer.sampleRate = DefaultSampleRate
	}
	return buffer.sampleRate
}

func (buffer *BufferHttpStreamInput) SetChannelCount(channelCount int) {
	buffer.highAdjustBuffer.ChannelCount = channelCount
}

func (buffer *BufferHttpStreamInput) SetSampleRate(sampleRate int) {
	buffer.sampleRate = sampleRate

	// Setup again buffers with current config
	buffer.Setup(buffer.config)
}

func (buffer *BufferHttpStreamInput) Setup(config *BufferHttpStreamInputConfig) {
	sampleRate := float64(buffer.SampleRate())
	sampleDuration := config.Duration.Seconds() * sampleRate

	Log.Debugf("Sample duration : %v", sampleDuration)

	buffer.lowAdjustBuffer.LimitSampleCount = 0
	buffer.lowAdjustBuffer.ThresholdSampleCount = uint32(config.LowAdjustThreshold / 100 * sampleDuration)

	buffer.lowRefillBuffer.MinSampleCount = uint32(config.LowRefill / 100 * sampleDuration)

	buffer.highAdjustBuffer.LimitSampleCount = uint32(sampleDuration)
	buffer.highAdjustBuffer.ThresholdSampleCount = uint32(config.HighAdjustThreshold / 100 * sampleDuration)

	buffer.highUnfillBuffer.UnfillSampleCount = uint32(config.HighUnfill / 100 * sampleDuration)
	buffer.highUnfillBuffer.MaxSampleCount = uint32(sampleDuration)

	Log.Debugf("Buffer setup: lowAdjustBuffer.ThresholdSampleCount %d lowRefillBuffer.MinSampleCount %d", buffer.lowAdjustBuffer.ThresholdSampleCount, buffer.lowRefillBuffer.MinSampleCount)

	buffer.config = config
}

func (buffer *BufferHttpStreamInput) AudioOut(audio *Audio) {
	buffer.mutexBuffer.AudioOut(audio)
}

func (buffer *BufferHttpStreamInput) Read() *Audio {
	return buffer.mutexBuffer.Read()
}

func (buffer *BufferHttpStreamInput) SampleCount() uint32 {
	return buffer.mutexBuffer.SampleCount()
}

type BufferHttpStreamInputConfig struct {
	Duration time.Duration

	LowAdjustThreshold  float64
	LowRefill           float64
	HighAdjustThreshold float64
	HighUnfill          float64
}

func NewBufferHttpStreamInputConfig() *BufferHttpStreamInputConfig {
	return &BufferHttpStreamInputConfig{
		Duration:            10 * time.Second,
		LowAdjustThreshold:  30,
		LowRefill:           50,
		HighAdjustThreshold: 70,
		HighUnfill:          30,
	}
}

func (config *BufferHttpStreamInputConfig) Flags(flags *flag.FlagSet, prefix string) {
	defaultConfig := NewBufferHttpStreamInputConfig()

	flags.DurationVar(&config.Duration, strings.Join([]string{prefix, "duration"}, "-"), defaultConfig.Duration, "Max duration of buffer (in seconds)")

	flags.Float64Var(&config.LowAdjustThreshold, strings.Join([]string{prefix, "low-adjust-threshold"}, "-"), defaultConfig.LowAdjustThreshold, "Threshold of low adjust buffer (% of the buffer)")
	flags.Float64Var(&config.LowRefill, strings.Join([]string{prefix, "low-refill"}, "-"), defaultConfig.LowRefill, "Duration to refill when buffer is empty (% of the buffer)")

	flags.Float64Var(&config.HighAdjustThreshold, strings.Join([]string{prefix, "high-adjust-threshold"}, "-"), defaultConfig.HighAdjustThreshold, "Threshold of high adjust buffer (% of the buffer)")
	flags.Float64Var(&config.HighUnfill, strings.Join([]string{prefix, "high-unfill"}, "-"), defaultConfig.HighUnfill, "Duration to unfill when buffer is full (% of the buffer)")
}
