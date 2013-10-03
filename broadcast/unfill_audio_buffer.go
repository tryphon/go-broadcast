package broadcast

import (
	"time"
)

type UnfillAudioBuffer struct {
	Buffer AudioBuffer

	MaxSampleCount    uint32
	UnfillSampleCount uint32
}

func (pseudoBuffer *UnfillAudioBuffer) full() bool {
	return pseudoBuffer.MaxSampleCount != 0 &&
		pseudoBuffer.UnfillSampleCount != 0 &&
		pseudoBuffer.Buffer.SampleCount() > pseudoBuffer.MaxSampleCount
}

func (pseudoBuffer *UnfillAudioBuffer) unfill() {
	initialSampleCount := pseudoBuffer.Buffer.SampleCount()
	targetSampleCount := pseudoBuffer.MaxSampleCount - pseudoBuffer.UnfillSampleCount

	for pseudoBuffer.Buffer.SampleCount() > targetSampleCount {
		pseudoBuffer.Buffer.Read()
	}
	Log.Debugf("%v Unfill duration : %d samples", time.Now(), initialSampleCount-pseudoBuffer.Buffer.SampleCount())
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
