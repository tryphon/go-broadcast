package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"io"
	alsa "github.com/Narsil/alsa-go"
	"github.com/grd/ogg"
	"github.com/grd/vorbis"
)

type OggDecoder struct {
	handler OggHandler

	oy  ogg.SyncState
	oss ogg.StreamState // take physical pages, weld into a logical stream of packets
	og  ogg.Page        // one Ogg bitstream page. Vorbis packets are inside
	op  ogg.Packet      // one raw packet of data for decode
}

type OggHandler interface {
	NewStream(serialNo int32)
	PacketOut(packet *ogg.Packet)
}

func (decoder *OggDecoder) Read(reader io.Reader) {
	buffer := decoder.oy.Buffer(4096)

	readLength, err := reader.Read(buffer)
	checkError(err)

	decoder.oy.Wrote(readLength)

	for decoder.oy.PageOut(&decoder.og) == 1 {
		fmt.Printf("Page number: %d, granule pos: %d\n", decoder.og.PageNo(), decoder.og.GranulePos())

		if decoder.oss.SerialNo != decoder.og.SerialNo() {
			fmt.Printf("Init Ogg Stream State %d\n", decoder.og.SerialNo())
			decoder.oss.Init(decoder.og.SerialNo())

			decoder.handler.NewStream(decoder.og.SerialNo())
		}

		err = decoder.oss.PageIn(&decoder.og)
		checkError(err)

		for decoder.oss.PacketOut(&decoder.op) == 1 {
			// fmt.Printf("PacketOut\n");

			// if result < 1 {
			// 	fmt.Printf("Error reading next packet.\n");
			// 	os.Exit(1)
			// }

			decoder.handler.PacketOut(&decoder.op)
		}
	}
}

func New(handler OggHandler) *OggDecoder {
	decoder := new(OggDecoder)
	decoder.handler = handler
	return decoder
}

type AudioHandler interface {
	AudioOut(audio *Audio)
}

type VorbisDecoder struct {
	streamStatus string

	vi vorbis.Info     // struct that stores all the static vorbis bitstream settings
	vc vorbis.Comment  // struct that stores all the user comments
	vd vorbis.DspState // central working state for the packet PCM decoder
	vb vorbis.Block    // local working space for packet PCM decode

	audioHandler AudioHandler
}

func (decoder *VorbisDecoder) NewStream(serialNo int32) {
	decoder.vi.Init()
	decoder.vc.Init()
}

func (decoder *VorbisDecoder) PacketOut(packet *ogg.Packet) {
	if decoder.streamStatus == "" {
		decoder.streamStatus = "vorbis_init_info"
	}

	switch decoder.streamStatus {
	case "vorbis_init_info", "vorbis_init_comments", "vorbis_init_codebooks":
		fmt.Printf("Init vorbis header %s.\n", decoder.streamStatus);

		if vorbis.SynthesisHeaderIn(&decoder.vi, &decoder.vc, packet) < 0 {
			fmt.Printf("This Ogg bitstream does not contain Vorbis audio data.\n");
			os.Exit(1)
		}

		switch decoder.streamStatus {
		case "vorbis_init_info":
			fmt.Printf("Bitstream is %d channel, %dHz\n",decoder.vi.Channels(),decoder.vi.Rate())
			decoder.streamStatus = "vorbis_init_comments"
		case "vorbis_init_comments":
			fmt.Printf("comments: %v\n", decoder.vc.UserComments())
			fmt.Printf("vendor: %v\n", decoder.vc.Vendor())
			decoder.streamStatus = "vorbis_init_codebooks"
		case "vorbis_init_codebooks":
			if vorbis.SynthesisInit(&decoder.vd,&decoder.vi) == 0 {
				decoder.vb.Init(&decoder.vd)
			}
			decoder.streamStatus = "vorbis_decode"
		}
	case "vorbis_decode":
		if vorbis.Synthesis(&decoder.vb, packet) == 0 {
			vorbis.SynthesisBlockin(&decoder.vd, &decoder.vb)
		}

		for samples := 1; samples > 0; {
			var rawFloatBuffer **float32
			samples = vorbis.SynthesisPcmout(&decoder.vd, &rawFloatBuffer)

			if samples > 0 {
				// fmt.Printf("read %d samples\n", samples)
				if decoder.audioHandler != nil {
					audio := new(Audio)
					audio.LoadPcmArray(&rawFloatBuffer, samples, 2)

					decoder.audioHandler.AudioOut(audio)
				}
				vorbis.SynthesisRead(&decoder.vd,samples)
			}
		}
	}

}

