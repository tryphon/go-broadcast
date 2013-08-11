package broadcast

import (

)

type ResizeAudio struct {
	Output AudioHandler
	SampleCount int
	ChannelCount int

	pendingAudio *Audio
	pendingSampleCount int
}

func (resize *ResizeAudio) newPendingAudio() {
	resize.pendingAudio = NewAudio(resize.SampleCount, resize.ChannelCount)
	for channel := 0; channel < resize.ChannelCount; channel++ {
		resize.pendingAudio.SetSamples(channel, make([]float32, resize.SampleCount))
	}
	resize.pendingSampleCount = 0
}

func (resize *ResizeAudio) pendingCapacity() (int) {
	return resize.SampleCount - resize.pendingSampleCount
}

func (resize *ResizeAudio) maxSampleCount(sampleCount int) (int) {
	if sampleCount <= resize.pendingCapacity() {
		return sampleCount
	} else {
		return resize.pendingCapacity()
	}
}

func (resize *ResizeAudio) AudioOut(audio *Audio) {
	if resize.pendingAudio == nil {
		resize.newPendingAudio()
	}

	for consumedSampleCount := 0; consumedSampleCount < audio.SampleCount(); {
		firstSampleSlice := consumedSampleCount

		sampleCount := resize.maxSampleCount(audio.SampleCount() - consumedSampleCount)
		lastSampleSlice := firstSampleSlice + sampleCount

		// Log.Debugf("copy %d:%d to %d:", firstSampleSlice, lastSampleSlice, resize.pendingSampleCount)

		for channel := 0; channel < resize.ChannelCount; channel++ {
			copy(resize.pendingAudio.Samples(channel)[resize.pendingSampleCount:], audio.Samples(channel)[firstSampleSlice:lastSampleSlice])
		}

		resize.pendingSampleCount += sampleCount
		consumedSampleCount += sampleCount

		// Log.Debugf("counters %d:%d", resize.pendingSampleCount, consumedSampleCount)

		if (resize.pendingCapacity() == 0) {
			// Log.Debugf("send Audio : %v", len(resize.pendingAudio.Samples(0)))
			resize.Output.AudioOut(resize.pendingAudio)
			resize.newPendingAudio()
		}
	}

	// sliceCount := audio.SampleCount() / resize.SampleCount

	// for slice := 0; slice < sliceCount; slice++ {
	// 	firstSampleSlice := slice * resize.SampleCount
	// 	lastSampleSlice := (slice + 1) * resize.SampleCount

	// 	resizedAudio := NewAudio(resize.SampleCount, resize.ChannelCount)
	// 	for channel := 0; channel < resize.ChannelCount; channel++ {
	// 		resizedAudio.SetSamples(channel, audio.Samples(channel)[firstSampleSlice:lastSampleSlice])
	// 	}
	// 	resize.Output.AudioOut(resizedAudio)
	// }
}
