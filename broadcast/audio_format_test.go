package broadcast

import (
	"reflect"
	"testing"
)

func TestParseAudioFormat(t *testing.T) {
	var conditions = []struct {
		definition          string
		expectedAudioFormat AudioFormat
	}{
		{"ogg/vorbis", AudioFormat{Encoding: "ogg/vorbis"}},
		{"mp3", AudioFormat{Encoding: "mp3"}},
		{"aac", AudioFormat{Encoding: "aac"}},
		{"aacp", AudioFormat{Encoding: "aacp"}},
		{"MP3", AudioFormat{Encoding: "mp3"}},
		{"dummy", AudioFormat{}},

		{"ogg/vorbis:vbr", AudioFormat{Encoding: "ogg/vorbis", Mode: "vbr"}},
		{"mp3:cbr", AudioFormat{Encoding: "mp3", Mode: "cbr"}},

		{"ogg/vorbis:vbr(q=10)", AudioFormat{Encoding: "ogg/vorbis", Mode: "vbr", Quality: 1}},
		{"ogg/vorbis:vbr(q=5)", AudioFormat{Encoding: "ogg/vorbis", Mode: "vbr", Quality: 0.5}},
		{"ogg/vorbis:vbr(q=0)", AudioFormat{Encoding: "ogg/vorbis", Mode: "vbr", Quality: 0}},

		{"mp3:cbr(b=96)", AudioFormat{Encoding: "mp3", Mode: "cbr", BitRate: 96000}},
		{"mp3:cbr(b=96,q=5)", AudioFormat{Encoding: "mp3", Mode: "cbr", Quality: 0.5, BitRate: 96000}},
		{"mp3:abr(b=96,q=5)", AudioFormat{Encoding: "mp3", Mode: "abr", Quality: 0.5, BitRate: 96000}},

		{"mp3:vbr(q=5):2", AudioFormat{Encoding: "mp3", Mode: "vbr", Quality: 0.5, ChannelCount: 2}},
		{"ogg/vorbis:vbr(q=5):8", AudioFormat{Encoding: "ogg/vorbis", Mode: "vbr", Quality: 0.5, ChannelCount: 8}},

		{"mp3:vbr(q=5):2:48000", AudioFormat{Encoding: "mp3", Mode: "vbr", Quality: 0.5, ChannelCount: 2, SampleRate: 48000}},
	}

	for _, condition := range conditions {
		audioFormat := ParseAudioFormat(condition.definition)
		if !reflect.DeepEqual(audioFormat, condition.expectedAudioFormat) {
			t.Errorf("Wrong AudioFormat :\n got: %v\nwant: %v", audioFormat, condition.expectedAudioFormat)
		}
	}
}
