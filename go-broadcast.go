package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"projects.tryphon.eu/go-broadcast/broadcast"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "http://host:port/mount_point.ogg")
		os.Exit(1)
	}
	url, err := url.Parse(os.Args[1])
	checkError(err)

	client := &http.Client{}
	request, err := http.NewRequest("GET", url.String(), nil)
	checkError(err)

	request.Header.Add("Cache-Control", "no-cache")
	request.Header.Add("User-Agent", "Go Broadcast v0")

	var (
		oggDecoder    broadcast.OggDecoder
		vorbisDecoder broadcast.VorbisDecoder
		alsaSink      broadcast.AlsaSink
	)

	cache := true

	err = alsaSink.Init()
	checkError(err)

	if cache {
		audioChannel := make(chan *broadcast.Audio, 1500)

		go func() {
			time.Sleep(10 * time.Second)
			for {
				audio := <-audioChannel
				alsaSink.AudioOut(audio)
			}
		}()

		vorbisDecoder.SetAudioHandler(broadcast.AudioHandlerFunc(func(audio *broadcast.Audio) {
			audioChannel <- audio
		}))
	} else {
		vorbisDecoder.SetAudioHandler(&alsaSink)
	}

	go func() {
		for {
			fmt.Printf("%v vorbis / alsa sampleCount : %d\n", time.Now(), vorbisDecoder.SampleCount()-alsaSink.SampleCount())
			time.Sleep(1 * time.Second)
		}
	}()

	oggDecoder.SetHandler(&vorbisDecoder)

	for {
		fmt.Println("New HTTP request")
		response, err := client.Do(request)
		if err == nil && response.Status == "200 OK" {
			reader := response.Body

			for oggDecoder.Read(reader) {
			}

			fmt.Println("End of HTTP stream")

			oggDecoder.Reset()
			vorbisDecoder.Reset()
		} else {
			if err != nil {
				fmt.Println("Error ", err.Error())
			} else {
				fmt.Println(response.Status)
			}

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
