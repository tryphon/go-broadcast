package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	"projects.tryphon.eu/go-broadcast/broadcast"

	"code.google.com/p/go.net/websocket"
	metrics "github.com/tryphon/go-metrics"
	"net/http"
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
	config := broadcast.UDPClientConfig{}

	flags := flag.NewFlagSet("udpclient", flag.ExitOnError)
	config.Flags(flags)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s udpclient [options]\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	if config.Udp.Target == "" {
		flag.Usage()
		os.Exit(1)
	}

	config.Alsa.SampleRate = 48000
	broadcast.Log.Printf("Config: %v", config)

	alsaInput := &broadcast.AlsaInput{}
	udpOutput := &broadcast.UDPOutput{}

	soundMeterAudioHandler := &broadcast.SoundMeterAudioHandler{
		Output: udpOutput,
	}
	alsaInput.SetAudioHandler(soundMeterAudioHandler)

	httpServer := &broadcast.HttpServer{SoundMeterAudioHandler: soundMeterAudioHandler}

	config.Apply(alsaInput, udpOutput, httpServer)

	err := alsaInput.Init()
	checkError(err)

	err = udpOutput.Init()
	checkError(err)

	err = httpServer.Init()
	checkError(err)

	go alsaInput.Run()

	for {
		time.Sleep(2 * time.Second)
	}
}

func udpServer(arguments []string) {
	config := broadcast.UDPServerConfig{}

	flags := flag.NewFlagSet("udpserver", flag.ExitOnError)
	config.Flags(flags)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] udpserver [bind]:<port>\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	config.Alsa.SampleRate = 48000
	broadcast.Log.Printf("Config: %v", config)

	alsaOutput := &broadcast.AlsaOutput{}
	udpInput := &broadcast.UDPInput{}

	soundMeterAudioHandler := &broadcast.SoundMeterAudioHandler{
		Output: alsaOutput,
	}
	httpServer := &broadcast.HttpServer{SoundMeterAudioHandler: soundMeterAudioHandler}

	config.Apply(alsaOutput, udpInput, httpServer)

	err := alsaOutput.Init()
	checkError(err)

	err = udpInput.Init()
	checkError(err)

	err = httpServer.Init()
	checkError(err)

	channel := make(chan *broadcast.Audio, 100)
	audioHandler := broadcast.AudioHandlerFunc(func(audio *broadcast.Audio) {
		channel <- audio
	})
	udpInput.SetAudioHandler(audioHandler)

	go udpInput.Run()

	for {
		select {
		case audio := <-channel:
			soundMeterAudioHandler.AudioOut(audio)
		default:
			soundMeterAudioHandler.AudioOut(broadcast.NewAudio(1024, 2))
		}
	}
}

