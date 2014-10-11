package broadcast

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type HttpInput struct {
	Url      string
	Username string
	Password string

	client     http.Client
	reader     io.ReadCloser
	connection net.Conn

	oggDecoder    OggDecoder
	vorbisDecoder VorbisDecoder

	ReadTimeout time.Duration
	WaitOnError time.Duration
}

func (input *HttpInput) dialTimeout(network, addr string) (net.Conn, error) {
	connection, err := net.DialTimeout(network, addr, input.GetReadTimeout())
	if err != nil {
		input.connection = nil
		return nil, err
	}

	input.connection = connection
	input.updateDeadline()

	return connection, nil
}

func (input *HttpInput) checkRedirect(request *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}

	input.setRequestHeaders(request)

	return nil
}

func (input *HttpInput) updateDeadline() {
	if input.connection != nil {
		input.connection.SetDeadline(time.Now().Add(input.GetReadTimeout()))
	}
}

func (input *HttpInput) setRequestHeaders(request *http.Request) {
	request.Header.Add("Cache-Control", "no-cache, must-revalidate")
	request.Header.Add("Pragma", "no-cache")

	request.Header.Add("User-Agent", "Go Broadcast v0")
}

func (input *HttpInput) Init() (err error) {
	transport := http.Transport{
		Dial: input.dialTimeout,
	}

	input.client = http.Client{
		Transport:     &transport,
		CheckRedirect: input.checkRedirect,
	}

	input.oggDecoder.SetHandler(&input.vorbisDecoder)

	return nil
}

func (input *HttpInput) Read() (err error) {
	if input.reader == nil {
		parsedUrl, err := url.Parse(input.Url)
		if err != nil {
			return err
		}

		request, err := http.NewRequest("GET", parsedUrl.String(), nil)
		if err != nil {
			return err
		}

		input.setRequestHeaders(request)

		if input.Username != "" || input.Password != "" {
			Log.Debugf("Use basic auth : %s/[PASSWORD]", input.Username)
			request.SetBasicAuth(input.Username, input.Password)
		}

		Log.Printf("New HTTP request (%s)", parsedUrl.String())
		response, err := input.client.Do(request)

		if err == nil {
			Log.Debugf("HTTP Response : %s", response.Status)
			if response.Status != "200 OK" {
				err = errors.New("HTTP Error")
			}
		}

		if err != nil {
			if response != nil {
				response.Body.Close()
			}
			return err
		}

		input.reader = NewMetricsReadCloser(response.Body, "http.input.Traffic")
	}

	if input.oggDecoder.Read(input.reader) {
		input.updateDeadline()
	} else {
		Log.Printf("End of HTTP stream")
		input.Reset()
	}

	return nil
}

func (input *HttpInput) Reset() {
	if input.reader != nil {
		input.reader.Close()
		input.reader = nil
	}

	input.oggDecoder.Reset()
	input.vorbisDecoder.Reset()
}

func (input *HttpInput) SetAudioHandler(audioHandler AudioHandler) {
	input.vorbisDecoder.audioHandler = audioHandler
}

func (input *HttpInput) Run() {
	for {
		err := input.Read()

		if err != nil {
			Log.Printf("HTTP Error : %s", err.Error())
			time.Sleep(input.GetWaitOnError())
		}
	}
}

func (input *HttpInput) GetWaitOnError() time.Duration {
	if input.WaitOnError == 0 {
		input.WaitOnError = 5 * time.Second
	}
	return input.WaitOnError
}

func (input *HttpInput) GetReadTimeout() time.Duration {
	if input.ReadTimeout == 0 {
		input.ReadTimeout = 10 * time.Second
	}
	return input.ReadTimeout
}
