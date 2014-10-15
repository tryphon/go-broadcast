package command

import (
	"io/ioutil"
	"os"
	"projects.tryphon.eu/go-broadcast/broadcast"
	"testing"
)

func testConfigFile(content string) string {
	file, _ := ioutil.TempFile("/tmp", "config")
	defer file.Close()

	file.WriteString(content)

	return file.Name()
}

func TestCommandConfig_Load(t *testing.T) {
	config := broadcast.CommandConfig{File: testConfigFile(`{"Log": {"Syslog": true}}`)}
	defer os.Remove(config.File)
	broadcast.LoadConfig(config.File, &config)

	if config.Log.Syslog != true {
		t.Errorf(" :\n got: %v\nwant: %v", config.Log.Syslog, true)
	}
}

func TestHttpSourceConfig_Load(t *testing.T) {
	config := HttpSourceConfig{}
	config.File = "testdata/http_source_config.json"

	err := broadcast.LoadConfig(config.File, &config)
	if err != nil {
		t.Fatal(err)
	}

	if config.Log.Syslog != true {
		t.Errorf(" :\n got: %v\nwant: %v", config.Log.Syslog, true)
	}
	if config.Alsa.Device != "hw:0" {
		t.Errorf(" :\n got: %v\nwant: %v", config.Alsa.Device, "hw:0")
	}
	if len(config.Http.Streams) != 2 {
		t.Errorf(" :\n got: %v\nwant: %v", len(config.Http.Streams), 2)
	}
	if config.Http.Streams[0].Target != "http://source:secret@stream-in.tryphon.eu:8000/gobroadcast.mp3" {
		t.Errorf(" :\n got: %v\nwant: %v", config.Http.Streams[0].Target, "http://source:secret@stream-in.tryphon.eu:8000/gobroadcast.mp3")
	}
}
