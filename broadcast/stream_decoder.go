package broadcast

import (
	"io"
)

type StreamDecoder interface {
	SetAudioHandler(audioHandler AudioHandler)
	Init() error

	Read(reader io.Reader) error
	Reset()
}

func NewStreamDecoder(encoding string) StreamDecoder {
	switch encoding {
	case "mp3":
		return &MadDecoder{}
	case "ogg/vorbis":
		return &OggDecoder{
			handler: &VorbisDecoder{},
		}
	}
	return nil
}
