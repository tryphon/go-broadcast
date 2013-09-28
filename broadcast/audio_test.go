package broadcast

import (
	"bytes"
	"encoding/binary"
	alsa "github.com/tryphon/alsa-go"
	"testing"
)

func TestAudio_floatSamplesToBytes(t *testing.T) {
	var conditions = []struct {
		floatSample float32
		bytesSample [2]byte
	}{
		{1, [2]byte{255, 127}},
		{0, [2]byte{0, 0}},
		{-1, [2]byte{1, 128}},
	}

	for i, condition := range conditions {
		byte1, byte2 := floatSamplesToBytes(condition.floatSample)
		bytes := [2]byte{byte1, byte2}
		if bytes != condition.bytesSample {
			t.Errorf("#%d: Wrong bytes value for %v:\n got: %v\nwant: %v", i, condition.floatSample, bytes, condition.bytesSample)
		}
	}
}

func TestAudio_pcmSample16BitsToFloat(t *testing.T) {
	var conditions = []struct {
		pcmSample   int16
		floatSample float32
	}{
		{32767, 1},
		{0, 0},
		{-32768, -1},
	}

	for i, condition := range conditions {
		floatSample := pcmSample16BitsToFloat(condition.pcmSample)
		if floatSample != condition.floatSample {
			t.Errorf("#%d: Wrong float value for %v:\n got: %v\nwant: %v", i, condition.pcmSample, floatSample, condition.floatSample)
		}
	}
}

func TestAudio_pcmSample32BitsToFloat(t *testing.T) {
	var conditions = []struct {
		pcmSample   int32
		floatSample float32
	}{
		{2147483647, 1},
		{0, 0},
		{-2147483648, -1},
	}

	for i, condition := range conditions {
		floatSample := pcmSample32BitsToFloat(condition.pcmSample)
		if floatSample != condition.floatSample {
			t.Errorf("#%d: Wrong float value for %v:\n got: %v\nwant: %v", i, condition.pcmSample, floatSample, condition.floatSample)
		}
	}
}

func TestAudio_PcmBytes_Length(t *testing.T) {
	var conditions = []struct {
		sampleCount  int
		channelCount int
	}{
		{1024, 2},
		{1024, 4},
	}

	for i, condition := range conditions {
		audio := Audio{sampleCount: condition.sampleCount, channelCount: condition.channelCount}

		byteCount := len(audio.PcmBytes())
		expectedByteCount := condition.sampleCount * condition.channelCount * 2

		if byteCount != expectedByteCount {
			t.Errorf("#%d: Wrong length:\n got: %d\nwant: %d", i, byteCount, expectedByteCount)
		}
	}
}

func TestAudio_PcmBytes_ChannelContent(t *testing.T) {
	audio := Audio{sampleCount: 4, channelCount: 2}

	for activeChannel := 0; activeChannel < audio.channelCount; activeChannel++ {
		audio.samples = make([][]float32, audio.channelCount)

		audio.samples[activeChannel] = make([]float32, audio.sampleCount)
		for samplePosition := 0; samplePosition < audio.sampleCount; samplePosition++ {
			audio.samples[activeChannel][samplePosition] = 1
		}

		buffer := bytes.NewBuffer(audio.PcmBytes())

		for samplePosition := 0; samplePosition < audio.sampleCount; samplePosition++ {
			for channel := 0; channel < audio.channelCount; channel++ {
				var pcmSample int16
				err := binary.Read(buffer, binary.LittleEndian, &pcmSample)
				if err != nil {
					t.Errorf("binary.Read failed %v", err)
				}

				var expectedPcmSample int16

				if channel == activeChannel {
					expectedPcmSample = 32767
				} else {
					expectedPcmSample = 0
				}

				if pcmSample != expectedPcmSample {
					t.Errorf("#sample:%d,channel:%d: Wrong pcm sample value:\n got: %d\nwant: %d", samplePosition, channel, pcmSample, expectedPcmSample)
				}
			}
		}

	}
}

func TestAudio_LoadPcmBytes(t *testing.T) {
	audio := Audio{}
	audio.LoadPcmBytes([]byte{255, 127, 0, 0, 0, 0, 1, 128}, 2, 2, alsa.SampleFormatS16LE)

	if audio.SampleCount() != 2 {
		t.Errorf("Wrong sample count value:\n got: %d\nwant: %d", audio.SampleCount(), 2)
	}
	if audio.Samples(0)[0] != 1 {
		t.Errorf("Wrong sample value:\n got: %d\nwant: %d", audio.Samples(0)[0], 1)
	}
	if audio.Samples(0)[1] != 0 {
		t.Errorf("Wrong sample value:\n got: %d\nwant: %d", audio.Samples(0)[1], 0)
	}
	if audio.Samples(1)[0] != 0 {
		t.Errorf("Wrong sample value:\n got: %d\nwant: %d", audio.Samples(1)[0], 0)
	}
	if audio.Samples(1)[1] != -1 {
		t.Errorf("Wrong sample value:\n got: %d\nwant: %d", audio.Samples(1)[1], -1)
	}
}

func TestAudio_SampleCount(t *testing.T) {
	audio := Audio{sampleCount: 1024}

	if audio.SampleCount() != 1024 {
		t.Errorf("Wrong SampleCount() value:\n got: %d\nwant: %d", audio.SampleCount(), 1024)
	}

}

func TestAudio_InterleavedFloats(t *testing.T) {
	audio := NewAudio(1024, 4)

	// Fill channel 0 with 0, channel 1 with 1, channel 2 with 2, ...
	for channel := 0; channel < audio.ChannelCount(); channel++ {
		samples := make([]float32, audio.SampleCount())
		for samplePosition := 0; samplePosition < audio.SampleCount(); samplePosition++ {
			samples[samplePosition] = float32(channel)
		}
		audio.SetSamples(channel, samples)
	}

	for position, float := range audio.InterleavedFloats() {
		expectedFloat := float32(position % audio.ChannelCount())
		if float != expectedFloat {
			t.Errorf("#sample:%d Wrong float sample value:\n got: %d\nwant: %d", position, float, expectedFloat)
		}
	}
}
