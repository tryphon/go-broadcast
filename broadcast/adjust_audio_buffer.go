package broadcast

import (
	metrics "github.com/rcrowley/go-metrics"
	"math"
	"math/rand"
)

type AdjustAudioBuffer struct {
	Buffer       AudioBuffer
	ChannelCount int

	LimitSampleCount     uint32
	ThresholdSampleCount uint32
}

func (pseudoBuffer *AdjustAudioBuffer) fillRate() float64 {
	sampleCount := pseudoBuffer.SampleCount()

	if pseudoBuffer.LimitSampleCount == 0 && pseudoBuffer.ThresholdSampleCount == 0 {
		return 0
	}

	rawRate := (float64(sampleCount) - float64(pseudoBuffer.ThresholdSampleCount)) / (float64(pseudoBuffer.LimitSampleCount) - float64(pseudoBuffer.ThresholdSampleCount))

	// Log.Debugf("SampleCount : %d, ThresholdSampleCount: %d, LimitSampleCount: %d, RawRate: %f", sampleCount, pseudoBuffer.ThresholdSampleCount, pseudoBuffer.LimitSampleCount, rawRate)
	return math.Min(1, math.Max(0, rawRate))
}

func (pseudoBuffer *AdjustAudioBuffer) adjustmentFactor() int {
	delta := int(pseudoBuffer.LimitSampleCount) - int(pseudoBuffer.ThresholdSampleCount)
	if delta > 0 {
		return -1
	}
	if delta < 0 {
		return +1
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
	if pseudoBuffer.ChannelCount == 0 {
		pseudoBuffer.ChannelCount = audio.ChannelCount()
	}

	pseudoBuffer.Buffer.AudioOut(audio)

	pseudoBuffer.adjustmentCounter()
	if pseudoBuffer.removeAudio() && pseudoBuffer.adjust() {
		pseudoBuffer.logAdjustment(pseudoBuffer.Buffer.Read())
	}
}

func (pseudoBuffer *AdjustAudioBuffer) adjust() bool {
	probability := rand.Float64()
	value := -math.Log(1-pseudoBuffer.fillRate()) / 5

	result := value > probability
	// if result {
	// 	Log.Debugf("Fill Rate : %f, Value: %f, Probability: %f", pseudoBuffer.fillRate(), value, probability)
	// }
	return result
}

func (pseudoBuffer *AdjustAudioBuffer) metricName() string {
	if pseudoBuffer.adjustmentFactor() > 0 {
		return "buffer.LowAdjustement"
	} else {
		return "buffer.HighAdjustement"
	}
}

func (pseudoBuffer *AdjustAudioBuffer) adjustmentCounter() metrics.Counter {
	return metrics.GetOrRegisterCounter(pseudoBuffer.metricName(), nil)
}

func (pseudoBuffer *AdjustAudioBuffer) logAdjustment(audio *Audio) *Audio {
	if audio != nil {
		pseudoBuffer.adjustmentCounter().Inc(int64(audio.SampleCount()))
	}
	return audio
}

func (pseudoBuffer *AdjustAudioBuffer) Read() (audio *Audio) {
	pseudoBuffer.adjustmentCounter()

	if pseudoBuffer.addAudio() && pseudoBuffer.adjust() {
		return pseudoBuffer.logAdjustment(NewAudio(1024, pseudoBuffer.ChannelCount))
	} else {
		return pseudoBuffer.Buffer.Read()
	}
}

func (pseudoBuffer *AdjustAudioBuffer) SampleCount() uint32 {
	return pseudoBuffer.Buffer.SampleCount()
}
