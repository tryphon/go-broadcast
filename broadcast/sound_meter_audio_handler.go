package broadcast

import (
	"container/list"
	"math"
)

type SoundMeterAudioHandler struct {
	Output      AudioHandler
	resizeAudio *ResizeAudio
	receivers   *list.List
}

type SoundMetrics struct {
	PeakLevel float32
}

func NewSoundMetrics(audio *Audio) []SoundMetrics {
	soundMetrics := make([]SoundMetrics, audio.ChannelCount())

	for channel := 0; channel < audio.ChannelCount(); channel++ {
		var peak float64 = 0
		for _, sample := range audio.Samples(channel) {
			value := math.Abs(float64(sample))
			if value > peak {
				peak = value
			}
		}
		soundMetrics[channel] = SoundMetrics{PeakLevel: float32(peak)}
	}

	return soundMetrics
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
		Channel:    make(chan []SoundMetrics),
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
	if soundMeter.receivers == nil {
		return
	}
	if soundMeter.receivers.Len() == 0 {
		return
	}

	soundMetrics := NewSoundMetrics(audio)
	soundMeter.sendMetrics(soundMetrics)
}

func (soundMeter *SoundMeterAudioHandler) sendMetrics(metrics []SoundMetrics) {
	if soundMeter.receivers == nil {
		return
	}

	for element := soundMeter.receivers.Front(); element != nil; element = element.Next() {
		receiver := element.Value.(*SoundMetricsReceiver)
		receiver.Channel <- metrics
	}
}

type SoundMetricsReceiver struct {
	Channel    chan []SoundMetrics
	soundMeter *SoundMeterAudioHandler
}

func (receiver *SoundMetricsReceiver) Close() {
	receiver.soundMeter.closeReceiver(receiver)
}
