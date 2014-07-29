package broadcast

import (
	"bytes"
	"math"
	"testing"
)

func TestInterleavedAudioCoder_Encode(t *testing.T) {
	encoder := InterleavedAudioCoder{SampleFormat: Sample16bLittleEndian, ChannelCount: 2}

	audio := NewAudio(2, 2)
	audio.SetSamples(0, []float32{0, 1})
	audio.SetSamples(1, []float32{1, 0})

	output, _ := encoder.Encode(audio)
	// 0.0, 1.0, 1.0, 0.0 in s16le :
	expectedOutput := []byte{0, 0, 255, 127, 255, 127, 0, 0}

	if !bytes.Equal(output, expectedOutput) {
		t.Errorf("Wrong Encode output for %v:\n got: %v\nwant: %v", audio, output, expectedOutput)
	}
}

func TestInterleavedAudioCoder_Encode_with_more_channels_than_audio(t *testing.T) {
	encoder := InterleavedAudioCoder{SampleFormat: Sample16bLittleEndian, ChannelCount: 3}

	audio := NewAudio(2, 2)
	audio.SetSamples(0, []float32{0, 1})
	audio.SetSamples(1, []float32{1, 0})

	output, _ := encoder.Encode(audio)
	// 0.0, 1.0, 0.0, 1.0, 0.0, 0.0 in s16le :
	expectedOutput := []byte{0, 0, 255, 127, 0, 0, 255, 127, 0, 0, 0, 0}

	if !bytes.Equal(output, expectedOutput) {
		t.Errorf("Wrong Encode output for %v:\n got: %v\nwant: %v", audio, output, expectedOutput)
	}
}

func TestInterleavedAudioCoder_Decode(t *testing.T) {
	coder := InterleavedAudioCoder{SampleFormat: Sample16bLittleEndian, ChannelCount: 2}

	// 0.0, 1.0, 1.0, 0.0 in s16le :
	input := []byte{0, 0, 255, 127, 255, 127, 0, 0}

	expectedAudio := NewAudio(2, 2)
	expectedAudio.SetSamples(0, []float32{0, 1})
	expectedAudio.SetSamples(1, []float32{1, 0})

	audio, _ := coder.Decode(input)

	if !sameAudio(audio, expectedAudio, 0) {
		t.Errorf("Wrong decoded audio output:\n got: %v\nwant: %v", audio, expectedAudio)
	}
}

func sameAudio(audio1 *Audio, audio2 *Audio, tolerance float64) bool {
	if audio1.SampleCount() != audio2.SampleCount() {
		return false
	}
	if audio1.ChannelCount() != audio2.ChannelCount() {
		return false
	}

	for channel := 0; channel < audio1.ChannelCount(); channel++ {
		for samplePosition := 0; samplePosition < audio1.SampleCount(); samplePosition++ {
			sample1 := audio1.Samples(channel)[samplePosition]
			sample2 := audio2.Samples(channel)[samplePosition]

			if math.Abs(float64(sample1-sample2)) > tolerance {
				return false
			}
		}
	}

	return true
}
