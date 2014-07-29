package broadcast

import (
	"net/http"
	"net"
	"time"
	"errors"
)

type HttpStreamOutput struct {
	Target string
	// FIXME
	Quality int

	client     http.Client
	connection net.Conn

	// oggEncoder    OggEncoder
	// vorbisEncoder VorbisEncoder
}

func (output *HttpStreamOutput) Init() error {
	transport := http.Transport{
		Dial: output.dialTimeout,
	}

	output.client = http.Client{
		Transport:     &transport,
	}

	request, err := http.NewRequest("SOURCE", output.Target, nil)
	if (err != nil) {
		return err
	}

	request.Header.Add("Content-type", "application/ogg")
	request.Header.Add("User-Agent", "Go Broadcast v0")

	// request.SetBasicAuth("source", password)

	response, err := output.client.Do(request)
	if err != nil {
		return err
	}

	Log.Debugf("HTTP Response : %s", response.Status)
	if response.Status != "200 OK" {
		err = errors.New("HTTP Error")
	}

	defer response.Body.Close()


	// body, err := ioutil.ReadAll(response.Body)
	// fmt.Println(body)

	return nil
}

func (output *HttpStreamOutput) dialTimeout(network, addr string) (net.Conn, error) {
	connection, err := net.DialTimeout(network, addr, output.GetReadTimeout())
	if err != nil {
		output.connection = nil
		return nil, err
	}

	output.connection = connection
	// output.updateDeadline()

	return connection, nil
}

func (output *HttpStreamOutput) GetReadTimeout() time.Duration {
	return 10 * time.Second
}


func (output *HttpStreamOutput) AudioOut(audio *Audio) {

}
