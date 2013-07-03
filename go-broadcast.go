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
		audioChannel := make(chan *broadcast.Audio, 1500)

		go func() {
			time.Sleep(10 * time.Second)
			for {
				audio := <-audioChannel
				alsaSink.AudioOut(audio)
			}
		}()

		httpInput.SetAudioHandler(broadcast.AudioHandlerFunc(func(audio *broadcast.Audio) {
			audioChannel <- audio
		}))
	} else {
		httpInput.SetAudioHandler(&alsaSink)
	}

	go func() {
		for {
			fmt.Printf("%v vorbis / alsa sampleCount : %d\n", time.Now(), httpInput.SampleCount()-alsaSink.SampleCount())
			time.Sleep(1 * time.Second)
		}
	}()

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
