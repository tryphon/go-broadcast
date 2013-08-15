package broadcast

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestSndFile_constants(t *testing.T) {
	var conditions = []struct {
		name       string
		goConstant int
		cConstant  int
	}{
		{"O_RDONLY", O_RDONLY, 0x10},
		{"O_WRONLY", O_WRONLY, 0x20},
		{"O_RDWR", O_RDWR, 0x30},
	}

	for _, condition := range conditions {
		if condition.goConstant != condition.cConstant {
			t.Errorf("Wrong constant value %v:\n got: %v\nwant: %v", condition.name, condition.goConstant, condition.cConstant)
		}
	}
}

func TestSndFile_Open_Read(t *testing.T) {
	file, err := SndFileOpen("testdata/empty.wav", O_RDONLY, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if file.Info().Frames() != 0 {
		t.Errorf("Wrong frame count:\n got: %v\nwant: 0", file.Info().Frames())
	}
	if file.Info().SampleRate() != 44100 {
		t.Errorf("Wrong sample rate:\n got: %v\nwant: 44100", file.Info().SampleRate())
	}
}

func TestSndFile_Open_Write(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "sndfile-open-write")
	if err != nil {
		t.Fatal(err)
	}

	fileName := tempFile.Name()
	defer os.Remove(fileName)

	var fileInfo Info
	fileInfo.SetSampleRate(44100)
	fileInfo.SetChannels(2)
	fileInfo.SetFormat(FORMAT_WAV | FORMAT_PCM_16)

	file, err := SndFileOpen(fileName, O_WRONLY, &fileInfo)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
}

func TestSnfFile_WriteFloat(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "sndfile-open-write")
	if err != nil {
		t.Fatal(err)
	}

	fileName := tempFile.Name()
	defer os.Remove(fileName)

	var fileInfo Info
	fileInfo.SetSampleRate(44100)
	fileInfo.SetChannels(2)
	fileInfo.SetFormat(FORMAT_WAV | FORMAT_PCM_16)

	file, err := SndFileOpen(fileName, O_WRONLY, &fileInfo)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	floats := make([]float32, 1024)
	writeCount := file.WriteFloat(floats)

	if writeCount != int64(len(floats)) {
		t.Errorf("Wrong write count:\n got: %v\nwant: %v", writeCount, len(floats))
	}
}
