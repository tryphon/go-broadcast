package broadcast

import (
	"testing"
)

func TestFileInput_Init(t *testing.T) {
	input := testFileInput()
	defer input.Close()

	if input.sndFile == nil {
		t.Errorf("FileInput.sndFile should be defined after init")
	}
}

func TestFileInput_ChannelCount(t *testing.T) {
	input := testFileInput()
	defer input.Close()

	if input.ChannelCount() != 2 {
		t.Errorf(" :\n got: %v\nwant: %v", input.ChannelCount(), 2)
	}
}

func TestFileInput_SampleRate(t *testing.T) {
	input := testFileInput()
	defer input.Close()

	if input.SampleRate() != 48000 {
		t.Errorf(" :\n got: %v\nwant: %v", input.SampleRate(), 48000)
	}
}

func TestFileInput_Read(t *testing.T) {
	input := testFileInput()
	defer input.Close()

	audios := []*Audio{}
	for {
		audio := input.Read()
		if audio == nil {
			break
		}
		audios = append(audios, audio)
	}

	sampleCount := 0
	for _, audio := range audios {
		sampleCount += audio.SampleCount()
	}

	if sampleCount != 480000 {
		t.Errorf(" :\n got: %v\nwant: %v", sampleCount, 480000)
	}
}

func TestFileInput_Close(t *testing.T) {
	input := testFileInput()
	input.Close()

	if input.sndFile != nil {
		t.Errorf("FileInput.sndFile should be nil after close")
	}
}

func testFileInput() *FileInput {
	input := FileInput{File: "testdata/sine-48000.flac"}
	input.Init()
	return &input
}
