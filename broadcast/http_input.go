package broadcast

import (
	"net/http"
	"net/url"
	"net"
	"time"
	"fmt"
)

type HttpInput struct {
	Url string
	request *http.Request
	client http.Client
	response *http.Response
	connection net.Conn

	oggDecoder    OggDecoder
	vorbisDecoder VorbisDecoder
}

func (input *HttpInput) dialTimeout(network, addr string) (net.Conn, error) {
	connection, err := net.DialTimeout(network, addr, 10 * time.Second)
	if err != nil {
		input.connection = nil
		return nil, err
	}

	input.connection = connection

	// TODO
	// In our case, need to be updated after each read
	//
	// deadline := time.Now().Add(10 * time.Second)
	// c.SetDeadline(deadline)

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
		if err == nil && response.Status == "200 OK" {
			input.response = response
		} else {
			return err
		}
	}

	if ! input.oggDecoder.Read(input.response.Body) {
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
