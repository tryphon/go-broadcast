package broadcast

import ()

type ResizeAudio struct {
	Output       AudioHandler
	SampleCount  int
	ChannelCount int

	pendingAudio       *Audio
	pendingSampleCount int
}

func (resize *ResizeAudio) newPendingAudio() {
	resize.pendingAudio = NewAudio(resize.SampleCount, resize.ChannelCount)
	for channel := 0; channel < resize.ChannelCount; channel++ {
		resize.pendingAudio.SetSamples(channel, make([]float32, resize.SampleCount))
	}
	resize.pendingSampleCount = 0
}

func (resize *ResizeAudio) pendingCapacity() int {
	return resize.SampleCount - resize.pendingSampleCount
}

func (resize *ResizeAudio) maxSampleCount(sampleCount int) int {
	if sampleCount <= resize.pendingCapacity() {
		return sampleCount
	} else {
		return resize.pendingCapacity()
	}
}

func (resize *ResizeAudio) AudioOut(audio *Audio) {
	if resize.ChannelCount == 0 {
		resize.ChannelCount = audio.ChannelCount()
	}

	if resize.pendingAudio == nil {
		resize.newPendingAudio()
	}

	for consumedSampleCount := 0; consumedSampleCount < audio.SampleCount(); {
		firstSampleSlice := consumedSampleCount

		sampleCount := resize.maxSampleCount(audio.SampleCount() - consumedSampleCount)
		lastSampleSlice := firstSampleSlice + sampleCount

		for channel := 0; channel < resize.ChannelCount; channel++ {
			copy(resize.pendingAudio.Samples(channel)[resize.pendingSampleCount:], audio.Samples(channel)[firstSampleSlice:lastSampleSlice])
		}

		resize.pendingSampleCount += sampleCount
		consumedSampleCount += sampleCount

		if resize.pendingCapacity() == 0 {
			resize.Output.AudioOut(resize.pendingAudio)
			resize.newPendingAudio()
		}
	}
}
