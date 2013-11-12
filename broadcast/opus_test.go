package broadcast

import (
	"bytes"
	"testing"
	// "os"
	"hash/crc32"
)

func TestOpus_encode(t *testing.T) {
	encoder, err := OpusEncoderCreate(512000)
	if err != nil {
		t.Fatal(err)
	}
	defer encoder.Destroy()

	file, err := SndFileOpen("testdata/sine-48000.flac", O_RDONLY, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// 960 samples is 20 ms
	frameCount := 960
	samples := make([]float32, frameCount*2)

	opusBytes := make([]byte, 2048)
	opusBuffer := &bytes.Buffer{}

	for {
		readCount := file.ReadFloat(samples)

		encodedLength, err := encoder.EncodeFloat(samples, frameCount, opusBytes, 1280)
		if err != nil {
			t.Fatal(err)
		}

		opusBuffer.Write(opusBytes[:encodedLength])

		if int(readCount) != len(samples) {
			break
		}
	}

	// opusFile, err := os.Create("/tmp/opus-encoder.opus")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer opusFile.Close()
	// opusBuffer.WriteTo(opusFile)

	var expectedHash uint32 = 928466767
	hash := crc32.NewIEEE()
	opusBuffer.WriteTo(hash)
	if hash.Sum32() != expectedHash {
		t.Errorf("Wrong opus data checksum:\n got: %v\nwant: %v", hash.Sum32(), expectedHash)
	}
}

func TestOpus_decode(t *testing.T) {
	decoder, err := OpusDecoderCreate()
	if err != nil {
		t.Fatal(err)
	}
	defer decoder.Destroy()

	encoder, err := OpusEncoderCreate(512000)
	if err != nil {
		t.Fatal(err)
	}
	defer encoder.Destroy()

	file, err := SndFileOpen("testdata/sine-48000.flac", O_RDONLY, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// 960 samples is 20 ms
	frameCount := 960
	samples := make([]float32, frameCount*2)

	opusBytes := make([]byte, 2048)

	for {
		readCount := file.ReadFloat(samples)

		encodedLength, err := encoder.EncodeFloat(samples, frameCount, opusBytes, 1280)
		if err != nil {
			t.Fatal(err)
		}

		decodedFrameCount, err := decoder.DecodeFloat(opusBytes[:encodedLength], samples, frameCount)
		if err != nil {
			t.Fatal(err)
		}

		if int(decodedFrameCount) != frameCount {
			t.Errorf("Wrong decoded frame count:\n got: %v\nwant: %v", decodedFrameCount, frameCount)
		}

		if int(readCount) != len(samples) {
			break
		}
	}
}
