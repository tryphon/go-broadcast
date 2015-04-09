package broadcast

import (
	"flag"
	"strings"
)

type Processing struct {
	Output AudioHandler

	amplifier Amplifier
	config    *ProcessingConfig
}

func (processing *Processing) AudioOut(audio *Audio) {
	if processing.amplifier.Output == nil {
		processing.amplifier.Output = processing.Output
	}

	processing.amplifier.AudioOut(audio)
}

func (processing *Processing) Setup(config *ProcessingConfig) {
	processing.amplifier.Amplification = float32(config.peakAmplification())
	processing.config = config
}

func (processing *Processing) Config() *ProcessingConfig {
	if processing.config != nil {
		return processing.config
	} else {
		return &ProcessingConfig{}
	}
}

type ProcessingConfig struct {
	Amplification float64
}

func (config *ProcessingConfig) peakAmplification() float32 {
	return float32(dBFSToPeak(config.Amplification)) - 1
}

func (config *ProcessingConfig) Flags(flags *flag.FlagSet, prefix string) {
	flags.Float64Var(&config.Amplification, strings.Join([]string{prefix, "amplification"}, "-"), 0, "The amplification in dBFS applied to the audio signal")
}

func (config *ProcessingConfig) Apply(processing *Processing) {
	processing.Setup(config)
}
