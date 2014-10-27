package broadcast

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type MockIcecast struct {
	server *httptest.Server

	invoked bool
	channel chan bool

	Request *http.Request
}

func NewMockIcecast() *MockIcecast {
	mock := &MockIcecast{}
	mock.Init()
	return mock
}

func (icecast *MockIcecast) Init() {
	icecast.server = httptest.NewServer(icecast)
	icecast.channel = make(chan bool)
}

func (icecast *MockIcecast) Wait() {
	icecast.invoked = <-icecast.channel
}

func (icecast *MockIcecast) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	icecast.Request = request

	hj, ok := response.(http.Hijacker)
	if !ok {
		http.Error(response, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, readerWriter, err := hj.Hijack()
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	go func() {
		// Don't forget to close the connection:
		defer conn.Close()

		readerWriter.WriteString("HTTP/1.1 200 OK\r\n")
		readerWriter.WriteString("\r\n")
		readerWriter.Flush()

		icecast.channel <- true
		ioutil.ReadAll(readerWriter)
	}()
}

func (icecast *MockIcecast) Close() {
	if icecast.server != nil {
		icecast.server.Close()
	}
}

func (icecast *MockIcecast) URL(password, mountPoint string) string {
	streamURL, _ := url.Parse(icecast.server.URL)
	streamURL.User = url.UserPassword("source", password)
	streamURL.Path = mountPoint
	return streamURL.String()
}

func TestHttpStreamOutput_Icecast2(t *testing.T) {
	icecast := NewMockIcecast()
	defer icecast.Close()

	stream := HttpStreamOutput{
		Target: icecast.URL("secret", "/test.ogg"),
		Format: AudioFormat{Encoding: "ogg/vorbis"},
		Provider: AudioProviderFunc(func() *Audio {
			return NewAudio(1024, 2)
		}),
		Description: &StreamDescription{Name: "Test stream"},
	}
	stream.Init()
	stream.Start()

	icecast.Wait()

	stream.Stop()

	if icecast.Request.Method != "SOURCE" {
		t.Errorf("Wrong request method :\n got: %v\nwant: %v", icecast.Request.Method, "SOURCE")
	}

	if icecast.Request.Header.Get("Content-Type") != "application/ogg" {
		t.Errorf("Wrong content type :\n got: %v\nwant: %v", icecast.Request.Header.Get("Content-Type"), "application/ogg")
	}

	if icecast.Request.URL.Path != "/test.ogg" {
		t.Errorf("Wrong URL path :\n got: %v\nwant: %v", icecast.Request.URL.Path, "/test.ogg")
	}

	if icecast.Request.Header.Get("Authorization") != "Basic c291cmNlOnNlY3JldA==" {
		t.Errorf("Wrong Authentication :\n got: %v\nwant: %v", icecast.Request.Header.Get("Authorization"), "Basic c291cmNlOnNlY3JldA==")
	}

	if icecast.Request.Header.Get("Ice-Name") != "Test stream" {
		t.Errorf("Wrong ice attributes :\n got: %v\nwant: %v", icecast.Request.Header.Get("Ice-Name"), "Test stream")
	}
}
