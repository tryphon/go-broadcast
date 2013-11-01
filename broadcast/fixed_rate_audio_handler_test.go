package broadcast

import (
	"testing"
	"time"
)

func TestFixedRateAudioHandler_audioDuration(t *testing.T) {
	audio := NewAudio(960, 2)
	fixedRateHandler := FixedRateAudioHandler{SampleRate: 48000}

	if fixedRateHandler.audioDuration(audio) != 20*time.Millisecond {
		t.Errorf("Wrong audioDuration duration:\n got: %d\nwant: %d", fixedRateHandler.audioDuration(audio), 20*time.Millisecond)
	}
}
