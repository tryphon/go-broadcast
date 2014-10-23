package broadcast

type UnfillAudioBuffer struct {
	Buffer AudioBuffer

	MaxSampleCount    uint32
	UnfillSampleCount uint32

	Metrics *LocalMetrics
}

func (buffer *UnfillAudioBuffer) metrics() *LocalMetrics {
	if buffer.Metrics == nil {
		buffer.Metrics = &LocalMetrics{}
	}
	return buffer.Metrics
}

func (pseudoBuffer *UnfillAudioBuffer) full() bool {
	return pseudoBuffer.MaxSampleCount != 0 &&
		pseudoBuffer.UnfillSampleCount != 0 &&
		pseudoBuffer.Buffer.SampleCount() > pseudoBuffer.MaxSampleCount
}

func (pseudoBuffer *UnfillAudioBuffer) unfill() {
	targetSampleCount := pseudoBuffer.MaxSampleCount - pseudoBuffer.UnfillSampleCount

	for pseudoBuffer.Buffer.SampleCount() > targetSampleCount {
		unfillAudio := pseudoBuffer.Buffer.Read()
		pseudoBuffer.metrics().Counter("buffer.Unfill").Inc(int64(unfillAudio.SampleCount()))
	}
}

func (pseudoBuffer *UnfillAudioBuffer) AudioOut(audio *Audio) {
	pseudoBuffer.Buffer.AudioOut(audio)

	if pseudoBuffer.full() {
		pseudoBuffer.unfill()
	}
}

func (pseudoBuffer *UnfillAudioBuffer) Read() (audio *Audio) {
	return pseudoBuffer.Buffer.Read()
}

func (pseudoBuffer *UnfillAudioBuffer) SampleCount() uint32 {
	return pseudoBuffer.Buffer.SampleCount()
}
