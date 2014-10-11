package broadcast

type StreamEncoder interface {
	AudioOut(audio *Audio)
	Init() error
}
