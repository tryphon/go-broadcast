package broadcast

import (
	"flag"
	"strings"
	"testing"
)

func TestProfilerConfig_Flags(t *testing.T) {
	config := ProfilerConfig{}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	config.Flags(flags, "profiler")

	flags.Parse(strings.Split("-profiler-cpu cpu-output -profiler-memory memory-output", " "))
	if config.CPU != "cpu-output" {
		t.Errorf("Wrong config CPU :\n got: %v\nwant: %v", config.CPU, "cpu-output")
	}
	if config.Memory != "memory-output" {
		t.Errorf("Wrong config Memory :\n got: %v\nwant: %v", config.Memory, "memory-output")
	}
}
