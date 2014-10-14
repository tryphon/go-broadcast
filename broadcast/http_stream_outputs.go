package broadcast

import (
	"flag"
	"fmt"
)

type HttpStreamOutputs struct {
	streams      []*BufferedHttpStreamOutput
	channelCount int
	sampleRate   int
}

func NewHttpStreamOutputs() *HttpStreamOutputs {
	return &HttpStreamOutputs{
		streams: make([]*BufferedHttpStreamOutput, 0),
	}
}

func (output *HttpStreamOutputs) AudioOut(audio *Audio) {
	for _, stream := range output.streams {
		stream.AudioOut(audio)
	}
}

func (output *HttpStreamOutputs) Init() error {
	Log.Debugf("Initialize %d stream(s)", len(output.streams))
	for _, stream := range output.streams {
		stream.Init()
	}
	return nil
}

func (output *HttpStreamOutputs) Run() {
	for _, stream := range output.streams {
		Log.Debugf("Run Stream %s", stream.output.Target)
		go stream.Run()
	}
}

func (output *HttpStreamOutputs) Create() *BufferedHttpStreamOutput {
	stream := NewBufferedHttpStreamOutput()
	output.streams = append(output.streams, stream)
	return stream
}

func (output *HttpStreamOutputs) SetChannelCount(channelCount int) {
	output.channelCount = channelCount
}

func (output *HttpStreamOutputs) SetSampleRate(sampleRate int) {
	output.sampleRate = sampleRate
}

type HttpStreamOutputsConfig struct {
	Streams []*BufferedHttpStreamOutputConfig
}

func (config *HttpStreamOutputsConfig) Empty() bool {
	if config.Streams != nil {
		for _, streamConfig := range config.Streams {
			if !streamConfig.Empty() {
				return false
			}
		}
	}
	return true
}

func (config *HttpStreamOutputsConfig) Flags(flags *flag.FlagSet, prefix string) {
	commandLineStreamCount := 4
	config.Streams = make([]*BufferedHttpStreamOutputConfig, commandLineStreamCount)

	for streamNumber := 1; streamNumber <= commandLineStreamCount; streamNumber++ {
		streamConfig := BufferedHttpStreamOutputConfig{}
		streamConfig.Flags(flags, fmt.Sprintf("%s-%d", prefix, streamNumber))
		config.Streams[streamNumber-1] = &streamConfig
	}
}

func (config *HttpStreamOutputsConfig) Apply(httpStreamOutputs *HttpStreamOutputs) {
	for _, streamConfig := range config.Streams {
		if !streamConfig.Empty() {
			stream := httpStreamOutputs.Create()
			streamConfig.Apply(stream)
		}
	}
}
