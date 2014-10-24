package broadcast

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestAACPEncoder_AudioOut(t *testing.T) {
	input := FileInput{File: "testdata/sine-48000.flac"}
	input.Init()
	defer input.Close()

	var buffer bytes.Buffer

	encoder := AACPEncoder{
		SampleRate: input.SampleRate(),
		Writer:     &buffer,
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

	ioutil.WriteFile("testdata/lame_encoder_sine_output.aacp", buffer.Bytes(), 0644)

	// expectedBytes, err := ioutil.ReadFile("testdata/lame_encoder_sine_output.mp3")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// if buffer.Len() != len(expectedBytes) {
	// 	t.Errorf("Wrong buffer length :\n got: %v\nwant: %v", buffer.Len(), len(expectedBytes))
	// }

	// if !reflect.DeepEqual(buffer.Bytes(), expectedBytes) {
	// 	ioutil.WriteFile("lame_encoder_sine_output.mp3", buffer.Bytes(), 644)
	// 	t.Errorf("Encoded bytes doesn't match expected output")
	// }
}
