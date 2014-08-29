package broadcast

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestLameEncoder_Init(t *testing.T) {
	encoder := LameEncoder{}
	encoder.Init()

	if encoder.handle == nil {
		t.Errorf("Encoder should have lame handle after Init")
	}
}

func TestLameEncoder_LameQuality(t *testing.T) {
	encoder := LameEncoder{}

	var conditions = []struct {
		quality     float32
		lameQuality int
	}{
		{1, 0},
		{0, 9},
		{0.5, 5},
	}

	for _, condition := range conditions {
		encoder.Quality = condition.quality
		if encoder.LameQuality() != condition.lameQuality {
			t.Errorf("With encoder.Quality = %f :\n got: %v\nwant: %v", condition.quality, encoder.LameQuality(), condition.lameQuality)
		}
	}
}

func TestLameEncoder_LameMode(t *testing.T) {
	encoder := LameEncoder{ChannelCount: 2}
	if encoder.LameMode() != JOINT_STEREO {
		t.Errorf(" :\n got: %v\nwant: %v", encoder.LameMode(), JOINT_STEREO)
	}

	encoder.ChannelCount = 1
	if encoder.LameMode() != MONO {
		t.Errorf(" :\n got: %v\nwant: %v", encoder.LameMode(), MONO)
	}
}

func TestLameEncoder_Close(t *testing.T) {
	encoder := LameEncoder{}
	encoder.Init()
	encoder.Close()

	if encoder.handle != nil {
		t.Errorf("Encoder should not have lame handle after Close")
	}
}

func TestLameEncoder_AudioOut(t *testing.T) {
	input := FileInput{File: "testdata/sine-48000.flac"}
	input.Init()
	defer input.Close()

	var buffer bytes.Buffer

	encoder := LameEncoder{SampleRate: input.SampleRate(), Quality: 1, Writer: &buffer}
	encoder.Init()

	for {
		audio := input.Read()
		if audio == nil {
			break
		}
		encoder.AudioOut(audio)
	}

	encoder.Flush()

	expectedBytes, err := ioutil.ReadFile("testdata/lame_encoder_sine_output.mp3")
	if err != nil {
		t.Fatal(err)
	}

	if buffer.Len() != len(expectedBytes) {
		t.Errorf("Wrong buffer length :\n got: %v\nwant: %v", buffer.Len(), len(expectedBytes))
	}

	// if !reflect.DeepEqual(buffer.Bytes(), expectedBytes) {
	// 	ioutil.WriteFile("lame_encoder_sine_output.mp3", buffer.Bytes(), 644)
	// 	t.Errorf("Encoded bytes doesn't match expected output")
	// }
}
