package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"projects.tryphon.eu/go-broadcast/broadcast"
)

func main() {
	var command string
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "httpclient":
		httpClient(os.Args[2:])
	case "udpclient":
		udpClient(os.Args[2:])
	case "udpserver":
		udpServer(os.Args[2:])
	case "backup":
		backup(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Usage: %s [options] httpclient|backup|udpclient|udpserver <url>\n", os.Args[0])
		os.Exit(1)
	}
}

func backup(arguments []string) {
	flags := flag.NewFlagSet("backup", flag.ExitOnError)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] backup <root-directory>\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	if flags.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	var rootDirectory string = flags.Arg(0)

	alsaInput := broadcast.AlsaInput{}

	err := alsaInput.Init()
	checkError(err)

	timedFileOutput := &broadcast.TimedFileOutput{RootDirectory: rootDirectory}
	// err = timedFileOutput.Init()
	// checkError(err)

	audioHandler := &broadcast.SoundMeterAudioHandler{
		Output: timedFileOutput,
	}
	alsaInput.SetAudioHandler(audioHandler)

	go alsaInput.Run()

	for {
		time.Sleep(2 * time.Second)
	}
}

func udpClient(arguments []string) {
	alsaInput := broadcast.AlsaInput{}

	err := alsaInput.Init()
	checkError(err)

	udpOutput := &broadcast.UDPOutput{Target: "localhost:7890"}
	err = udpOutput.Init()
	checkError(err)

	audioHandler := &broadcast.SoundMeterAudioHandler{
		Output: &broadcast.ResizeAudio{
			Output:       udpOutput,
			SampleCount:  150,
			ChannelCount: 2,
		},
	}
	alsaInput.SetAudioHandler(audioHandler)

	go alsaInput.Run()

	for {
		time.Sleep(2 * time.Second)
	}
}

func udpServer(arguments []string) {
	alsaOutput := &broadcast.AlsaOutput{}

	err := alsaOutput.Init()
	checkError(err)

	udpInput := &broadcast.UDPInput{Bind: ":7890"}
	err = udpInput.Init()
	checkError(err)

	audioHandler := &broadcast.ResizeAudio{
		Output: &broadcast.SoundMeterAudioHandler{
			Output: alsaOutput,
		},
		SampleCount:  1024,
		ChannelCount: 2,
	}

	udpInput.SetAudioHandler(audioHandler)

	go udpInput.Run()

	for {
		time.Sleep(2 * time.Second)
	}
}

