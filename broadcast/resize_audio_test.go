package broadcast

import (
	"testing"
)

func sameSamples(samples1 []float32, samples2 []float32) (bool) {
	if len(samples1) != len(samples2) {
		return false
	}
	for index := range samples1 {
		if samples1[index] != samples2[index] {
			return false
		}
	}
	return true
}

func makeSampleSequence(sampleCount int) ([]float32) {
	sequence := make([]float32, sampleCount)
	for index := range sequence {
		sequence[index] = float32(index % 3 - 1)
	}
	return sequence
}

func TestResizeAudio_resize_smaller(t *testing.T) {
	sequence := makeSampleSequence(1024)

	expectedSampleCount := 2
	receivedSampleCount := 0

	audioHandler := AudioHandlerFunc(func(audio *Audio) {
		if audio.SampleCount() != expectedSampleCount {
			t.Errorf("Wrong sample count:\n got: %v\nwant: %v", audio.SampleCount(), expectedSampleCount)
		}

		expectedSamples := sequence[receivedSampleCount:receivedSampleCount+expectedSampleCount]
		if ! sameSamples(audio.Samples(0), expectedSamples) {
			t.Errorf("Wrong sample values:\n got: %v\nwant: %v", audio.Samples(0), expectedSamples)
		}

		receivedSampleCount += audio.SampleCount()
	})
	resizeAudio := ResizeAudio{SampleCount: expectedSampleCount, ChannelCount: 2, Output: audioHandler}

	audio := NewAudio(1024,2)
	audio.SetSamples(0, sequence)
	audio.SetSamples(1, sequence)

	resizeAudio.AudioOut(audio)

	if receivedSampleCount != audio.SampleCount() {
		t.Errorf("Wrong received sample count:\n got: %v\nwant: %v", receivedSampleCount, audio.SampleCount())
	}
}

func TestResizeAudio_resize_bigger(t *testing.T) {
	sequence := makeSampleSequence(1024)

	receivedSampleCount := 0

	audioHandler := AudioHandlerFunc(func(audio *Audio) {
		if audio.SampleCount() != len(sequence) {
			t.Errorf("Wrong sample count:\n got: %v\nwant: %v", audio.SampleCount(), len(sequence))
		}

		if ! sameSamples(audio.Samples(0), sequence) {
			t.Errorf("Wrong sample values:\n got: %v\nwant: %v", audio.Samples(0), sequence)
		}

		receivedSampleCount += audio.SampleCount()
	})

	resizeAudio := ResizeAudio{SampleCount: 1024, ChannelCount: 2, Output: audioHandler}

	audioLength := 2
	sliceCount := len(sequence) / audioLength

	for slice := 0; slice < sliceCount; slice++ {
		firstSampleSlice := slice * audioLength
		lastSampleSlice := (slice + 1) * audioLength

		audio := NewAudio(2,2)
		audio.SetSamples(0, sequence[firstSampleSlice:lastSampleSlice])
		audio.SetSamples(1, sequence[firstSampleSlice:lastSampleSlice])

		resizeAudio.AudioOut(audio)
	}

	if receivedSampleCount != len(sequence) {
		t.Errorf("Wrong received sample count:\n got: %v\nwant: %v", receivedSampleCount, len(sequence))
	}
}
