package broadcast

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type HttpInput struct {
	Url        string
	request    *http.Request
	client     http.Client
	response   *http.Response
	connection net.Conn

	oggDecoder    OggDecoder
	vorbisDecoder VorbisDecoder
}

func (input *HttpInput) dialTimeout(network, addr string) (net.Conn, error) {
	connection, err := net.DialTimeout(network, addr, 10*time.Second)
	if err != nil {
		input.connection = nil
		return nil, err
	}

	input.connection = connection

	return connection, nil
}

func (input *HttpInput) checkRedirect(request *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 5 redirects")
	}

	input.setRequestHeaders(request)

	return nil
}

func (input *HttpInput) setRequestHeaders(request *http.Request) {
	request.Header.Add("Cache-Control", "no-cache, must-revalidate")
	request.Header.Add("Pragma", "no-cache")

	request.Header.Add("User-Agent", "Go Broadcast v0")
}

func (input *HttpInput) Init() (err error) {
	parsedUrl, err := url.Parse(input.Url)
	if err != nil {
		return err
	}

	transport := http.Transport{
		Dial: input.dialTimeout,
	}

	input.client = http.Client{
		Transport:     &transport,
		CheckRedirect: input.checkRedirect,
	}

	input.oggDecoder.SetHandler(&input.vorbisDecoder)

	input.request, err = http.NewRequest("GET", parsedUrl.String(), nil)

	if err != nil {
		return err
	}

	input.setRequestHeaders(input.request)

	return nil
}

func (input *HttpInput) Read() (err error) {
	if input.response == nil {
		fmt.Println("New HTTP request")
		response, err := input.client.Do(input.request)

		if err == nil {
			fmt.Println("HTTP Response : ", response.Status)
			if response.Status != "200 OK" {
				err = errors.New("HTTP Error")
			}
		}

		if err != nil {
			if input.response != nil {
				input.response.Body.Close()
			}
			input.response = nil
			return err
		}

		input.response = response
	}

	if input.oggDecoder.Read(input.response.Body) {
		deadline := time.Now().Add(15 * time.Second)
		if input.connection != nil {
			input.connection.SetDeadline(deadline)
		}
	} else {
		fmt.Println("End of HTTP stream")
		input.Reset()
	}

	return nil
}

func (input *HttpInput) Reset() {
	input.response.Body.Close()
	input.response = nil

	input.oggDecoder.Reset()
	input.vorbisDecoder.Reset()
}

func (input *HttpInput) SetAudioHandler(audioHandler AudioHandler) {
	input.vorbisDecoder.audioHandler = audioHandler
}

func (input *HttpInput) SampleCount() int64 {
	return input.vorbisDecoder.sampleCount
}

func (input *HttpInput) Run() {
	for {
		err := input.Read()

		if err != nil {
			fmt.Println("Error ", err.Error())
			time.Sleep(5 * time.Second)
		}
	}
}
