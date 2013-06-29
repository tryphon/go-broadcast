package broadcast

import (
	"fmt"
	alsa "github.com/Narsil/alsa-go"
	"os"
)

type AlsaSink struct {
	handle      alsa.Handle
	sampleCount int64
}

func (sink *AlsaSink) Init() error {
	err := sink.handle.Open("default", alsa.StreamTypePlayback, alsa.ModeBlock)
	return err

	sink.handle.SampleFormat = alsa.SampleFormatS16LE
	sink.handle.SampleRate = 44100
	sink.handle.Channels = 2

	err = sink.handle.ApplyHwParams()
	return err
}

func (alsa *AlsaSink) AudioOut(audio *Audio) {
	pcmBytes := audio.PcmBytes()
	alsaWriteLength, err := alsa.handle.Write(pcmBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Write alsa buffer ", err.Error())
	}

	wroteSamples := int64(alsaWriteLength / len(pcmBytes) * audio.sampleCount)
	// fmt.Printf("wrote %d samples in alsa\n", wroteSamples)
	alsa.sampleCount += wroteSamples

	if alsaWriteLength != len(pcmBytes) {
		fmt.Fprintf(os.Stderr, "Did not write whole alsa buffer (Wrote %v, expected %v)\n", alsaWriteLength, pcmBytes)
	}

	// fmt.Printf("%v alsa sampleCount : %d\n", time.Now(), alsa.sampleCount)

	// fmt.Printf("wrote %d bytes in alsa\n", alsaWriteLength)
}

func (alsa *AlsaSink) SampleCount() int64 {
	return alsa.sampleCount
}
