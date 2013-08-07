package broadcast

type AudioBuffer interface {
	AudioOut(audio *Audio)
	Read() *Audio
	SampleCount() uint32
}

// const maxAudioBufferSize uint32 = 4096

// type AudioBuffer struct {
// 	audios [maxAudioBufferSize]*Audio

// 	sampleCount uint32

// 	nextFreeIndex uint32
// 	readIndex     uint32

// 	MinSampleCount uint32
// 	refill         bool

// 	MaxSampleCount uint32
// 	UnfillSampleCount uint32

// 	full bool

// 	mutex sync.Mutex
// }

// func NewAudioBuffer() *AudioBuffer {
// 	return &AudioBuffer{refill: true}
// }

// func (buffer *AudioBuffer) Empty() bool {
// 	return !buffer.full && buffer.nextFreeIndex == buffer.readIndex
// }

// func (buffer *AudioBuffer) Full() bool {
// 	return buffer.full
// }

// func (buffer *AudioBuffer) Refill() bool {
// 	if buffer.refill && buffer.sampleCount > buffer.MinSampleCount {
// 		buffer.refill = false
// 	}

// 	return buffer.refill
// }

// func (buffer *AudioBuffer) SampleCount() uint32 {
// 	return buffer.sampleCount
// }

// func (buffer *AudioBuffer) unfill() {
// 	if buffer.Full() {
// 		// Buffer is full, moving reader to read oldest audio
// 		buffer.Read()
// 	}

// 	if buffer.MaxSampleCount != 0 && buffer.UnfillSampleCount != 0 && buffer.sampleCount > buffer.MaxSampleCount {
// 		initialSampleCount := buffer.sampleCount
// 		for buffer.sampleCount > buffer.MaxSampleCount - buffer.UnfillSampleCount {
// 			buffer.Read()
// 		}
// 		Log.Debugf("%v Unfill duration : %d samples\n", time.Now(), initialSampleCount - buffer.sampleCount)
// 	}
// }

// func (buffer *AudioBuffer) AudioOut(audio *Audio) {
// 	// buffer.Dump()

// 	buffer.unfill()

// 	buffer.mutex.Lock()
// 	defer buffer.mutex.Unlock()

// 	buffer.audios[buffer.nextFreeIndex] = audio
// 	buffer.sampleCount += uint32(audio.SampleCount())

// 	buffer.nextFreeIndex = (buffer.nextFreeIndex + 1) % maxAudioBufferSize
// 	buffer.full = (buffer.nextFreeIndex == buffer.readIndex)
// }

// func (buffer *AudioBuffer) Read() (audio *Audio) {
// 	// buffer.Dump()

// 	buffer.mutex.Lock()
// 	defer buffer.mutex.Unlock()

// 	if buffer.Empty() {
// 		buffer.refill = true
// 		return nil
// 	}

// 	if buffer.Refill() {
// 		return nil
// 	}

// 	audio = buffer.audios[buffer.readIndex]
// 	buffer.sampleCount -= uint32(audio.SampleCount())
// 	buffer.audios[buffer.readIndex] = nil

// 	buffer.readIndex = (buffer.readIndex + 1) % maxAudioBufferSize
// 	buffer.full = false

// 	return
// }

// func (buffer *AudioBuffer) Dump() {
// 	Log.Debugf("%v SampleCount: %d, NextFreeIndex: %d, ReadIndex: %d, Full: %v, Refill: %v\n", time.Now(), buffer.sampleCount, buffer.nextFreeIndex, buffer.readIndex, buffer.full, buffer.Refill())
// }
