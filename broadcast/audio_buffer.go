package broadcast

type AudioBuffer interface {
	AudioOut(audio *Audio)
	Read() *Audio
	SampleCount() uint32
}
