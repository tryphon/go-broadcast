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
		pseudoBuffer.readable = true
	}
}

func (pseudoBuffer *RefillAudioBuffer) Read() (audio *Audio) {
	if !pseudoBuffer.readable {
		return nil
	}

	audio = pseudoBuffer.Buffer.Read()

	if audio == nil {
		pseudoBuffer.readable = false
	}

	return audio
}

func (pseudoBuffer *RefillAudioBuffer) SampleCount() uint32 {
	return pseudoBuffer.Buffer.SampleCount()
}
