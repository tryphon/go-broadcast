package broadcast

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpInput struct {
	Url      string
	Username string
	Password string

	client     http.Client
	reader     io.ReadCloser
	connection net.Conn

	streamDecoder StreamDecoder
	audioHandler  AudioHandler

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

	return nil
}

func (input *HttpInput) Read() (err error) {
	if input.reader == nil {
		parsedUrl, err := url.Parse(input.Url)
		if err != nil {
			return err
		}

		if parsedUrl.User != nil {
			if input.Username == "" && parsedUrl.User.Username() != "" {
				input.Username = parsedUrl.User.Username()
			}

			if input.Password == "" {
				password, ok := parsedUrl.User.Password()
				if ok {
					input.Password = password
				}
			}
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

		contentType := response.Header.Get("Content-Type")
		Log.Printf("Stream content type: %s", contentType)

		audioEncoding := FindEncodingByContentType(contentType)
		Log.Printf("Stream audio encoding: %s", audioEncoding)

		input.streamDecoder = NewStreamDecoder(audioEncoding)

		if input.streamDecoder == nil {
			return fmt.Errorf("Unsupported audio type: content-type=%s encoding=%s", contentType, audioEncoding)
		}

		input.streamDecoder.SetAudioHandler(input.audioHandler)
		input.streamDecoder.Init()

		input.reader = NewMetricsReadCloser(response.Body, "http.input.Traffic")
	}

	if err = input.streamDecoder.Read(input.reader); err == nil {
		input.updateDeadline()
	} else {
		Log.Printf("End of HTTP stream : %v", err)
		input.Reset()
	}

	return nil
}

func (input *HttpInput) Reset() {
	if input.reader != nil {
		input.reader.Close()
		input.reader = nil
	}

	input.streamDecoder.Reset()
}

func (input *HttpInput) SetAudioHandler(audioHandler AudioHandler) {
	input.audioHandler = audioHandler
	if input.streamDecoder != nil {
		input.streamDecoder.SetAudioHandler(audioHandler)
	}
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

func (input *HttpInput) Setup(config *HttpStreamInputConfig) {
	input.Url = config.Url

	input.ReadTimeout = config.ReadTimeout
	input.WaitOnError = config.WaitOnError
}

type HttpStreamInputConfig struct {
	Url string

	ReadTimeout time.Duration
	WaitOnError time.Duration
}

func (config *HttpStreamInputConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.StringVar(&config.Url, strings.Join([]string{prefix, "url"}, "-"), "", "URL of played stream")

	flags.DurationVar(&config.ReadTimeout, strings.Join([]string{prefix, "readtimeout"}, "-"), 10*time.Second, "Timeout on read operations")
	flags.DurationVar(&config.WaitOnError, strings.Join([]string{prefix, "waitonerror"}, "-"), 5*time.Second, "Delay after a network error")
}
