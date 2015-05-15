package broadcast

import (
	"flag"
	"testing"
	"time"
)

func TestHttpStreamInputConfig_Flags(t *testing.T) {
	config := HttpStreamInputConfig{}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	config.Flags(flags, "http")

	flags.Parse([]string{"-http-url=http://stream.tryphon.eu/dummy.mp3"})
	if expectedUrl := "http://stream.tryphon.eu/dummy.mp3"; config.Url != expectedUrl {
		t.Errorf("Url should be 'url' flag value :\n got: %v\nwant: %v", config.Url, expectedUrl)
	}

	flags.Parse([]string{"-http-readtimeout=20s"})
	if expectedTimeout := 20 * time.Second; config.ReadTimeout != expectedTimeout {
		t.Errorf("ReadTimmeout should be 'readtimeout' flag value :\n got: %v\nwant: %v", config.ReadTimeout, expectedTimeout)
	}

	flags.Parse([]string{"-http-waitonerror=20s"})
	if expected := 20 * time.Second; config.WaitOnError != expected {
		t.Errorf("ReadTimmeout should be 'waitonerror' flag value :\n got: %v\nwant: %v", config.WaitOnError, expected)
	}
}
