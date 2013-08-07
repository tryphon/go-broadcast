package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"projects.tryphon.eu/go-broadcast/broadcast"
)

func main() {
	var sampleRate uint = 44100

	var lowAdjustLimit, lowAdjustThreshold, lowRefillMin, highAdjustLimit, highAdjustThreshold, highUnfillMax, highUnfill float64
	var statusLoop, httpReadTimeout, httpWaitOnError time.Duration

	flag.Float64Var(&lowAdjustLimit, "low-adjust-limit", 0, "Limit of low adjust buffer (in seconds)")
	flag.Float64Var(&lowAdjustThreshold, "low-adjust-threshold", 3, "Limit of low adjust buffer (in seconds)")
	flag.Float64Var(&lowRefillMin, "low-refill", 3, "Duration to refill when buffer is empty (in seconds)")
	flag.Float64Var(&highAdjustThreshold, "high-adjust-threshold", 7, "Limit of high adjust buffer (in seconds)")
	flag.Float64Var(&highAdjustLimit, "high-adjust-limit", 10, "Limit of high adjust buffer (in seconds)")
	flag.Float64Var(&highUnfillMax, "high-max", 10, "Max duration of buffer (in seconds)")
	flag.Float64Var(&highUnfill, "high-unfill", 3, "Duration to unfill when buffer is full (in seconds)")
	flag.DurationVar(&statusLoop, "status-loop", 0, "Duration between two status dump (0 to disable)")
	flag.DurationVar(&httpReadTimeout, "http-read-timeout", 10*time.Second, "Read timeout before creating a new http connection")
	flag.DurationVar(&httpWaitOnError, "http-wait-on-error", 5*time.Second, "Delay after http error")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <url>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	httpInput := broadcast.HttpInput{Url: flag.Arg(0), ReadTimeout: httpReadTimeout, WaitOnError: httpWaitOnError}
	err := httpInput.Init()
	checkError(err)

	alsaSink := broadcast.AlsaSink{}

	err = alsaSink.Init()
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

	httpInput.SetAudioHandler(audioBuffer)

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
			audio = broadcast.NewAudio()
			blankDuration += uint32(audio.SampleCount())
		} else {
			if blankDuration > 0 {
				broadcast.Log.Printf("Blank duration : %d samples", blankDuration)
				blankDuration = 0
			}
		}

		alsaSink.AudioOut(audio)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		broadcast.Log.Printf("Fatal error : %s", err.Error())
		os.Exit(1)
	}
}
