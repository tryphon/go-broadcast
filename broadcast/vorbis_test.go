package broadcast

import (
	"io"
	"math"
	"testing"
	"time"
)

func TestVorbis_Encode_Decode(t *testing.T) {
	reader, writer := io.Pipe()

	oggEncoder := OggEncoder{
		Encoder: VorbisEncoder{
			Quality:      1,
			ChannelCount: 2,
			SampleRate:   44100,
		},
		Writer: writer,
	}
	oggEncoder.Encoder.PacketHandler = &oggEncoder

	var oggDecoder OggDecoder
	var vorbisDecoder VorbisDecoder

	oggDecoder.SetHandler(&vorbisDecoder)

	var receivedAudios []*Audio

	audioHandler := AudioHandlerFunc(func(audio *Audio) {
		receivedAudios = append(receivedAudios, audio)
	})
	vorbisDecoder.audioHandler = audioHandler

	go func() {
		for oggDecoder.Read(reader) {
		}
	}()

	sampleCount := 1000

	oggEncoder.Init()
	for number := 0; number < sampleCount; number++ {
		oggEncoder.AudioOut(NewAudio(1024, 2))
	}
	oggEncoder.Flush()

	for len(receivedAudios) <= int(math.Floor(float64(sampleCount)*0.98)) {
		time.Sleep(time.Second)
	}

	writer.Close()
	reader.Close()
}
