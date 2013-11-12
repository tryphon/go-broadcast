package broadcast

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewSoundMetrics(t *testing.T) {
	audio := NewAudio(1024, 2)
	audio.SetSample(0, 512, 1)

	soundMetrics := NewSoundMetrics(audio)

	if soundMetrics.ChannelCount() != audio.ChannelCount() {
		t.Errorf("Should have a SoundChannelMetrics for each channel:\n got: %d\nwant: %d", soundMetrics.ChannelCount(), audio.ChannelCount())
	}

	channelOnePeakLevel := soundMetrics.ChannelMetrics[0]
	if channelOnePeakLevel.PeakLevel != 1 {
		t.Errorf("Wrong peak level on channel one:\n got: %d\nwant: %d", channelOnePeakLevel.PeakLevel, 1)
	}

	channelTwoPeakLevel := soundMetrics.ChannelMetrics[1]
	if channelTwoPeakLevel.PeakLevel != 0 {
		t.Errorf("Wrong peak level on channel one:\n got: %d\nwant: %d", channelTwoPeakLevel.PeakLevel, 0)
	}
}

func TestSoundMeterAudioHandler_NewReceiver(t *testing.T) {
	soundMeter := SoundMeterAudioHandler{}

	receiver := soundMeter.NewReceiver()
	defer receiver.Close()

	var soundMetrics *SoundMetrics
	go func() {
		soundMetrics = <-receiver.Channel
	}()

	expectedMetrics := &SoundMetrics{
		ChannelMetrics: []SoundChannelMetrics{SoundChannelMetrics{PeakLevel: 1}},
	}
	soundMeter.sendMetrics(expectedMetrics)

	if !reflect.DeepEqual(soundMetrics, expectedMetrics) {
		t.Errorf("MetricsOutputChannel should send SoundMetrics")
	}
}

func TestSoundMeterAudioHandler_computeAudio(t *testing.T) {
	soundMeter := SoundMeterAudioHandler{}
	audio := NewAudio(1024, 2)
	audio.SetSample(0, 0, 1.0) // First sample at 1.0

	soundMeter.computeAudio(audio)

	receiver := soundMeter.NewReceiver()
	defer receiver.Close()

	var soundMetrics *SoundMetrics
	go func() {
		soundMetrics = <-receiver.Channel
	}()
	soundMeter.computeAudio(audio)

	if !reflect.DeepEqual(soundMetrics, NewSoundMetrics(audio)) {
		t.Errorf("MetricsOutputChannel should send SoundMetrics created with given audio")
	}

	peak := soundMeter.history.GlobalMetrics().ChannelMetrics[0].PeakLevel
	if peak != 1.0 {
		t.Errorf("computeAudio should update history and its peak level :\n got: %v\nwant: %v", peak, 1.0)
	}

}

func TestSoundChannelMetrics_json(t *testing.T) {
	metrics := SoundChannelMetrics{PeakLevel: 0.333}
	jsonBytes, _ := json.Marshal(metrics)

	expectedJson := []byte("{\"PeakLevel\":0.333}")
	if !bytes.Equal(jsonBytes, expectedJson) {
		t.Errorf("Wrong JSON output:\n got: %s\nwant: %s", jsonBytes, expectedJson)
	}
}

func TestSoundMetrics_json(t *testing.T) {
	metrics := &SoundMetrics{
		ChannelMetrics: []SoundChannelMetrics{SoundChannelMetrics{PeakLevel: 0.333}},
	}

	jsonBytes, _ := json.Marshal(metrics)

	expectedJson := []byte("[{\"PeakLevel\":0.333}]")
	if !bytes.Equal(jsonBytes, expectedJson) {
		t.Errorf("Wrong JSON output:\n got: %s\nwant: %s", jsonBytes, expectedJson)
	}
}

func TestSoundMetricsHistory_GlobalMetrics_Empty(t *testing.T) {
	history := NewSoundMetricsHistory(10, 2)

	globalMetrics := history.GlobalMetrics()

	if globalMetrics.ChannelCount() != history.ChannelCount {
		t.Errorf("GlobalMetrics should have the same channel count as History:\n got: %v\nwant: %v", globalMetrics.ChannelCount(), history.ChannelCount)
	}

	for channel := 0; channel < history.ChannelCount; channel++ {
		if globalMetrics.ChannelMetrics[channel].PeakLevel != 0 {
			t.Errorf("GlobalMetrics of an empty history should have zero PeakLevel :\n got: %v", globalMetrics.ChannelMetrics[channel].PeakLevel)
		}
	}
}

func TestSoundMetricsHistory_GlobalMetrics(t *testing.T) {
	history := NewSoundMetricsHistory(10, 2)

	metrics := &SoundMetrics{
		ChannelMetrics: []SoundChannelMetrics{
			SoundChannelMetrics{PeakLevel: 1},
			SoundChannelMetrics{PeakLevel: 1},
		},
	}
	history.Update(metrics)

	for count := 0; count < 10; count++ {
		metrics := &SoundMetrics{
			ChannelMetrics: []SoundChannelMetrics{
				SoundChannelMetrics{PeakLevel: 0.5},
				SoundChannelMetrics{PeakLevel: 0.5},
			},
		}
		history.Update(metrics)
	}

	globalMetrics := history.GlobalMetrics()
	if globalMetrics.ChannelCount() != history.ChannelCount {
		t.Errorf("GlobalMetrics should have the same channel count as History:\n got: %v\nwant: %v", globalMetrics.ChannelCount(), history.ChannelCount)
	}

	for channel := 0; channel < history.ChannelCount; channel++ {
		if globalMetrics.ChannelMetrics[channel].PeakLevel != 0.5 {
			t.Errorf("GlobalMetrics PeakLevel shoud max PeakLevel in history :\n got: %v", globalMetrics.ChannelMetrics[channel].PeakLevel)
		}
	}
}

func TestSoundMeterAudioHandler_json(t *testing.T) {
	soundMeter := SoundMeterAudioHandler{}

	audio := NewAudio(1024, 2)
	audio.SetSample(0, 0, 1.0) // First sample at 1.0
	soundMeter.computeAudio(audio)

	jsonBytes, _ := json.Marshal(soundMeter)

	expectedJson := []byte("{\"history\":{\"300\":[{\"PeakLevel\":1},{\"PeakLevel\":0}]}}")
	if !bytes.Equal(jsonBytes, expectedJson) {
		t.Errorf("Wrong JSON output:\n got: %s\nwant: %s", jsonBytes, expectedJson)
	}
}
