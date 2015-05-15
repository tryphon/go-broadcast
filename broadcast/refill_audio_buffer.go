package broadcast

type RefillAudioBuffer struct {
	Buffer AudioBuffer

	MinSampleCount uint32
	readable       bool
}

func (pseudoBuffer *RefillAudioBuffer) AudioOut(audio *Audio) {
	pseudoBuffer.Buffer.AudioOut(audio)

	if !pseudoBuffer.readable &&
		pseudoBuffer.SampleCount() > pseudoBuffer.MinSampleCount {
		Log.Debugf("RefillAudioBuffer readeable")
		pseudoBuffer.readable = true
	}
}

func (pseudoBuffer *RefillAudioBuffer) Read() (audio *Audio) {
	if !pseudoBuffer.readable {
		return nil
	}

	if pseudoBuffer.SampleCount() == 0 {
		Log.Debugf("RefillAudioBuffer unreadeable")
		pseudoBuffer.readable = false
		return nil
	}

	// Log.Debugf("SampleCount: %d", pseudoBuffer.SampleCount())

	audio = pseudoBuffer.Buffer.Read()

	if audio == nil {
		Log.Debugf("RefillAudioBuffer unreadeable")
		pseudoBuffer.readable = false
	}

	return audio
}

func (pseudoBuffer *RefillAudioBuffer) SampleCount() uint32 {
	return pseudoBuffer.Buffer.SampleCount()
}
