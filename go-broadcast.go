package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	alsa "github.com/Narsil/alsa-go"
	"github.com/grd/ogg"
	"github.com/grd/vorbis"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "http://host:port/page")
		os.Exit(1)
	}
	url, err := url.Parse(os.Args[1])
	checkError(err)

	client := &http.Client{}

	request, err := http.NewRequest("GET", url.String(), nil)
	// only accept UTF-8
	// request.Header.Add("Accept-Charset", "UTF-8;q=1, ISO-8859-1;q=0")
	checkError(err)

	response, err := client.Do(request)
	if response.Status != "200 OK" {
		fmt.Println(response.Status)
		os.Exit(2)
	}

	// filename := flag.String("file", "default", "The file to play/record. Default is stdin for play, stdout for record.")
	filename := "output.ogg"

	var file *os.File
	file, err = os.Create(filename)

	var (
		oy  ogg.SyncState
		oss ogg.StreamState // take physical pages, weld into a logical stream of packets
		og  ogg.Page        // one Ogg bitstream page. Vorbis packets are inside
		op  ogg.Packet      // one raw packet of data for decode

		vi vorbis.Info     // struct that stores all the static vorbis bitstream settings
		vc vorbis.Comment  // struct that stores all the user comments
		vd vorbis.DspState // central working state for the packet PCM decoder
		vb vorbis.Block    // local working space for packet PCM decode
	)

	bufferSize := 4096
	stream_status := "new"

	var floatBuffer **float32

	handle := alsa.New()
	err = handle.Open("default", alsa.StreamTypePlayback, alsa.ModeBlock)
	checkError(err)

	handle.SampleFormat = alsa.SampleFormatS16LE
	handle.SampleRate = 44100
	handle.Channels = 2

	alsaSampleSize := 4
	alsaBufferLength := 2048 * alsaSampleSize
	var alsaBuffer = make([]byte, alsaBufferLength)

	handle.ApplyHwParams()
	checkError(err)

	reader := response.Body
	fmt.Println("got body")
	for {
		fmt.Printf("Read buffer\n");
		buffer := oy.Buffer(bufferSize)

		readLength, err := reader.Read(buffer)
		checkError(err)

		oy.Wrote(readLength)

		for oy.PageOut(&og) == 1 {
			if oss.SerialNo != og.SerialNo() {
				stream_status = "new"
			}

			fmt.Printf("Page number: %d, granule pos: %d\n", og.PageNo(), og.GranulePos())


			if stream_status == "new" {
				fmt.Printf("Init Ogg Stream State %d\n", og.SerialNo());
				oss.Init(og.SerialNo())

				vi.Init()
				vc.Init()

				stream_status = "vorbis_init_info"
			}

			err = oss.PageIn(&og)
			checkError(err)

			for oss.PacketOut(&op) == 1 {

				// fmt.Printf("PacketOut\n");

				// if result < 1 {
				// 	fmt.Printf("Error reading next packet.\n");
				// 	os.Exit(1)
				// }

				switch stream_status {
				case "vorbis_init_info", "vorbis_init_comments", "vorbis_init_codebooks":
					fmt.Printf("Init vorbis header %s.\n", stream_status);

					if vorbis.SynthesisHeaderIn(&vi, &vc, &op) < 0 {
						fmt.Printf("This Ogg bitstream does not contain Vorbis audio data.\n");
						os.Exit(1)
					}

					switch stream_status {
					case "vorbis_init_info":
						fmt.Printf("Bitstream is %d channel, %dHz\n",vi.Channels(),vi.Rate())
						stream_status = "vorbis_init_comments"
					case "vorbis_init_comments":
						fmt.Printf("comments: %v\n", vc.UserComments())
						fmt.Printf("vendor: %v\n", vc.Vendor())
						stream_status = "vorbis_init_codebooks"
					case "vorbis_init_codebooks":
						if vorbis.SynthesisInit(&vd,&vi) == 0 {
							vb.Init(&vd)
						}

						stream_status = "vorbis_decode"

					}
				case "vorbis_decode":
					if vorbis.Synthesis(&vb, &op) == 0 {
						vorbis.SynthesisBlockin(&vd, &vb)
					}

					for samples := 1; samples > 0; {
						samples = vorbis.SynthesisPcmout(&vd, &floatBuffer)
						fmt.Printf("read %d samples\n", samples)

						if samples > 0 {
							// sample1_left_byte1, sample1_left_byte2, sample1_right_byte1, sample1_right_byte2, sample2_left_byte1, ...
							for samplePosition := 0; samplePosition < samples; samplePosition++ {
								alsaBuffer[samplePosition * alsaSampleSize], alsaBuffer[samplePosition * alsaSampleSize + 1] = floatSamplesToBytes(vorbis.PcmArrayHelper(floatBuffer, 0, samplePosition))
								alsaBuffer[samplePosition * alsaSampleSize + 2], alsaBuffer[samplePosition * alsaSampleSize + 3] = floatSamplesToBytes(vorbis.PcmArrayHelper(floatBuffer, 1, samplePosition))
							}

							alsaSamplesLength := samples * alsaSampleSize

							alsaWriteLength, err := handle.Write(alsaBuffer[:alsaSamplesLength])
							checkError(err)
							if alsaWriteLength != alsaSamplesLength {
								fmt.Fprintf(os.Stderr, "Did not write whole alsa buffer (Wrote %v, expected %v)\n", alsaWriteLength, alsaSamplesLength)
							}

							fmt.Printf("wrote %d bytes in alsa\n", alsaWriteLength)

							vorbis.SynthesisRead(&vd,samples)
						}
					}
				}

				switch stream_status {
				case "vorbis_init_info":
				case "vorbis_init_comments":

				}
			}

			fmt.Printf("New stream status %s.\n", stream_status);
		}

		writeLength, err := file.Write(buffer[:readLength])
		checkError(err)
		if writeLength != readLength {
			fmt.Fprintf(os.Stderr, "Did not write whole buffer (Wrote %v, expected %v)\n", writeLength, readLength)
			os.Exit(1)
		}
	}

	os.Exit(0)
}

func floatSamplesToBytes(sample float32) (byte, byte) {
	integerValue := int16(sample * 32768)
	return byte(integerValue),byte(integerValue >> 8)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
