package broadcast

import (
	"io"
	"os"
	"testing"
)

func BenchmarkMadDecoder_Decode(b *testing.B) {
	// Log.Debug = true

	decoder := MadDecoder{}
	decoder.Init()

	file, err := os.Open("testdata/sine-48000.mp3")
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	var audios []*Audio

	decoder.SetAudioHandler(AudioHandlerFunc(func(audio *Audio) {
		audios = append(audios, audio)
	}))

	for {
		err := decoder.Read(file)
		if err == io.EOF {
			break
		}
		if err != nil {
			b.Fatal(err)
		}
	}

	sampleCount := 0
	for _, audio := range audios {
		sampleCount += audio.SampleCount()
	}
	// 480384
	Log.Debugf("Decoded sample count: %d", sampleCount)
}
