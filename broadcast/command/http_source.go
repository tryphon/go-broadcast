package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"projects.tryphon.eu/go-broadcast/broadcast"
)

type HttpSource struct {
	alsaInput         *broadcast.AlsaInput
	httpStreamOutputs *broadcast.HttpStreamOutputs
	httpServer        *broadcast.HttpServer

	config *HttpSourceConfig
}

func (command *HttpSource) Config() HttpSourceConfig {
	command.config.Http = command.httpStreamOutputs.Config()
	return *command.config
}

func (command *HttpSource) ConfigToJSON() []byte {
	jsonBytes, _ := json.Marshal(command.Config())
	return jsonBytes
}

func (command *HttpSource) SaveConfig() error {
	config := command.Config()
	return broadcast.SaveConfig(config.File, config)
}

func (command *HttpSource) checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		broadcast.Log.Printf("Fatal error : %s", err.Error())
		os.Exit(1)
	}
}

func (command *HttpSource) Setup(config *HttpSourceConfig) {
	config.BaseApply(command.httpServer)

	config.Alsa.Apply(command.alsaInput)

	command.httpStreamOutputs.SetChannelCount(command.alsaInput.Channels)
	command.httpStreamOutputs.SetSampleRate(command.alsaInput.SampleRate)

	config.Http.Apply(command.httpStreamOutputs)

	command.config = config
}

func (command *HttpSource) Main(arguments []string) {
	config := HttpSourceConfig{}

	flags := flag.NewFlagSet("httpsource", flag.ExitOnError)
	config.Flags(flags)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s httpsource [options]\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	err := broadcast.LoadConfig(config.File, &config)
	command.checkError(err)

	if config.Empty() {
		broadcast.Log.Printf("No stream is configured")
	}

	command.alsaInput = &broadcast.AlsaInput{}
	command.httpStreamOutputs = broadcast.NewHttpStreamOutputs()

	soundMeterAudioHandler := &broadcast.SoundMeterAudioHandler{
		Output: command.httpStreamOutputs,
	}

	command.alsaInput.SetAudioHandler(&broadcast.ResizeAudio{
		Output:      soundMeterAudioHandler,
		SampleCount: 1024,
	})

	command.httpServer = &broadcast.HttpServer{
		SoundMeterAudioHandler: soundMeterAudioHandler,
	}

	httpStreamOutputsController := broadcast.NewHttpStreamOutputsController(command.httpStreamOutputs)
	command.httpServer.Register("/streams.json", httpStreamOutputsController)
	command.httpServer.Register("/streams/", httpStreamOutputsController)
	command.httpServer.Register("/events.json", httpStreamOutputsController)

	configController := broadcast.NewConfigController(command)
	command.httpServer.Register("/config.json", configController)
	command.httpServer.Register("/config/", configController)

	command.Setup(&config)

	err = command.alsaInput.Init()
	command.checkError(err)

	err = command.httpStreamOutputs.Init()
	command.checkError(err)

	err = command.httpServer.Init()
	command.checkError(err)

	go command.httpStreamOutputs.Run()

	command.alsaInput.Run()
}

type HttpSourceConfig struct {
	broadcast.CommandConfig

	Alsa broadcast.AlsaInputConfig
	Http broadcast.HttpStreamOutputsConfig
}

func (config *HttpSourceConfig) Flags(flags *flag.FlagSet) {
	config.BaseFlags(flags)

	config.Alsa.Flags(flags, "alsa")
	config.Http.Flags(flags, "stream")
}

func (config *HttpSourceConfig) Apply(command *HttpSource) {
	command.Setup(config)
}

func (config *HttpSourceConfig) Empty() bool {
	return config.Http.Empty()
}

func (config *HttpSourceConfig) ToJSON() []byte {
	jsonBytes, _ := json.Marshal(config)
	return jsonBytes
}
