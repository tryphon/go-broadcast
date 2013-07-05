package main

import (
	"fmt"
	"os"
	"time"

	"projects.tryphon.eu/go-broadcast/broadcast"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "http://host:port/mount_point.ogg")
		os.Exit(1)
	}

	httpInput := broadcast.HttpInput{Url: os.Args[1]}

	err := httpInput.Init()
	checkError(err)

	alsaSink := broadcast.AlsaSink{}

	err = alsaSink.Init()
	checkError(err)

	cache := true
	if cache {
		audioBuffer := broadcast.NewAudioBuffer()
		audioBuffer.MinSampleCount = 44100 * 5

		fmt.Printf("AudioBuffer MinSampleCount : %d samples\n", audioBuffer.MinSampleCount)

		go func() {
			for {
				// fmt.Printf("%v vorbis / alsa sampleCount : %d\n", time.Now(), httpInput.SampleCount()-alsaSink.SampleCount())
				// fmt.Printf("%v buffer SampleCount : %d\n", time.Now(), audioBuffer.Dump())
				audioBuffer.Dump()
				time.Sleep(1 * time.Second)
			}
		}()

		go func() {
			// time.Sleep(3 * time.Second)

			defer func() {
				if err := recover(); err != nil {
					fmt.Println("Exception occured: ", err)
				}
			}()

			var blankDuration uint32

			for {
				audio := audioBuffer.Read()
				if audio == nil {
					audio = broadcast.NewAudio()
					blankDuration += uint32(audio.SampleCount())
				} else {
					if blankDuration > 0 {
						fmt.Printf("%v Blank duration : %d samples\n", time.Now(), blankDuration)
						blankDuration = 0
					}
				}

				alsaSink.AudioOut(audio)
			}
		}()

		httpInput.SetAudioHandler(audioBuffer)
	} else {
		httpInput.SetAudioHandler(&alsaSink)
	}

	for {
		err := httpInput.Read()

		if err != nil {
			fmt.Println("Error ", err.Error())
			time.Sleep(5 * time.Second)
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
