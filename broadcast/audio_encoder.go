package broadcast

type AudioEncoder interface {
	Encode(audio *Audio) ([]byte, error)
}

type AudioDecoder interface {
	Decode([]byte) (*Audio, error)
}
