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

		modeRegExp := regexp.MustCompile("^(cbr|abr|vbr)(\\((.*)\\))?$")
		if modeRegExp.MatchString(modePart) {
			match := modeRegExp.FindStringSubmatch(modePart)
			audio.Mode = match[1]

			if len(match) > 3 {
				attributeRegExp := regexp.MustCompile("^([bq])=([0-9]+)$")
				for _, attribute := range strings.Split(match[3], ",") {
					if attributeRegExp.MatchString(attribute) {
						match := attributeRegExp.FindStringSubmatch(attribute)
						name := match[1]
						value := match[2]

						switch {
						case name == "b":
							userBitRate, _ := strconv.Atoi(value)
							audio.BitRate = userBitRate * 1000
						case name == "q":
							userQuality, _ := strconv.Atoi(value)
							if userQuality >= 0 && userQuality <= 10 {
								audio.Quality = float32(userQuality) / 10
							}
						}
					}
				}
			}
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

func FindEncodingByContentType(contentType string) string {
	switch contentType {
	case "audio/mpeg":
		return "mp3"
	case "application/ogg":
		return "ogg/vorbis"
	case "audio/aac":
		return "aac"
	case "audio/aacp":
		return "aacp"
	}
	return ""
}
