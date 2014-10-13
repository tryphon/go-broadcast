package command

import (
	"flag"
	"fmt"
	"os"
	"projects.tryphon.eu/go-broadcast/broadcast"
)

type HttpSource struct {
}

func (command *HttpSource) checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		broadcast.Log.Printf("Fatal error : %s", err.Error())
		os.Exit(1)
	}
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

	if config.Stream.Target == "" {
		flag.Usage()
		os.Exit(1)
	}

	broadcast.Log.Printf("Config: %v", config)

	alsaInput := &broadcast.AlsaInput{}
	httpStreamOutput := broadcast.NewBufferedHttpStreamOutput()

	soundMeterAudioHandler := &broadcast.SoundMeterAudioHandler{
		Output: httpStreamOutput,
	}

	alsaInput.SetAudioHandler(&broadcast.ResizeAudio{
		Output:      soundMeterAudioHandler,
		SampleCount: 1024,
	})

	httpServer := &broadcast.HttpServer{SoundMeterAudioHandler: soundMeterAudioHandler}

	config.Apply(alsaInput, httpStreamOutput, httpServer)

	err := alsaInput.Init()
	command.checkError(err)

	err = httpStreamOutput.Init()
	command.checkError(err)

	err = httpServer.Init()
	command.checkError(err)

	go httpStreamOutput.Run()

	alsaInput.Run()
}

type HttpSourceConfig struct {
	broadcast.CommandConfig

	Alsa   broadcast.AlsaInputConfig
	Stream broadcast.BufferedHttpStreamOutputConfig
}

func (config *HttpSourceConfig) Flags(flags *flag.FlagSet) {
	config.BaseFlags(flags)

	config.Alsa.Flags(flags, "alsa")
	config.Stream.Flags(flags, "stream")
}

func (config *HttpSourceConfig) Apply(alsaInput *broadcast.AlsaInput, httpStreamOutput *broadcast.BufferedHttpStreamOutput, httpServer *broadcast.HttpServer) {
	config.BaseApply(httpServer)

	config.Alsa.Apply(alsaInput)

	httpStreamOutput.SetChannelCount(alsaInput.Channels)
	httpStreamOutput.SetSampleRate(alsaInput.SampleRate)

	config.Stream.Apply(httpStreamOutput)
}