func httpClient(arguments []string) {
	var sampleRate uint = 44100

	flags := flag.NewFlagSet("httpclient", flag.ExitOnError)

	var lowAdjustLimit, lowAdjustThreshold, lowRefillMin, highAdjustLimit, highAdjustThreshold, highUnfillMax, highUnfill float64
	var statusLoop, httpReadTimeout, httpWaitOnError time.Duration
	var httpUsername, httpPassword string

	flags.Float64Var(&lowAdjustLimit, "low-adjust-limit", 0, "Limit of low adjust buffer (in seconds)")
	flags.Float64Var(&lowAdjustThreshold, "low-adjust-threshold", 3, "Limit of low adjust buffer (in seconds)")
	flags.Float64Var(&lowRefillMin, "low-refill", 3, "Duration to refill when buffer is empty (in seconds)")

	flags.Float64Var(&highAdjustThreshold, "high-adjust-threshold", 7, "Limit of high adjust buffer (in seconds)")
	flags.Float64Var(&highAdjustLimit, "high-adjust-limit", 10, "Limit of high adjust buffer (in seconds)")
	flags.Float64Var(&highUnfillMax, "high-max", 10, "Max duration of buffer (in seconds)")
	flags.Float64Var(&highUnfill, "high-unfill", 3, "Duration to unfill when buffer is full (in seconds)")

	flags.DurationVar(&statusLoop, "status-loop", 0, "Duration between two status dump (0 to disable)")

	flags.DurationVar(&httpReadTimeout, "http-read-timeout", 10*time.Second, "Read timeout before creating a new http connection")
	flags.DurationVar(&httpWaitOnError, "http-wait-on-error", 5*time.Second, "Delay after http error")

	flags.StringVar(&httpUsername, "http-username", "", "Username used for http authentification")
	flags.StringVar(&httpPassword, "http-password", "", "Password used for http authentification")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] httpclient <url>\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	if flags.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	httpInput := broadcast.HttpInput{Url: flags.Arg(0), ReadTimeout: httpReadTimeout, WaitOnError: httpWaitOnError, Username: httpUsername, Password: httpPassword}
	err := httpInput.Init()
	checkError(err)

	alsaOutput := broadcast.AlsaOutput{}

	err = alsaOutput.Init()
	checkError(err)

	lowAdjustLimitSampleCount := uint32(lowAdjustLimit * float64(sampleRate))
	lowAdjustThresholdSampleCount := uint32(lowAdjustThreshold * float64(sampleRate))
	lowRefillMinSampleCount := uint32(lowRefillMin * float64(sampleRate))
	highAdjustLimitSampleCount := uint32(highAdjustLimit * float64(sampleRate))
	highAdjustThresholdSampleCount := uint32(highAdjustThreshold * float64(sampleRate))
	highUnfillMaxSampleCount := uint32(highUnfillMax * float64(sampleRate))
	highUnfillSampleCount := uint32(highUnfill * float64(sampleRate))

	audioBuffer := &broadcast.MutexAudioBuffer{
		Buffer: &broadcast.UnfillAudioBuffer{
			Buffer: &broadcast.AdjustAudioBuffer{
				Buffer: &broadcast.RefillAudioBuffer{
					Buffer: &broadcast.AdjustAudioBuffer{
						Buffer:               &broadcast.MemoryAudioBuffer{},
						LimitSampleCount:     lowAdjustLimitSampleCount,
						ThresholdSampleCount: lowAdjustThresholdSampleCount,
					},
					MinSampleCount: lowRefillMinSampleCount,
				},
				LimitSampleCount:     highAdjustLimitSampleCount,
				ThresholdSampleCount: highAdjustThresholdSampleCount,
			},
			MaxSampleCount:    highUnfillMaxSampleCount,
			UnfillSampleCount: highUnfillSampleCount,
		},
	}

	broadcast.Log.Debugf("AudioBuffer low-adjust-limit %d, low-adjust-threshold %d, low-refill %d", lowAdjustLimitSampleCount, lowAdjustThresholdSampleCount, lowRefillMinSampleCount)
	broadcast.Log.Debugf("AudioBuffer high-adjust-threshold %d, high-adjust-limit %d, high-max %d, high-unfill %d", highAdjustLimitSampleCount, highAdjustThresholdSampleCount, highUnfillMaxSampleCount, highUnfillSampleCount)

	httpInput.SetAudioHandler(
		&broadcast.ResizeAudio{
			Output:       audioBuffer,
			SampleCount:  1024,
			ChannelCount: 2,
		},
	)

	if statusLoop > 0 {
		go func() {
			for {
				time.Sleep(statusLoop)
				broadcast.Log.Debugf("SampleCount: %d", audioBuffer.SampleCount())
			}
		}()
	}

	go httpInput.Run()

	var blankDuration uint32

	for {
		audio := audioBuffer.Read()
		if audio == nil {
			audio = broadcast.NewAudio(1024, 2)
			blankDuration += uint32(audio.SampleCount())
		} else {
			if blankDuration > 0 {
				broadcast.Log.Printf("Blank duration : %d samples", blankDuration)
				blankDuration = 0
			}
		}

		alsaOutput.AudioOut(audio)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		broadcast.Log.Printf("Fatal error : %s", err.Error())
		os.Exit(1)
	}
}
