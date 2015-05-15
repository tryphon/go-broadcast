package broadcast

import (
	"flag"
	"testing"
	"time"
)

func TestBufferedHttpStreamInputConfig_Flags(t *testing.T) {
	config := BufferedHttpStreamInputConfig{}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	config.Flags(flags, "http")

	flags.Parse([]string{"-http-url=http://stream.tryphon.eu/dummy.mp3"})
	if expectedUrl := "http://stream.tryphon.eu/dummy.mp3"; config.Url != expectedUrl {
		t.Errorf("Url should be 'url' flag value :\n got: %v\nwant: %v", config.Url, expectedUrl)
	}

	flags.Parse([]string{"-http-buffer-duration=20s"})

	if config.Buffer.Duration != 20*time.Second {
		t.Errorf("Buffer.Duration should be 'buffer-duration' flag value :\n got: %v\nwant: %v", config.Buffer.Duration, 20*time.Second)
	}

	flags.Parse([]string{"-http-buffer-low-adjust-threshold=50"})

	if config.Buffer.LowAdjustThreshold != 50 {
		t.Errorf("Buffer.Duration should be 'buffer-low-adjust-threshold' flag value :\n got: %v\nwant: %v", config.Buffer.LowAdjustThreshold, 50)
	}

	flags.Parse([]string{"-http-buffer-low-refill=10"})

	if config.Buffer.LowRefill != 10 {
		t.Errorf("Buffer.Duration should be 'buffer-low-refill' flag value :\n got: %v\nwant: %v", config.Buffer.LowRefill, 10)
	}

	flags.Parse([]string{"-http-buffer-high-adjust-threshold=50"})

	if config.Buffer.HighAdjustThreshold != 50 {
		t.Errorf("Buffer.Duration should be 'buffer-high-adjust-threshold' flag value :\n got: %v\nwant: %v", config.Buffer.HighAdjustThreshold, 50)
	}

	flags.Parse([]string{"-http-buffer-high-unfill=10"})

	if config.Buffer.HighUnfill != 10 {
		t.Errorf("Buffer.Duration should be 'buffer-high-unfill' flag value :\n got: %v\nwant: %v", config.Buffer.HighUnfill, 10)
	}
}
