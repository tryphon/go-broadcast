package broadcast

import (
	"flag"
	"math"
	"testing"
)

func TestProcessing_Config(t *testing.T) {
	processing := Processing{}

	config := processing.Config()
	if config.Amplification != 0 {
		t.Errorf("Wrong Amplification :\n got: %v\nwant: %v", config.Amplification, 0)
	}

	processing.Setup(&ProcessingConfig{Amplification: -3})
	config = processing.Config()
	if config.Amplification != -3 {
		t.Errorf("Wrong Amplification :\n got: %v\nwant: %v", config.Amplification, -3)
	}
}

func TestProcessingConfig_Flags(t *testing.T) {
	config := ProcessingConfig{}

	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	config.Flags(flags, "processing")

	flags.Parse([]string{"-processing-amplification=1"})

	if config.Amplification != 1 {
		t.Errorf("Amplification should be 'amplification' flag value :\n got: %v\nwant: %v", config.Amplification, 1)
	}
}

func TestProcessingConfig_Apply(t *testing.T) {
	processing := &Processing{}
	config := ProcessingConfig{Amplification: 3}
	config.Apply(processing)

	if math.Abs(float64(processing.amplifier.Amplification-1)) > 0.01 {
		t.Errorf("Processing amplifier should be setup with given dB amplication :\n got: %v\nwant: %v", processing.amplifier.Amplification, 1)
	}
}

func TestProcessingConfig_peakAmplification(t *testing.T) {
	var conditions = []struct {
		dbAmplication   float64
		peakAmplication float32
	}{
		{0, 0},
		{-3, -0.5},
		{3, 1},
	}

	for _, condition := range conditions {
		config := ProcessingConfig{Amplification: condition.dbAmplication}

		if math.Abs(float64(config.peakAmplification()-condition.peakAmplication)) > 0.01 {
			t.Errorf("Wrong peak amplification for %f dBFS :\n got: %v\nwant: %v", condition.dbAmplication, config.peakAmplification(), condition.peakAmplication)
		}
	}
}
