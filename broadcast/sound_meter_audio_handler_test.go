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

	if len(soundMetrics) != audio.ChannelCount() {
		t.Errorf("Should have a SoundMetrics for each channel:\n got: %d\nwant: %d", len(soundMetrics), audio.ChannelCount())
	}

	channelOnePeakLevel := soundMetrics[0]
	if channelOnePeakLevel.PeakLevel != 1 {
		t.Errorf("Wrong peak level on channel one:\n got: %d\nwant: %d", channelOnePeakLevel.PeakLevel, 1)
	}

	channelTwoPeakLevel := soundMetrics[1]
	if channelTwoPeakLevel.PeakLevel != 0 {
		t.Errorf("Wrong peak level on channel one:\n got: %d\nwant: %d", channelTwoPeakLevel.PeakLevel, 0)
	}
}

func TestSoundMeterAudioHandler_NewReceiver(t *testing.T) {
	soundMeter := SoundMeterAudioHandler{}

	receiver := soundMeter.NewReceiver()
	defer receiver.Close()

	var soundMetrics []SoundMetrics
	go func() {
		soundMetrics = <-receiver.Channel
	}()

	expectedMetrics := []SoundMetrics{SoundMetrics{PeakLevel: 1}}
	soundMeter.sendMetrics(expectedMetrics)

	if !reflect.DeepEqual(soundMetrics, expectedMetrics) {
		t.Errorf("MetricsOutputChannel should send SoundMetrics")
	}
}

func TestSoundMeterAudioHandler_computeAudio(t *testing.T) {
	soundMeter := SoundMeterAudioHandler{}
	audio := NewAudio(1024, 2)

	soundMeter.computeAudio(audio)

	receiver := soundMeter.NewReceiver()
	defer receiver.Close()

	var soundMetrics []SoundMetrics
	go func() {
		soundMetrics = <-receiver.Channel
	}()
	soundMeter.computeAudio(audio)

	if !reflect.DeepEqual(soundMetrics, NewSoundMetrics(audio)) {
		t.Errorf("MetricsOutputChannel should send SoundMetrics created with given audio")
	}
}

func TestSoundMetrics_json(t *testing.T) {
	metrics := SoundMetrics{PeakLevel: 0.333}
	jsonBytes, _ := json.Marshal(metrics)

	expectedJson := []byte("{\"PeakLevel\":0.333}")
	if !bytes.Equal(jsonBytes, expectedJson) {
		t.Errorf("Wrong JSON output:\n got: %s\nwant: %s", jsonBytes, expectedJson)
	}
}

func TestSoundMetrics_SliceJson(t *testing.T) {
	metrics := []SoundMetrics{SoundMetrics{PeakLevel: 0.333}}
	jsonBytes, _ := json.Marshal(metrics)

	expectedJson := []byte("[{\"PeakLevel\":0.333}]")
	if !bytes.Equal(jsonBytes, expectedJson) {
		t.Errorf("Wrong JSON output:\n got: %s\nwant: %s", jsonBytes, expectedJson)
	}
}
