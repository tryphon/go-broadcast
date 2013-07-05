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

func (input *HttpInput) Init() (err error) {
	parsedUrl, err := url.Parse(input.Url)
	if err != nil {
		return err
	}

	transport := http.Transport{
		Dial: input.dialTimeout,
	}

	input.client = http.Client{
		Transport: &transport,
	}

	input.oggDecoder.SetHandler(&input.vorbisDecoder)

	input.request, err = http.NewRequest("GET", parsedUrl.String(), nil)
	return err
}

func (input *HttpInput) Read() (err error) {
	if input.response == nil {
		fmt.Println("New HTTP request")
		response, err := input.client.Do(input.request)

		if response.Status != "200 OK" {
			err = errors.New("HTTP Error")
		}

		if err == nil {
			input.response = response
		} else {
			return err
		}
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
