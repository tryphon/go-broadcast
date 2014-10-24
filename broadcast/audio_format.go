package broadcast

import (
	"regexp"
	"strconv"
	"strings"
)

type AudioFormat struct {
	Encoding string
	Mode     string

	BitRate int
	Quality float32

	ChannelCount int
	SampleRate   int
}

// ogg/vorbis:vbr(q=5):2:44100
// mp3:cbr(b=128):2:44100

func ParseAudioFormat(definition string) AudioFormat {
	parts := strings.Split(strings.ToLower(definition), ":")

	audio := AudioFormat{}
	if len(parts) >= 1 {
		encoding := parts[0]
		if regexp.MustCompile("^ogg/vorbis|mp3|aac|aacp$").MatchString(encoding) {
			audio.Encoding = encoding
		}
	}

	if len(parts) >= 2 {
		modePart := parts[1]

		vbrRegExp := regexp.MustCompile("^vbr\\(q=([0-9]+)\\)$")
		cbrRegExp := regexp.MustCompile("^cbr\\(b=([0-9]+)\\)$")

		switch {
		case vbrRegExp.MatchString(modePart):
			audio.Mode = "vbr"
			userQuality, _ := strconv.Atoi(vbrRegExp.FindStringSubmatch(modePart)[1])
			if userQuality >= 0 && userQuality <= 10 {
				audio.Quality = float32(userQuality) / 10
			}
		case cbrRegExp.MatchString(modePart):
			audio.Mode = "cbr"
			userBitRate, _ := strconv.Atoi(cbrRegExp.FindStringSubmatch(modePart)[1])
			audio.BitRate = userBitRate * 1000
		}
	}

	if len(parts) >= 3 {
		audio.ChannelCount, _ = strconv.Atoi(parts[2])
	}

	if len(parts) >= 4 {
		audio.SampleRate, _ = strconv.Atoi(parts[3])
	}

	return audio
}

func (format *AudioFormat) ContentType() string {
	switch {
	case format.Encoding == "mp3":
		return "audio/mpeg"
	case strings.HasPrefix(format.Encoding, "ogg/"):
		return "application/ogg"
	case format.Encoding == "aac":
		return "audio/aac"
	case format.Encoding == "aacp":
		return "audio/aacp"
	}
	return ""
}
