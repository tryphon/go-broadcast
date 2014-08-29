package broadcast

import (
	"testing"
)

func TestLameEncoder_Init(t *testing.T) {
	encoder := LameEncoder{}
	encoder.Init()

	if encoder.handle == nil {
		t.Errorf("Encoder should have lame handle after Init")
	}
}

func TestLameEncoder_LameQuality(t *testing.T) {
	encoder := LameEncoder{}

	var conditions = []struct {
		quality     float32
		lameQuality int
	}{
		{1, 0},
		{0, 9},
		{0.5, 5},
	}

	for _, condition := range conditions {
		encoder.Quality = condition.quality
		if encoder.LameQuality() != condition.lameQuality {
			t.Errorf("With encoder.Quality = %f :\n got: %v\nwant: %v", condition.quality, encoder.LameQuality(), condition.lameQuality)
		}
	}
}

func TestLameEncoder_LameMode(t *testing.T) {
	encoder := LameEncoder{ChannelCount: 2}
	if encoder.LameMode() != JOINT_STEREO {
		t.Errorf(" :\n got: %v\nwant: %v", encoder.LameMode(), JOINT_STEREO)
	}

	encoder.ChannelCount = 1
	if encoder.LameMode() != MONO {
		t.Errorf(" :\n got: %v\nwant: %v", encoder.LameMode(), MONO)
	}
}

func TestLameEncoder_Close(t *testing.T) {
	encoder := LameEncoder{}
	encoder.Init()
	encoder.Close()

	if encoder.handle != nil {
		t.Errorf("Encoder should not have lame handle after Close")
	}
}