type Audio struct {
	samples [][]float32
	channelCount int
	sampleCount int
}

func (audio *Audio) LoadPcmArray(pcmArray ***float32, sampleCount int, channelCount int) {
	audio.samples = make([][]float32, 2)
	audio.channelCount = channelCount
	audio.sampleCount = sampleCount

	// OPTIMISE see vorbis.AnalysisBuffer
	for channel := 0; channel < channelCount; channel++ {
		audio.samples[channel] = make([]float32, sampleCount)
		for samplePosition := 0; samplePosition < sampleCount; samplePosition++ {
			audio.samples[channel][samplePosition] = vorbis.PcmArrayHelper(*pcmArray, channel, samplePosition)
		}
	}
}

func floatSamplesToBytes(sample float32) (byte, byte) {
	integerValue := int16(sample * 32768)
	return byte(integerValue),byte(integerValue >> 8)
}

func (audio *Audio) PcmBytes() ([]byte) {
	pcmSampleSize := 4
	pcmBytesLength := audio.sampleCount * pcmSampleSize
	pcmBytes := make([]byte, pcmBytesLength)

	for samplePosition := 0; samplePosition < audio.sampleCount; samplePosition++ {
		pcmBytes[samplePosition * pcmSampleSize], pcmBytes[samplePosition * pcmSampleSize + 1] = floatSamplesToBytes(audio.samples[0][samplePosition])
		pcmBytes[samplePosition * pcmSampleSize + 2], pcmBytes[samplePosition * pcmSampleSize + 3] = floatSamplesToBytes(audio.samples[1][samplePosition])
	}

	return pcmBytes
}

type AlsaSink struct {
	handle alsa.Handle
}

func (sink *AlsaSink) Init() {
	err := sink.handle.Open("default", alsa.StreamTypePlayback, alsa.ModeBlock)
	checkError(err)

	sink.handle.SampleFormat = alsa.SampleFormatS16LE
	sink.handle.SampleRate = 44100
	sink.handle.Channels = 2

	err = sink.handle.ApplyHwParams()
	checkError(err)
}

func (alsa *AlsaSink) AudioOut(audio *Audio) {
	pcmBytes := audio.PcmBytes()
	alsaWriteLength, err := alsa.handle.Write(pcmBytes)
	checkError(err)

	if alsaWriteLength != len(pcmBytes) {
	 	fmt.Fprintf(os.Stderr, "Did not write whole alsa buffer (Wrote %v, expected %v)\n", alsaWriteLength, pcmBytes)
	}

	// fmt.Printf("wrote %d bytes in alsa\n", alsaWriteLength)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "http://host:port/page")
		os.Exit(1)
	}
	url, err := url.Parse(os.Args[1])
	checkError(err)

	client := &http.Client{}

	request, err := http.NewRequest("GET", url.String(), nil)
	checkError(err)

	response, err := client.Do(request)
	if response.Status != "200 OK" {
		fmt.Println(response.Status)
		os.Exit(2)
	}

	var (
		oggDecoder OggDecoder
		vorbisDecoder VorbisDecoder
		alsaSink AlsaSink
	)

	alsaSink.Init()
	vorbisDecoder.audioHandler = &alsaSink
	oggDecoder.handler = &vorbisDecoder

	reader := response.Body
	for {
		oggDecoder.Read(reader)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
