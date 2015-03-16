package broadcast

import (
	"io"
)

type StreamEncoder interface {
	AudioOut(audio *Audio)
	Init() error
	Close()
}

func NewStreamEncoder(format AudioFormat, writer io.Writer) StreamEncoder {
	switch {
	case format.Encoding == "mp3":
		return &LameEncoder{
			SampleRate:   int(format.SampleRate),
			ChannelCount: int(format.ChannelCount),
			Mode:         format.Mode,
			Quality:      format.Quality,
			BitRate:      format.BitRate,
			Writer:       writer,
		}
	case format.Encoding == "aacp":
		return &AACPEncoder{
			SampleRate:   int(format.SampleRate),
			ChannelCount: int(format.ChannelCount),
			BitRate:      format.BitRate,
			Writer:       writer,
		}
	case format.Encoding == "aac":
		return &AACEncoder{
			SampleRate:   int(format.SampleRate),
			ChannelCount: int(format.ChannelCount),
			BitRate:      format.BitRate,
			Writer:       writer,
		}
	case format.Encoding == "ogg/vorbis":
		encoder := OggEncoder{
			Encoder: VorbisEncoder{
				Mode:         format.Mode,
				Quality:      format.Quality,
				BitRate:      format.BitRate,
				ChannelCount: format.ChannelCount,
				SampleRate:   format.SampleRate,
			},
			Writer: writer,
		}
		return &encoder
	}
	return nil
}