func httpClient(arguments []string) {
	var sampleRate int = 44100

	flags := flag.NewFlagSet("httpclient", flag.ExitOnError)

	var lowAdjustLimit, lowAdjustThreshold, lowRefillMin, highAdjustLimit, highAdjustThreshold, highUnfillMax, highUnfill float64
	var statusLoop, httpReadTimeout, httpWaitOnError time.Duration
	var httpUsername, httpPassword string
	var alsaDevice /*, alsaSampleFormat */ string
	var alsaChannels int

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

	flags.StringVar(&alsaDevice, "alsa-device", "default", "The alsa device used to play sound")
	flags.IntVar(&alsaChannels, "alsa-channels", 2, "The channel count used with alsa device")
	// flags.StringVar(&alsaSampleFormat, "alsa-sample-format", "auto", "The sample format used to record sound (s16le, s32le, s32be)")

	var fixedRateTolerance float64
	flags.Float64Var(&fixedRateTolerance, "sample-rate-tolerance", 1, "Tolerance on sample format (0 is no tolerance, 1 is no fixed sample format)")

	var cpuProfile, memProfile string
	flags.StringVar(&cpuProfile, "cpuprofile", "", "Write cpu profile to file")
	flags.StringVar(&memProfile, "memprofile", "", "Write memory profile to this file")

	flags.BoolVar(&broadcast.Log.Debug, "debug", false, "Enable debug messages")
	flags.BoolVar(&broadcast.Log.Syslog, "syslog", false, "Redirect messages to syslog")

	var httpServer string
	flags.StringVar(&httpServer, "http-server", "", "Http server binding")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] httpclient <url>\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	if flags.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {
				broadcast.Log.Debugf("Receive interrupt signal: %v", sig)
				pprof.StopCPUProfile()
				os.Exit(0)
			}
		}()
	}

	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatal(err)
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {
				broadcast.Log.Debugf("Receive interrupt signal: %v", sig)
				pprof.WriteHeapProfile(f)
				f.Close()
				os.Exit(0)
			}
		}()
	}

	httpInput := broadcast.HttpInput{Url: flags.Arg(0), ReadTimeout: httpReadTimeout, WaitOnError: httpWaitOnError, Username: httpUsername, Password: httpPassword}
	err := httpInput.Init()
	checkError(err)

	alsaOutput := broadcast.AlsaOutput{Device: alsaDevice, SampleRate: sampleRate, Channels: alsaChannels}
	var outputHandler broadcast.AudioHandler = &alsaOutput

	soundMeterAudioHandler := &broadcast.SoundMeterAudioHandler{Output: outputHandler}
	outputHandler = soundMeterAudioHandler

	if fixedRateTolerance > 0 && fixedRateTolerance < 1 {
		fixedRateOutput := broadcast.FixedRateAudioHandler{
			Output:     &alsaOutput,
			SampleRate: uint(sampleRate),
			Tolerance:  fixedRateTolerance,
		}

		broadcast.Log.Debugf("Fixed Rate Output with %d%% tolerance", int(fixedRateTolerance*100))
		outputHandler = &fixedRateOutput
	}

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
			Buffer:         lowAdjustBuffer,
			MinSampleCount: lowRefillMinSampleCount,
		},
		LimitSampleCount:     highAdjustLimitSampleCount,
		ThresholdSampleCount: highAdjustThresholdSampleCount,
	}

	audioBuffer := &broadcast.MutexAudioBuffer{
		Buffer: &broadcast.UnfillAudioBuffer{
			Buffer:            highAdjustBuffer,
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

	if httpServer != "" {
		metricsJSON := func(response http.ResponseWriter, request *http.Request) {
			response.Header().Set("Content-Type", "application/json")
			response.Header().Set("Access-Control-Allow-Origin", "*")

			jsonBytes, _ := json.Marshal(metrics.DefaultRegistry)
			response.Write(jsonBytes)
		}
		http.HandleFunc("/metrics", metricsJSON)

		soundMeterJSON := func(response http.ResponseWriter, request *http.Request) {
			response.Header().Set("Content-Type", "application/json")
			response.Header().Set("Access-Control-Allow-Origin", "*")

			jsonBytes, _ := json.Marshal(soundMeterAudioHandler)
			response.Write(jsonBytes)
		}
		http.HandleFunc("/soundmeter.json", soundMeterJSON)

		soundMeterWebSocket := func(webSocket *websocket.Conn) {
			broadcast.Log.Debugf("New SoundMeter websocket connection")

			receiver := soundMeterAudioHandler.NewReceiver()
			defer receiver.Close()

			go func() {
				for metrics := range receiver.Channel {
					jsonBytes, _ := json.Marshal(metrics)
					err := websocket.Message.Send(webSocket, string(jsonBytes))
					if err != nil {
						broadcast.Log.Debugf("Can't send websocket message: %v", err)
						break
					}
				}
			}()

			for {
				var message string
				err := websocket.Message.Receive(webSocket, &message)
				if err != nil {
					break
				}
			}

			broadcast.Log.Debugf("Close SoundMeter websocket connection")
		}
		http.Handle("/soundmeter.ws", websocket.Handler(soundMeterWebSocket))

		go http.ListenAndServe(httpServer, nil)
	}

	if statusLoop > 0 {
		go metrics.Log(metrics.DefaultRegistry, statusLoop, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
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

		outputHandler.AudioOut(audio)
	}
}

func loopback(arguments []string) {
	flags := flag.NewFlagSet("httpclient", flag.ExitOnError)

	var bufferDuration time.Duration
	var inputDevice, outputDevice string
	var inputSampleFormat, outputSampleFormat string
	var sampleRate int

	flags.BoolVar(&broadcast.Log.Debug, "debug", false, "Enable debug messages")
	flags.BoolVar(&broadcast.Log.Syslog, "syslog", false, "Redirect messages to syslog")

	flags.StringVar(&inputDevice, "input-device", "default", "The alsa device used to capture sound")
	flags.StringVar(&inputSampleFormat, "input-sample-format", "auto", "The sample format used to capture sound (s16le, s32le, s32be)")

	flags.StringVar(&outputDevice, "output-device", "default", "The alsa device used to play sound")
	flags.StringVar(&outputSampleFormat, "output-sample-format", "auto", "The sample format used to play sound (s16le, s32le, s32be)")

	flags.DurationVar(&bufferDuration, "buffer-duration", 250*time.Millisecond, "Buffer duration")

	flags.IntVar(&sampleRate, "sample-rate", 44100, "Sample rate")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] loopback\n", os.Args[0])
		flags.PrintDefaults()
	}

	flags.Parse(arguments)

	bufferSampleCount := int(float64(sampleRate) * bufferDuration.Seconds())
	broadcast.Log.Debugf("Alsa bufferSampleCount: %d", bufferSampleCount)

	alsaInput := broadcast.AlsaInput{Device: inputDevice, SampleRate: sampleRate, BufferSampleCount: bufferSampleCount, SampleFormat: broadcast.ParseSampleFormat(inputSampleFormat)}
	err := alsaInput.Init()
	checkError(err)

	alsaOutput := broadcast.AlsaOutput{Device: outputDevice, SampleRate: sampleRate, SampleFormat: broadcast.ParseSampleFormat(outputSampleFormat)}
	err = alsaOutput.Init()
	checkError(err)

	channel := make(chan *broadcast.Audio, 100)
	audioHandler := broadcast.AudioHandlerFunc(func(audio *broadcast.Audio) {
		channel <- audio
	})
	alsaInput.SetAudioHandler(audioHandler)

	go alsaInput.Run()

	time.Sleep(bufferDuration * 2)

	for {
		audio := <-channel
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
