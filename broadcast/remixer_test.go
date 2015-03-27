package broadcast

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestNewRemixer_SingleChannels(t *testing.T) {
	remixer := NewRemixer("1:2")

	if remixer.OutputChannels[0].InputChannels[0] != 0 {
		t.Errorf("Wrong first channel input :\n got: %v\nwant: %v", remixer.OutputChannels[0].InputChannels[0], 0)
	}
	if remixer.OutputChannels[1].InputChannels[0] != 1 {
		t.Errorf("Wrong second channel input :\n got: %v\nwant: %v", remixer.OutputChannels[1].InputChannels[0], 1)
	}
}

func TestNewRemixer_MultipleChannels(t *testing.T) {
	remixer := NewRemixer("1,8:2,9")

	if !reflect.DeepEqual(remixer.OutputChannels[0].InputChannels, []int{0, 7}) {
		t.Errorf("Wrong first channel :\n got: %v\nwant: %v", remixer.OutputChannels[0].InputChannels, []int{0, 7})
	}
	if !reflect.DeepEqual(remixer.OutputChannels[1].InputChannels, []int{1, 8}) {
		t.Errorf("Wrong second channel :\n got: %v\nwant: %v", remixer.OutputChannels[1].InputChannels, []int{1, 8})
	}
}

func TestNewRemixer_EmptyChannel(t *testing.T) {
	remixer := NewRemixer("0")

	if len(remixer.OutputChannels[0].InputChannels) != 0 {
		t.Errorf("No input channel should be defined :\n got: %v", remixer.OutputChannels[0].InputChannels)
	}
}

func TestRemixer_AudioOut(t *testing.T) {
	remixer := NewRemixer("8:9:8,9:0")
	originalAudio := NewAudio(1024, 12)

	originalAudio.Process(func(channel int, samplePosition int, sample float32) float32 {
		switch channel {
		case 7:
			return 1
		case 8:
			return -1
		}
		return rand.Float32()
	})

	remixer.Output = AudioHandlerFunc(func(audio *Audio) {
		if audio.ChannelCount() != 4 {
			t.Errorf("Wrong remixed audio channel count :\n got: %v\nwant: %v", audio.ChannelCount(), 2)
		}
		if audio.SampleCount() != originalAudio.SampleCount() {
			t.Errorf("Wrong remixed audio sample count :\n got: %v\nwant: %v", audio.SampleCount(), originalAudio.SampleCount())
		}

		silence := make([]float32, originalAudio.SampleCount())

		if !reflect.DeepEqual(audio.Samples(0), originalAudio.Samples(7)) {
			t.Errorf("Remixed channel 0 doesn't contain channel 8")
		}
		if !reflect.DeepEqual(audio.Samples(1), originalAudio.Samples(8)) {
			t.Errorf("Remixed channel 1 doesn't contain channel 9")
		}
		if !reflect.DeepEqual(audio.Samples(2), silence) {
			t.Errorf("Remixed channel 2 doesn't contain a mix of channels 8 and 9")
		}
		if !reflect.DeepEqual(audio.Samples(3), silence) {
			t.Errorf("Remixed channel 3 doesn't contain silence")
		}
	})

	remixer.AudioOut(originalAudio)
}

func TestRemixerChannel_String(t *testing.T) {
	var conditions = []struct {
		inputChannels  []int
		expectedString string
	}{
		{[]int{0}, "1"},
		{[]int{0, 1}, "1,2"},
		{[]int{}, "0"},
	}

	for _, condition := range conditions {
		channel := RemixerChannel{InputChannels: condition.inputChannels}
		if channel.String() != condition.expectedString {
			t.Errorf("Wrong channel.String() :\n got: %v\nwant: %v", channel.String(), condition.expectedString)
		}
	}
}

func TestRemixer_String(t *testing.T) {
	var conditions = []struct {
		channels       [][]int
		expectedString string
	}{
		{[][]int{{0}, {1}}, "1:2"},
		{[][]int{{0, 1}, {}, {2}}, "1,2:0:3"},
	}

	for _, condition := range conditions {
		remixer := Remixer{}

		remixer.OutputChannels = make([]RemixerChannel, len(condition.channels))
		for index, inputChannels := range condition.channels {
			remixer.OutputChannels[index] = RemixerChannel{InputChannels: inputChannels}
		}

		if remixer.String() != condition.expectedString {
			t.Errorf("Wrong Remixer.String() :\n got: %v\nwant: %v", remixer.String(), condition.expectedString)
		}
	}
}
