package broadcast

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type AdjustAudioBuffer struct {
	Buffer AudioBuffer

	LimitSampleCount     uint32
	ThresholdSampleCount uint32
}

func (pseudoBuffer *AdjustAudioBuffer) fillRate() float64 {
	sampleCount := pseudoBuffer.SampleCount()

	if pseudoBuffer.LimitSampleCount == 0 && pseudoBuffer.ThresholdSampleCount == 0 {
		return 0
	}

	rawRate := (float64(sampleCount) - float64(pseudoBuffer.ThresholdSampleCount)) / (float64(pseudoBuffer.LimitSampleCount) - float64(pseudoBuffer.ThresholdSampleCount))

	// fmt.Printf("SampleCount : %d, ThresholdSampleCount: %d, LimitSampleCount: %d, RawRate: %f\n", sampleCount, pseudoBuffer.ThresholdSampleCount, pseudoBuffer.LimitSampleCount, rawRate)
	return math.Min(1, math.Max(0, rawRate))
}

func (pseudoBuffer *AdjustAudioBuffer) adjustmentFactor() int {
	delta := pseudoBuffer.LimitSampleCount - pseudoBuffer.ThresholdSampleCount
	if delta > 0 {
		return 1
	}
	if delta < 0 {
		return -1
	}
	return 0
}

func (pseudoBuffer *AdjustAudioBuffer) addAudio() bool {
	return pseudoBuffer.LimitSampleCount < pseudoBuffer.ThresholdSampleCount
}

func (pseudoBuffer *AdjustAudioBuffer) removeAudio() bool {
	return !pseudoBuffer.addAudio()
}

func (pseudoBuffer *AdjustAudioBuffer) AudioOut(audio *Audio) {
	pseudoBuffer.Buffer.AudioOut(audio)

	if pseudoBuffer.removeAudio() && pseudoBuffer.adjust() {
		pseudoBuffer.logAdjustment(pseudoBuffer.Buffer.Read())
	}
}

func (pseudoBuffer *AdjustAudioBuffer) adjust() bool {
	probability := rand.Float64()
	value := -math.Log(1-pseudoBuffer.fillRate()) / 5

	result := value > probability
	if result {
		fmt.Printf("Fill Rate : %f, Value: %f, Probability: %f\n", pseudoBuffer.fillRate(), value, probability)
	}
	return result
}

func (pseudoBuffer *AdjustAudioBuffer) logAdjustment(audio *Audio) *Audio {
	if audio != nil {
		fmt.Printf("%v Adjustment : %d samples\n", time.Now(), pseudoBuffer.adjustmentFactor()*audio.SampleCount())
	}
	return audio
}

func (pseudoBuffer *AdjustAudioBuffer) Read() (audio *Audio) {
	if pseudoBuffer.addAudio() && pseudoBuffer.adjust() {
		return pseudoBuffer.logAdjustment(&Audio{sampleCount: 512, channelCount: 2})
	} else {
		return pseudoBuffer.Buffer.Read()
	}
}

func (pseudoBuffer *AdjustAudioBuffer) SampleCount() uint32 {
	return pseudoBuffer.Buffer.SampleCount()
}
