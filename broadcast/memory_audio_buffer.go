package broadcast

import ()

const maxAudioBufferSize uint32 = 4096

type MemoryAudioBuffer struct {
	audios [maxAudioBufferSize]*Audio

	sampleCount uint32

	nextFreeIndex uint32
	readIndex     uint32

	full    bool
	Metrics *LocalMetrics
}

func (buffer *MemoryAudioBuffer) Empty() bool {
	return !buffer.full && buffer.nextFreeIndex == buffer.readIndex
}

func (buffer *MemoryAudioBuffer) Full() bool {
	return buffer.full
}

func (buffer *MemoryAudioBuffer) SampleCount() uint32 {
	return buffer.sampleCount
}

func (buffer *MemoryAudioBuffer) metrics() *LocalMetrics {
	if buffer.Metrics == nil {
		buffer.Metrics = &LocalMetrics{}
	}
	return buffer.Metrics
}

func (buffer *MemoryAudioBuffer) changeSampleCount(delta int) {
	buffer.sampleCount += uint32(delta)
	buffer.metrics().Gauge("buffer.Size").Update(int64(buffer.sampleCount))
	buffer.metrics().Histogram("buffer.SizeHistory").Update(int64(buffer.sampleCount))
}

func (buffer *MemoryAudioBuffer) AudioOut(audio *Audio) {
	if buffer.Full() {
		// Buffer is full, moving reader to read oldest audio
		buffer.Read()
	}

	buffer.audios[buffer.nextFreeIndex] = audio
	buffer.changeSampleCount(audio.SampleCount())

	buffer.nextFreeIndex = (buffer.nextFreeIndex + 1) % maxAudioBufferSize
	buffer.full = (buffer.nextFreeIndex == buffer.readIndex)
}

func (buffer *MemoryAudioBuffer) Read() (audio *Audio) {
	if buffer.Empty() {
		return nil
	}

	audio = buffer.audios[buffer.readIndex]
	buffer.changeSampleCount(-audio.SampleCount())
	buffer.audios[buffer.readIndex] = nil

	buffer.readIndex = (buffer.readIndex + 1) % maxAudioBufferSize
	buffer.full = false

	return
}
