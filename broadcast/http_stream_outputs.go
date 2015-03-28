package broadcast

import (
	"flag"
	"fmt"
	"hash/crc32"
	"strconv"
)

type HttpStreamOutputs struct {
	streams      []*BufferedHttpStreamOutput
	channelCount int
	sampleRate   int

	config *HttpStreamOutputsConfig
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

func (output *HttpStreamOutputs) Start() {
	for _, stream := range output.streams {
		Log.Debugf("Start Stream %s", stream.output.Target)
		stream.Start()
	}
}

func (output *HttpStreamOutputs) Stop() {
	for _, stream := range output.streams {
		Log.Debugf("Stop Stream %s", stream.output.Target)
		stream.Stop()
	}
}

func (output *HttpStreamOutputs) Run() {
	for _, stream := range output.streams {
		Log.Debugf("Run Stream %s", stream.output.Target)
		go stream.Run()
	}
}

func (output *HttpStreamOutputs) Create(config *BufferedHttpStreamOutputConfig) *BufferedHttpStreamOutput {
	stream := NewBufferedHttpStreamOutput()
	if config == nil {
		config = &BufferedHttpStreamOutputConfig{}
	}

	output.uniqStreamIdentifier(config)
	stream.Setup(config)

	output.streams = append(output.streams, stream)
	return stream
}

func (output *HttpStreamOutputs) uniqStreamIdentifier(config *BufferedHttpStreamOutputConfig) {
	if config.Identifier == "" && config.Target != "" {
		config.Identifier = strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(config.Target))), 16)
	}

	validate := func(identifier string) bool {
		if identifier == "" {
			return false
		}

		for _, stream := range output.streams {
			if stream.Identifier == identifier {
				return false
			}
		}

		return true
	}

	if !validate(config.Identifier) {
		generator := IdentifierGenerator{
			Prefix:   config.Identifier,
			Validate: validate,
		}
		config.Identifier = generator.Generate()
		Log.Debugf("Change stream identifier : %s", config.Identifier)
	}
}

type IdentifierValidator func(identifier string) bool

type IdentifierGenerator struct {
	Prefix   string
	Index    int
	Validate IdentifierValidator
}

func (generator *IdentifierGenerator) NextIdentifier() string {
	generator.Index += 1
	return generator.Identifier()
}

func (generator *IdentifierGenerator) Identifier() string {
	if generator.Prefix != "" {
		return fmt.Sprintf("%s-%d", generator.Prefix, generator.Index)
	} else {
		return strconv.Itoa(generator.Index)
	}
}

func (generator *IdentifierGenerator) Generate() string {
	for !generator.Validate(generator.NextIdentifier()) {

	}
	return generator.Identifier()
}

func (output *HttpStreamOutputs) SetChannelCount(channelCount int) {
	output.channelCount = channelCount
}

func (output *HttpStreamOutputs) SetSampleRate(sampleRate int) {
	output.sampleRate = sampleRate
}

func (output *HttpStreamOutputs) Config() HttpStreamOutputsConfig {
	output.config.Streams = make([]BufferedHttpStreamOutputConfig, 0)
	for _, stream := range output.streams {
		output.config.Streams = append(output.config.Streams, stream.Config())
	}
	return *output.config
}

func (output *HttpStreamOutputs) Status() HttpStreamOutputsStatus {
	status := HttpStreamOutputsStatus{
		Streams: make([]BufferedHttpStreamOutputStatus, 0),
		Events:  EventLog.Events(),
	}
	for _, stream := range output.streams {
		status.Streams = append(status.Streams, stream.Status())
	}
	return status
}

func (output *HttpStreamOutputs) Setup(config *HttpStreamOutputsConfig) {
	config.Compact()
	for index, _ := range config.Streams {
		streamConfig := config.Streams[index]
		if !streamConfig.Empty() {
			output.Create(&streamConfig)
		}
	}
	output.config = config
}

func (output *HttpStreamOutputs) Stream(identifier string) *BufferedHttpStreamOutput {
	for _, stream := range output.streams {
		if stream.Identifier == identifier {
			return stream
		}
	}
	return nil
}

func (output *HttpStreamOutputs) Destroy(identifier string) *BufferedHttpStreamOutput {
	for index, stream := range output.streams {
		if stream.Identifier == identifier {
			output.streams[index], output.streams = output.streams[len(output.streams)-1], output.streams[:len(output.streams)-1]
			return stream
		}
	}
	return nil
}

type HttpStreamOutputsConfig struct {
	Streams []BufferedHttpStreamOutputConfig
}

type HttpStreamOutputsStatus struct {
	Streams []BufferedHttpStreamOutputStatus
	Events  []*Event
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

func (config *HttpStreamOutputsConfig) Compact() {
	notEmptyStreams := []BufferedHttpStreamOutputConfig{}
	for _, streamConfig := range config.Streams {
		if !streamConfig.Empty() {
			notEmptyStreams = append(notEmptyStreams, streamConfig)
		}
	}
	config.Streams = notEmptyStreams
}

func (config *HttpStreamOutputsConfig) Flags(flags *flag.FlagSet, prefix string) {
	config.Streams = make([]BufferedHttpStreamOutputConfig, 4)

	for index, _ := range config.Streams {
		config.Streams[index].Flags(flags, fmt.Sprintf("%s-%d", prefix, index+1))
	}
}

func (config *HttpStreamOutputsConfig) Apply(httpStreamOutputs *HttpStreamOutputs) {
	httpStreamOutputs.Setup(config)
}
