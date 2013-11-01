package broadcast

import (
	alsa "github.com/tryphon/alsa-go"
)

func ToAlsaSampleFormat(sampleFormat SampleFormat) alsa.SampleFormat {
	switch sampleFormat {
	case Sample16bLittleEndian:
		return alsa.SampleFormatS16LE
	case Sample32bLittleEndian:
		return alsa.SampleFormatS32LE
	}
	return 0
}

func FromAlsaSampleFormat(alsaSampleFormat alsa.SampleFormat) SampleFormat {
	switch alsaSampleFormat {
	case alsa.SampleFormatS16LE:
		return Sample16bLittleEndian
	case alsa.SampleFormatS32LE:
		return Sample32bLittleEndian
	}
	return nil
}
