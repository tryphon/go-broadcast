package command

import (
	// "encoding/json"
	"flag"
	"fmt"
	"os"
	"projects.tryphon.eu/go-broadcast/broadcast"
)

type Play struct {
	Base

	alsaOutput      broadcast.AlsaOutput
	httpStreamInput *broadcast.BufferedHttpStreamInput
	httpServer      broadcast.HttpServer
	processing      broadcast.Processing

	config *PlayConfig
}

func (command *Play) Main(arguments []string) {
	config := &PlayConfig{}

	flags := flag.NewFlagSet("play", flag.ExitOnError)
	config.Flags(flags)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s play [options]\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	if config.IsEmpty() {
		flag.Usage()
		os.Exit(1)
	}

	command.Setup(config)
	command.checkError(command.Init())

	command.Run()
}

func (command *Play) Init() error {
	soundMeterAudioHandler := &broadcast.SoundMeterAudioHandler{
		Output: &command.alsaOutput,
	}

	command.processing.SetAudioHandler(soundMeterAudioHandler)
	command.httpServer.SoundMeterAudioHandler = soundMeterAudioHandler

	// if fixedRateTolerance > 0 && fixedRateTolerance < 1 {
	// 	fixedRateOutput := broadcast.FixedRateAudioHandler{
	// 		Output:     &alsaOutput,
	// 		SampleRate: uint(sampleRate),
	// 		Tolerance:  fixedRateTolerance,
	// 	}

	// 	broadcast.Log.Debugf("Fixed Rate Output with %d%% tolerance", int(fixedRateTolerance*100))
	// 	outputHandler = &fixedRateOutput
	// }

	err := command.httpStreamInput.Init()
	if err != nil {
		return err
	}

	err = command.alsaOutput.Init()
	if err != nil {
		return err
	}

	err = command.httpServer.Init()
	if err != nil {
		return err
	}

	return nil
}

func (command *Play) Setup(config *PlayConfig) {
	if command.httpStreamInput == nil {
		command.httpStreamInput = broadcast.NewBufferedHttpStreamInput()
	}

	config.BaseApply(&command.httpServer)

	config.Alsa.Apply(&command.alsaOutput)

	command.httpStreamInput.SetSampleRate(command.alsaOutput.SampleRate)
	command.httpStreamInput.SetChannelCount(command.alsaOutput.Channels)

	config.Http.Apply(command.httpStreamInput)

	command.config = config
}

func (command *Play) Run() {
	go command.httpStreamInput.Run()

	var blankDuration uint32
	for {
		audio := command.httpStreamInput.Read()
		if audio == nil {
			audio = broadcast.NewAudio(1024, command.alsaOutput.Channels)
			blankDuration += uint32(audio.SampleCount())
		} else {
			if blankDuration > 0 {
				broadcast.Log.Printf("Blank duration : %d samples", blankDuration)
				blankDuration = 0
			}
		}
		command.processing.AudioOut(audio)
	}
}

type PlayConfig struct {
	broadcast.CommandConfig

	Alsa       broadcast.AlsaOutputConfig
	Http       broadcast.BufferedHttpStreamInputConfig
	Processing broadcast.ProcessingConfig
}

func (config *PlayConfig) IsEmpty() bool {
	return config.Http.Url == ""
}

func (config *PlayConfig) Flags(flags *flag.FlagSet) {
	config.BaseFlags(flags)

	config.Alsa.Flags(flags, "alsa")
	config.Http.Flags(flags, "stream")
	config.Processing.Flags(flags, "processing")
}
