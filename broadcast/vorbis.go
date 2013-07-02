package broadcast

import (
	"fmt"
	"github.com/grd/ogg"
	"github.com/grd/vorbis"
)

type VorbisDecoder struct {
	streamStatus string

	vi vorbis.Info     // struct that stores all the static vorbis bitstream settings
	vc vorbis.Comment  // struct that stores all the user comments
	vd vorbis.DspState // central working state for the packet PCM decoder
	vb vorbis.Block    // local working space for packet PCM decode

	audioHandler AudioHandler
	sampleCount  int64
}

func (decoder *VorbisDecoder) SetAudioHandler(audioHandler AudioHandler) {
	decoder.audioHandler = audioHandler
}

func (decoder *VorbisDecoder) SampleCount() int64 {
	return decoder.sampleCount
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
		fmt.Printf("Init vorbis header %s.\n", decoder.streamStatus)

		if vorbis.SynthesisHeaderIn(&decoder.vi, &decoder.vc, packet) < 0 {
			// TODO we should close ogg and reader to retry
			fmt.Printf("This Ogg bitstream does not contain Vorbis audio data.\n")
			return false
		}

		switch decoder.streamStatus {
		case "vorbis_init_info":
			fmt.Printf("Bitstream is %d channel, %dHz\n", decoder.vi.Channels(), decoder.vi.Rate())
			decoder.streamStatus = "vorbis_init_comments"
		case "vorbis_init_comments":
			fmt.Printf("comments: %v\n", decoder.vc.UserComments())
			fmt.Printf("vendor: %v\n", decoder.vc.Vendor())
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
				decoder.sampleCount += int64(samples)

				if packet.GranulePos > -1 {
					// fmt.Printf("sampleCount: %d\n", decoder.sampleCount)
					// fmt.Printf("granule pos: %d\n", packet.GranulePos)
					// fmt.Printf("%v vorbis sampleCount : %d\n", time.Now(), decoder.sampleCount)
				}

				// fmt.Printf("read %d samples\n", samples)
				if decoder.audioHandler != nil {
					audio := new(Audio)
					audio.LoadPcmArray(&rawFloatBuffer, samples, 2)

					decoder.audioHandler.AudioOut(audio)
				}
				vorbis.SynthesisRead(&decoder.vd, samples)
			}
		}
	}

	return true
}
