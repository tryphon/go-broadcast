package broadcast

import (
	"container/list"
	"container/ring"
	"encoding/json"
	"math"
)

type SoundMeterAudioHandler struct {
	Output AudioHandler

	resizeAudio *ResizeAudio
	receivers   *list.List
	history     *SoundMetricsHistory
}

type SoundChannelMetrics struct {
	PeakLevel float32
}

type SoundMetrics struct {
	ChannelMetrics []SoundChannelMetrics
}

func NewSoundMetrics(audio *Audio) *SoundMetrics {
	soundMetrics := &SoundMetrics{
		ChannelMetrics: make([]SoundChannelMetrics, audio.ChannelCount()),
	}

	for channel := 0; channel < audio.ChannelCount(); channel++ {
		var peak float64 = 0
		for _, sample := range audio.Samples(channel) {
			value := math.Abs(float64(sample))
			if value > peak {
				peak = value
			}
		}
		soundMetrics.ChannelMetrics[channel].PeakLevel = float32(peak)
	}

	return soundMetrics
}

func (metrics *SoundMetrics) ChannelCount() int {
	return len(metrics.ChannelMetrics)
}

func (metrics *SoundMetrics) MarshalJSON() ([]byte, error) {
	return json.Marshal(metrics.ChannelMetrics)
}

func (soundMeter *SoundMeterAudioHandler) AudioOut(audio *Audio) {
	if soundMeter.resizeAudio == nil {
		audioHandler := AudioHandlerFunc(soundMeter.computeAudio)
		soundMeter.resizeAudio = &ResizeAudio{SampleCount: 4410, ChannelCount: 2, Output: audioHandler}
	}

	soundMeter.resizeAudio.AudioOut(audio)
	soundMeter.Output.AudioOut(audio)
}

func (soundMeter *SoundMeterAudioHandler) NewReceiver() *SoundMetricsReceiver {
	receiver := &SoundMetricsReceiver{
		Channel:    make(chan *SoundMetrics),
		soundMeter: soundMeter,
	}
	if soundMeter.receivers == nil {
		soundMeter.receivers = list.New()
	}
	Log.Debugf("New SoundMetrics Receiver")
	soundMeter.receivers.PushFront(receiver)
	return receiver
}

func (soundMeter *SoundMeterAudioHandler) closeReceiver(receiver *SoundMetricsReceiver) {
	if soundMeter.receivers == nil {
		return
	}

	for element := soundMeter.receivers.Front(); element != nil; element = element.Next() {
		if element.Value == receiver {
			Log.Debugf("Close SoundMetrics Receiver")
			soundMeter.receivers.Remove(element)
			return
		}
	}
}

func (soundMeter *SoundMeterAudioHandler) computeAudio(audio *Audio) {
	if soundMeter.history == nil {
		soundMeter.history = NewSoundMetricsHistory(44100/4410*300, 2)
	}

	soundMetrics := NewSoundMetrics(audio)

	soundMeter.history.Update(soundMetrics)
	soundMeter.sendMetrics(soundMetrics)
}

func (soundMeter *SoundMeterAudioHandler) sendMetrics(metrics *SoundMetrics) {
	if soundMeter.receivers == nil {
		return
	}

	for element := soundMeter.receivers.Front(); element != nil; element = element.Next() {
		receiver := element.Value.(*SoundMetricsReceiver)
		receiver.Channel <- metrics
	}
}

func (soundMeter SoundMeterAudioHandler) MarshalJSON() ([]byte, error) {
	data := map[string]map[string]interface{}{
		"history": map[string]interface{}{
			"300": soundMeter.history.GlobalMetrics(),
		},
	}
	return json.Marshal(data)
}

type SoundMetricsReceiver struct {
	Channel    chan *SoundMetrics
	soundMeter *SoundMeterAudioHandler
}

func (receiver *SoundMetricsReceiver) Close() {
	receiver.soundMeter.closeReceiver(receiver)
}

type SoundMetricsHistory struct {
	ChannelCount int
	metrics      *ring.Ring
}

func NewSoundMetricsHistory(length int, channelCount int) *SoundMetricsHistory {
	return &SoundMetricsHistory{
		ChannelCount: channelCount,
		metrics:      ring.New(length),
	}
}

func (history *SoundMetricsHistory) Update(metrics *SoundMetrics) {
	history.metrics.Value = metrics
	history.metrics = history.metrics.Next()
}

func (history *SoundMetricsHistory) GlobalMetrics() *SoundMetrics {
	globalMetrics := &SoundMetrics{
		ChannelMetrics: make([]SoundChannelMetrics, history.ChannelCount),
	}

	history.metrics.Do(func(value interface{}) {
		if value == nil {
			return
		}

		metrics := value.(*SoundMetrics)

		for channel := 0; channel < history.ChannelCount; channel++ {
			globalChannelMetrics := globalMetrics.ChannelMetrics[channel]
			channelMetrics := metrics.ChannelMetrics[channel]

			if channelMetrics.PeakLevel > globalChannelMetrics.PeakLevel {
				globalMetrics.ChannelMetrics[channel].PeakLevel = channelMetrics.PeakLevel
			}
		}
	})

	return globalMetrics
}
