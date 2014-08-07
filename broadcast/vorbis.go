package broadcast

import (
	"errors"
	metrics "github.com/tryphon/go-metrics"
	ogg "github.com/tryphon/go-ogg"
	vorbis "github.com/tryphon/go-vorbis"
	"github.com/tryphon/go-vorbis/vorbisenc"
)

type VorbisDecoder struct {
	streamStatus string

	vi vorbis.Info     // struct that stores all the static vorbis bitstream settings
	vc vorbis.Comment  // struct that stores all the user comments
	vd vorbis.DspState // central working state for the packet PCM decoder
	vb vorbis.Block    // local working space for packet PCM decode

	audioHandler AudioHandler
}

func (decoder *VorbisDecoder) SetAudioHandler(audioHandler AudioHandler) {
	decoder.audioHandler = audioHandler
}

func (decoder *VorbisDecoder) Reset() {
	decoder.vi.Clear()
	decoder.vc.Clear()
	decoder.vd.Clear()
	decoder.vb.Clear()

	decoder.streamStatus = ""
}

func (decoder *VorbisDecoder) NewStream(serialNo int32) {
	decoder.vi.Init()
	decoder.vc.Init()
}

func (decoder *VorbisDecoder) PacketOut(packet *ogg.Packet) (result bool) {
	if decoder.streamStatus == "" {
		decoder.streamStatus = "vorbis_init_info"
	}

	switch decoder.streamStatus {
	case "vorbis_init_info", "vorbis_init_comments", "vorbis_init_codebooks":
		Log.Debugf("Init vorbis header %s.", decoder.streamStatus)

		if vorbis.SynthesisHeaderIn(&decoder.vi, &decoder.vc, packet) < 0 {
			Log.Printf("This Ogg bitstream does not contain Vorbis audio data.\n")
			return false
		}

		switch decoder.streamStatus {
		case "vorbis_init_info":
			Log.Debugf("Bitstream is %d channel, %dHz", decoder.vi.Channels(), decoder.vi.Rate())
			decoder.streamStatus = "vorbis_init_comments"
		case "vorbis_init_comments":
			Log.Debugf("comments: %v", decoder.vc.UserComments())
			Log.Debugf("vendor: %v", decoder.vc.Vendor())
			decoder.streamStatus = "vorbis_init_codebooks"
		case "vorbis_init_codebooks":
			if vorbis.SynthesisInit(&decoder.vd, &decoder.vi) == 0 {
				decoder.vb.Init(&decoder.vd)
			}
			decoder.streamStatus = "vorbis_decode"
		}
	case "vorbis_decode":
		// TODO can raise : panic: runtime error: index out of range
		if vorbis.Synthesis(&decoder.vb, packet) == 0 {
			vorbis.SynthesisBlockin(&decoder.vd, &decoder.vb)
		}

		for samples := 1; samples > 0; {
			var rawFloatBuffer **float32
			samples = vorbis.SynthesisPcmout(&decoder.vd, &rawFloatBuffer)

			if samples > 0 {
				metrics.GetOrRegisterCounter("vorbis.SampleCount", nil).Inc(int64(samples))

				if packet.GranulePos > -1 {
					// Log.Debugf("sampleCount: %d", decoder.sampleCount)
					// Log.Debugf("granule pos: %d", packet.GranulePos)
					// Log.Debugf("%v vorbis sampleCount : %d", time.Now(), decoder.sampleCount)
				}

				// Log.Debugf("read %d samples", samples)
				if decoder.audioHandler != nil {
					decoder.audioHandler.AudioOut(decoder.newAudio(&rawFloatBuffer, samples))
				}
				vorbis.SynthesisRead(&decoder.vd, samples)
			}
		}
	}

	return true
}

func (decoder *VorbisDecoder) newAudio(pcmArray ***float32, sampleCount int) *Audio {
	audio := NewAudio(sampleCount, int(decoder.vi.Channels()))
	// OPTIMISE see vorbis.AnalysisBuffer
	for channel := 0; channel < audio.ChannelCount(); channel++ {
		audio.SetSamples(channel, make([]float32, sampleCount))
		for samplePosition := 0; samplePosition < sampleCount; samplePosition++ {
			audio.samples[channel][samplePosition] = vorbis.PcmArrayHelper(*pcmArray, channel, samplePosition)
		}
	}
	return audio
}

type VorbisEncoder struct {
	Quality      float32
	ChannelCount int32
	SampleRate   int32

	PacketHandler OggPacketHandler

	vi vorbis.Info     // struct that stores all the static vorbis bitstream settings
	vc vorbis.Comment  // struct that stores all the user comments
	vd vorbis.DspState // central working state for the packet PCM decoder
	vb vorbis.Block    // local working space for packet PCM decode
}

func (encoder *VorbisEncoder) Init() error {
	encoder.vi.Init()

	if encoder.ChannelCount == 0 {
		encoder.ChannelCount = 2
	}
	if encoder.SampleRate == 0 {
		encoder.SampleRate = 44100
	}
	if vorbisenc.InitVbr(&encoder.vi, encoder.ChannelCount, encoder.SampleRate, encoder.Quality) != 0 {
		return errors.New("Can't initialize vorbis encoder")
	}

	if vorbis.AnalysisInit(&encoder.vd, &encoder.vi) != 0 {
		return errors.New("Can't initialize vorbis analysis")
	}

	if encoder.vb.Init(&encoder.vd) != 0 {
		return errors.New("Can't initialize vorbis block")
	}

	var (
		header     ogg.Packet
		headerComm ogg.Packet
		headerCode ogg.Packet
	)

	encoder.vc.Init()
	encoder.vc.AddTag("ENCODER", "Go Broadcast v0")

	vorbis.AnalysisHeaderOut(&encoder.vd, &encoder.vc, &header, &headerComm, &headerCode)
	encoder.sendPacket(&header)
	encoder.sendPacket(&headerComm)
	encoder.sendPacket(&headerCode)

	encoder.vc.Clear()

	return nil
}

func (encoder *VorbisEncoder) sendPacket(packet *ogg.Packet) {
	if encoder.PacketHandler != nil {
		encoder.PacketHandler.PacketAvailable(packet)
	}
}

func (encoder *VorbisEncoder) AudioOut(audio *Audio) {
	buffer := vorbis.AnalysisBuffer(&encoder.vd, audio.SampleCount())

	for samplePosition := 0; samplePosition < audio.SampleCount(); samplePosition++ {
		for channel := 0; channel < audio.ChannelCount(); channel++ {
			buffer[channel][samplePosition] = audio.Sample(channel, samplePosition)
		}
	}

	vorbis.AnalysisWrote(&encoder.vd, audio.SampleCount())

	for vorbis.AnalysisBlockOut(&encoder.vd, &encoder.vb) == 1 {
		vorbis.Analysis(&encoder.vb, nil)
		vorbis.BitrateAddBlock(&encoder.vb)

		var packet ogg.Packet

		for vorbis.BitrateFlushPacket(&encoder.vd, &packet) != 0 {
			encoder.sendPacket(&packet)
		}
	}
}
