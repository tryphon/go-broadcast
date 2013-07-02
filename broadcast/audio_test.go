package broadcast

import (
	"bytes"
	"encoding/binary"
	"testing"
)

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
