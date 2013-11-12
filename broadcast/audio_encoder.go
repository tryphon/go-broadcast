package broadcast

type AudioEncoder interface {
	Encode(audio *Audio) ([]byte, error)
	Init() error
}

type AudioDecoder interface {
	Decode([]byte) (*Audio, error)
	Init() error
}
