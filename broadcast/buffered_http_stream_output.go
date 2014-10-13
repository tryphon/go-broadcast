package broadcast

import (
	"flag"
	"strings"
	"time"
)

type BufferedHttpStreamOutput struct {
	output            *HttpStreamOutput
	buffer            AudioBuffer
	unfillAudioBuffer *UnfillAudioBuffer
}

func NewBufferedHttpStreamOutput() *BufferedHttpStreamOutput {
	unfillAudioBuffer := &UnfillAudioBuffer{
		Buffer:            &MemoryAudioBuffer{},
		MaxSampleCount:    5 * 44100,
		UnfillSampleCount: 1024,
	}

	return &BufferedHttpStreamOutput{
		output:            &HttpStreamOutput{},
		buffer:            &MutexAudioBuffer{Buffer: unfillAudioBuffer},
		unfillAudioBuffer: unfillAudioBuffer,
	}
}

func (output *BufferedHttpStreamOutput) Init() error {
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
	BufferDuration time.Duration
}

func (config *BufferedHttpStreamOutputConfig) Flags(flags *flag.FlagSet, prefix string) {
	config.httpStreamOutputConfigFlags(flags, prefix)
	flags.DurationVar(&config.BufferDuration, strings.Join([]string{prefix, "buffer-duration"}, "-"), 10*time.Second, "The maximum duration of saved sound")
}

func (config *BufferedHttpStreamOutputConfig) Apply(bufferedHttpStreamOutput *BufferedHttpStreamOutput) {
	config.httpStreamOutputApply(bufferedHttpStreamOutput.output)
	bufferedHttpStreamOutput.unfillAudioBuffer.MaxSampleCount = uint32(float64(bufferedHttpStreamOutput.output.SampleRate) * config.BufferDuration.Seconds())
}
