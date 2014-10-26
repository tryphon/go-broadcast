package broadcast

import (
	"io"
)

type StreamEncoder interface {
	AudioOut(audio *Audio)
	Init() error
}

func NewStreamEncoder(format AudioFormat, writer io.Writer) StreamEncoder {
	switch {
	case format.Encoding == "mp3":
		return &LameEncoder{
			SampleRate:   int(format.SampleRate),
			ChannelCount: int(format.ChannelCount),
			Quality:      format.Quality,
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
				Quality:      format.Quality,
				ChannelCount: format.ChannelCount,
				SampleRate:   format.SampleRate,
			},
			Writer: writer,
		}
		encoder.Encoder.PacketHandler = &encoder
		return &encoder
	}
	return nil
}
