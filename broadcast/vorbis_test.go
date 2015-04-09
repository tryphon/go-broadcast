package broadcast

import (
	"bytes"
	"testing"
)

func BenchmarkVorbis_EncodeDecode(b *testing.B) {
	memBenchmark := NewMemoryBenchmark(b)
	defer memBenchmark.Complete()

	for i := 0; i < b.N; i++ {
		testVorbisEncoderDecoder().Benchmark(b)
	}
}

func BenchmarkVorbis_EncodeDecodeCBR(b *testing.B) {
	memBenchmark := NewMemoryBenchmark(b)
	defer memBenchmark.Complete()

	for i := 0; i < b.N; i++ {
		encoderDecoder := testVorbisEncoderDecoder()

		encoderDecoder.VorbisEncoder.Mode = "cbr"
		encoderDecoder.VorbisEncoder.BitRate = 128000

		encoderDecoder.Benchmark(b)
	}
}

func BenchmarkVorbis_EncodeDecodeABR(b *testing.B) {
	memBenchmark := NewMemoryBenchmark(b)
	defer memBenchmark.Complete()

	for i := 0; i < b.N; i++ {
		encoderDecoder := testVorbisEncoderDecoder()

		encoderDecoder.VorbisEncoder.Mode = "abr"
		encoderDecoder.VorbisEncoder.BitRate = 128000

		encoderDecoder.Benchmark(b)
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

type TestVorbisEncoderDecoder struct {
	OggEncoder    *OggEncoder
	VorbisEncoder *VorbisEncoder
	OggDecoder    OggDecoder
	VorbisDecoder VorbisDecoder
	SampleCount   int

	ReceivedAudios []*Audio
	Buffer         *bytes.Buffer
}

func (encoderDecoder *TestVorbisEncoderDecoder) Init() {
	encoderDecoder.VorbisEncoder = &VorbisEncoder{
		Quality:      1,
		ChannelCount: 2,
		SampleRate:   44100,
	}

	encoderDecoder.Buffer = bytes.NewBuffer(make([]byte, 4096*1000))
	encoderDecoder.OggEncoder = &OggEncoder{
		Encoder: encoderDecoder.VorbisEncoder,
		Writer:  encoderDecoder.Buffer,
	}

	encoderDecoder.OggDecoder.SetHandler(&encoderDecoder.VorbisDecoder)

	encoderDecoder.VorbisDecoder.audioHandler = AudioHandlerFunc(func(audio *Audio) {
		encoderDecoder.ReceivedAudios = append(encoderDecoder.ReceivedAudios, audio)
	})

	encoderDecoder.SampleCount = 1000
}

func (encoderDecoder *TestVorbisEncoderDecoder) Run() error {
	err := encoderDecoder.OggEncoder.Init()
	if err != nil {
		return err
	}

	for number := 0; number < encoderDecoder.SampleCount; number++ {
		encoderDecoder.OggEncoder.AudioOut(NewAudio(1024, 2))
	}
	encoderDecoder.OggEncoder.Close()

	for encoderDecoder.OggDecoder.Read(encoderDecoder.Buffer) {
	}

	encoderDecoder.OggDecoder.Reset()
	encoderDecoder.VorbisDecoder.Reset()
	return nil
}

func (encoderDecoder *TestVorbisEncoderDecoder) Check() bool {
	receivedSamples := 0
	for _, receivedAudio := range encoderDecoder.ReceivedAudios {
		receivedSamples += receivedAudio.SampleCount()
	}

	return receivedSamples == encoderDecoder.SampleCount*1024
}

func (encoderDecoder *TestVorbisEncoderDecoder) Benchmark(b *testing.B) {
	if err := encoderDecoder.Run(); err != nil {
		b.Fatal(err)
	}
	if !encoderDecoder.Check() {
		b.Errorf("Encode/decode not validated")
	}
}

func testVorbisEncoderDecoder() *TestVorbisEncoderDecoder {
	encoderDecoder := &TestVorbisEncoderDecoder{}
	encoderDecoder.Init()
	return encoderDecoder
}
