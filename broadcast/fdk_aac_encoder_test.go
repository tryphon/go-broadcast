package broadcast

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestFDKAACEncoder_AudioOut(t *testing.T) {
	input := FileInput{File: "testdata/sine-48000.flac"}
	input.Init()
	defer input.Close()

	var buffer bytes.Buffer

	encoder := FDKAACEncoder{
		SampleRate:   input.SampleRate(),
		ChannelCount: 2,
		Writer:       &buffer,
	}
	err := encoder.Init()
	if err != nil {
		t.Fatal(err)
	}

	for {
		audio := input.Read()
		if audio == nil {
			break
		}
		encoder.AudioOut(audio)
	}

	encoder.Close()

	ioutil.WriteFile("testdata/fdk_aac_encoder_sine_output.aacp", buffer.Bytes(), 0644)
}
