package broadcast

import (
	"bytes"
	"testing"
)

func BenchmarkVorbis_Encode_Decode(b *testing.B) {
	memBenchmark := NewMemoryBenchmark(b)
	defer memBenchmark.Complete()

	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer(make([]byte, 4096*1000))

		oggEncoder := OggEncoder{
			Encoder: VorbisEncoder{
				Quality:      1,
				ChannelCount: 2,
				SampleRate:   44100,
			},
			Writer: buffer,
		}

		var oggDecoder OggDecoder
		var vorbisDecoder VorbisDecoder

		oggDecoder.SetHandler(&vorbisDecoder)

		var receivedAudios []*Audio

		audioHandler := AudioHandlerFunc(func(audio *Audio) {
			receivedAudios = append(receivedAudios, audio)
		})
		vorbisDecoder.audioHandler = audioHandler

		sampleCount := 1000

		oggEncoder.Init()
		for number := 0; number < sampleCount; number++ {
			oggEncoder.AudioOut(NewAudio(1024, 2))
		}
		oggEncoder.Close()

		for oggDecoder.Read(buffer) {
		}

		receivedSamples := 0
		for _, receivedAudio := range receivedAudios {
			receivedSamples += receivedAudio.SampleCount()
		}

		if receivedSamples != sampleCount*1024 {
			b.Errorf("Wrong number of decoded samples :\n got: %v\nwant: %v", receivedSamples, sampleCount*1024)
		}

		oggDecoder.Reset()
		vorbisDecoder.Reset()
	}
}

func BenchmarkVorbisDecoder_InitReset(b *testing.B) {
	memBenchmark := NewMemoryBenchmark(b)
	defer memBenchmark.Complete()

	for i := 0; i < b.N; i++ {
		vorbisDecoder := VorbisDecoder{}
		vorbisDecoder.NewStream(1)
		vorbisDecoder.Reset()
	}
}

func BenchmarkVorbisEncoder_InitReset(b *testing.B) {
	memBenchmark := NewMemoryBenchmark(b)
	defer memBenchmark.Complete()

	for i := 0; i < b.N; i++ {
		vorbisEncoder := VorbisEncoder{}
		vorbisEncoder.Init()
		vorbisEncoder.Reset()
	}
}
