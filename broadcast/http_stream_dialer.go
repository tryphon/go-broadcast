package broadcast

import (
	"errors"
	"fmt"
	shoutcast "github.com/tryphon/go-shoutcast"
	"net"
	"net/http"
	"net/url"
)

type HttpStreamDialer interface {
	Connect(output *HttpStreamOutput) (net.Conn, error)
}

type Icecast2Dialer struct {
}

func (dialer *Icecast2Dialer) Connect(output *HttpStreamOutput) (net.Conn, error) {
	var connection net.Conn

	dialTimeout := func(network, addr string) (net.Conn, error) {
		newConnection, err := net.DialTimeout(network, addr, output.GetWriteTimeout())
		if err != nil {
			return nil, err
		}

		connection = newConnection
		return newConnection, nil
	}

	transport := http.Transport{Dial: dialTimeout}
	client := http.Client{Transport: &transport}

	request, err := http.NewRequest("SOURCE", output.Target, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-type", output.Format.ContentType())
	request.Header.Add("User-Agent", "Go Broadcast v0")

	if output.Description != nil {
		for attribute, value := range output.Description.IcecastHeaders() {
			Log.Debugf("IceCast header: %s=%s", attribute, value)
			request.Header.Add(attribute, value)
		}
	}

	// request.SetBasicAuth("source", password)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	Log.Debugf("HTTP Response : %s", response.Status)
	if response.Status != "200 OK" {
		err = fmt.Errorf("Server Error : %s", response.Status)
		return nil, err
	}

	return connection, nil
}

type ShoutcastDialer struct {
}

func (dialer *ShoutcastDialer) Client(output *HttpStreamOutput) (*shoutcast.Client, error) {
	targetURL, err := url.Parse(output.Target)
	if err != nil {
		return nil, err
	}
	password, ok := targetURL.User.Password()
	if !ok {
		return nil, errors.New("No specified password")
	}

	description := output.Description
	if description == nil {
		description = &StreamDescription{}
	}
	headers := description.ShoutcastHeaders()
	headers["content-type"] = output.Format.ContentType()

	Log.Debugf("ShoutCast headers: %v", headers)

	client := &shoutcast.Client{
		Host:     targetURL.Host,
		Password: password,
		Timeout:  output.GetWriteTimeout(),
		Headers:  headers,
	}
	return client, nil
}

func (dialer *ShoutcastDialer) Connect(output *HttpStreamOutput) (net.Conn, error) {
	client, err := dialer.Client(output)
	if err != nil {
		return nil, err
	}
	return client.Connect()
}
