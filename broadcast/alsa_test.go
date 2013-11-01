package broadcast

import (
	alsa "github.com/tryphon/alsa-go"
	"testing"
)

var alsaSampleFormatMapping = []struct {
	alsaSampleFormat alsa.SampleFormat
	sampleFormat     SampleFormat
}{
	{alsa.SampleFormatS16LE, Sample16bLittleEndian},
	{alsa.SampleFormatS32LE, Sample32bLittleEndian},
}

func TestToAlsaSampleFormat(t *testing.T) {
	for _, condition := range alsaSampleFormatMapping {
		alsaSampleFormat := ToAlsaSampleFormat(condition.sampleFormat)
		if alsaSampleFormat != condition.alsaSampleFormat {
			t.Errorf("Wrong alsa SampleFormat output for %v:\n got: %v\nwant: %v", condition.sampleFormat.Name(), alsaSampleFormat, condition.alsaSampleFormat)
		}
	}
}

func TestFromAlsaSampleFormat(t *testing.T) {
	for _, condition := range alsaSampleFormatMapping {
		sampleFormat := FromAlsaSampleFormat(condition.alsaSampleFormat)
		if sampleFormat != condition.sampleFormat {
			t.Errorf("Wrong SampleFormat for %v:\n got: %v\nwant: %v", condition.alsaSampleFormat, sampleFormat, condition.sampleFormat)
		}
	}
}
