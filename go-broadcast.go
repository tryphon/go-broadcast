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
	case "loopback":
		loopback(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Usage: %s [options] httpclient|backup|udpclient|udpserver <url>\n", os.Args[0])
		os.Exit(1)
	}
}

func backup(arguments []string) {
	flags := flag.NewFlagSet("backup", flag.ExitOnError)

	var fileDuration, bufferDuration time.Duration
	var sampleRate int
	var alsaDevice, alsaSampleFormat string

	flags.StringVar(&alsaDevice, "alsa-device", "default", "The alsa device used to record sound")
	flags.StringVar(&alsaSampleFormat, "alsa-sample-format", "auto", "The sample format used to record sound (s16le, s32le, s32be)")
	flags.DurationVar(&fileDuration, "file-duration", 5*time.Minute, "Change file duration")
	flags.DurationVar(&bufferDuration, "buffer-duration", 250*time.Millisecond, "Buffer duration")
	flags.IntVar(&sampleRate, "sample-rate", 44100, "Sample rate")

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

	bufferSampleCount := int(float64(sampleRate) * bufferDuration.Seconds())
	broadcast.Log.Debugf("Alsa bufferSampleCount: %d", bufferSampleCount)

	alsaInput := broadcast.AlsaInput{Device: alsaDevice, BufferSampleCount: bufferSampleCount, SampleRate: sampleRate, SampleFormat: broadcast.ParseSampleFormat(alsaSampleFormat)}
	err := alsaInput.Init()
	checkError(err)

	timedFileOutput := &broadcast.TimedFileOutput{RootDirectory: rootDirectory}
	timedFileOutput.SetFileDuration(fileDuration)
	timedFileOutput.SetSampleRate(sampleRate)

	channel := make(chan *broadcast.Audio, 100)
	audioHandler := broadcast.AudioHandlerFunc(func(audio *broadcast.Audio) {
		channel <- audio
	})
	alsaInput.SetAudioHandler(audioHandler)

	go alsaInput.Run()

	for {
		audio := <-channel
		timedFileOutput.AudioOut(audio)
	}
}

func udpClient(arguments []string) {
	flags := flag.NewFlagSet("udpclient", flag.ExitOnError)

	var alsaDevice string
	flags.StringVar(&alsaDevice, "alsa-device", "default", "The alsa device used to capture sound")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] udpclient <host>:<port>\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	if flags.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	alsaInput := broadcast.AlsaInput{Device: alsaDevice, SampleRate: 48000}

	err := alsaInput.Init()
	checkError(err)

	udpOutput := &broadcast.UDPOutput{Target: flags.Arg(0)}
	err = udpOutput.Init()
	checkError(err)

	audioHandler := &broadcast.SoundMeterAudioHandler{
		Output: udpOutput,
	}
	alsaInput.SetAudioHandler(audioHandler)

	go alsaInput.Run()

	for {
		time.Sleep(2 * time.Second)
	}
}

func udpServer(arguments []string) {
	flags := flag.NewFlagSet("udpserver", flag.ExitOnError)

	var alsaDevice string
	flags.StringVar(&alsaDevice, "alsa-device", "default", "The alsa device used to play sound")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] udpserver [bind]:<port>\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	if flags.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	alsaOutput := &broadcast.AlsaOutput{Device: alsaDevice, SampleRate: 48000}

	err := alsaOutput.Init()
	checkError(err)

	udpInput := &broadcast.UDPInput{Bind: flags.Arg(0)}
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

  lowAdjustBuffer := &broadcast.AdjustAudioBuffer{
		Buffer:               &broadcast.MemoryAudioBuffer{},
		LimitSampleCount:     lowAdjustLimitSampleCount,
		ThresholdSampleCount: lowAdjustThresholdSampleCount,
	}

	highAdjustBuffer := &broadcast.AdjustAudioBuffer{
		Buffer: &broadcast.RefillAudioBuffer{
			Buffer: lowAdjustBuffer,
			MinSampleCount: lowRefillMinSampleCount,
		},
		LimitSampleCount:     highAdjustLimitSampleCount,
		ThresholdSampleCount: highAdjustThresholdSampleCount,
	}

	audioBuffer := &broadcast.MutexAudioBuffer{
		Buffer: &broadcast.UnfillAudioBuffer{
			Buffer: highAdjustBuffer,
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
			output := "SampleCount: %d, Low Adjustment: %d, High Adjustment: %d, Alsa SampleCount: %d, Alsa delay: %d, Vorbis: %d"
			for {
				time.Sleep(statusLoop)
				broadcast.Log.Debugf(output, audioBuffer.SampleCount(), lowAdjustBuffer.AdjustmentSampleCount(), highAdjustBuffer.AdjustmentSampleCount(), alsaOutput.SampleCount(), alsaOutput.Delay(), httpInput.SampleCount())
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

func loopback(arguments []string) {
	flags := flag.NewFlagSet("httpclient", flag.ExitOnError)

	var bufferDuration time.Duration
	var inputDevice, outputDevice string
	// var inputSampleFormat, outputSampleFormat string
	var sampleRate int

	flags.StringVar(&inputDevice, "input-device", "default", "The alsa device used to capture sound")
	// flags.StringVar(&inputSampleFormat, "input-sample-format", "auto", "The sample format used to capture sound (s16le, s32le, s32be)")

	flags.StringVar(&outputDevice, "output-device", "default", "The alsa device used to play sound")
	// flags.StringVar(&outputSampleFormat, "output-sample-format", "auto", "The sample format used to play sound (s16le, s32le, s32be)")

	flags.DurationVar(&bufferDuration, "buffer-duration", 250*time.Millisecond, "Buffer duration")

	flags.IntVar(&sampleRate, "sample-rate", 44100, "Sample rate")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] loopback\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	bufferSampleCount := int(float64(sampleRate) * bufferDuration.Seconds())
	broadcast.Log.Debugf("Alsa bufferSampleCount: %d", bufferSampleCount)

	alsaInput := broadcast.AlsaInput{Device: inputDevice, SampleRate: sampleRate, BufferSampleCount: bufferSampleCount, SampleFormat: broadcast.ParseSampleFormat("s16le")}
	err := alsaInput.Init()
	checkError(err)
	
	alsaOutput := broadcast.AlsaOutput{Device: outputDevice, SampleRate: sampleRate}
	err = alsaOutput.Init()
	checkError(err)

	channel := make(chan *broadcast.Audio, 100)
	audioHandler := broadcast.AudioHandlerFunc(func(audio *broadcast.Audio) {
		// fmt.Fprintf(os.Stderr, "read %d\n", audio.SampleCount())
		channel <- audio
	})
	alsaInput.SetAudioHandler(audioHandler)

	go alsaInput.Run()

	time.Sleep(bufferDuration * 2)

	for {
		audio := <-channel
		// fmt.Fprintf(os.Stderr, "write %d\n", audio.SampleCount())
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
