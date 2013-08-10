package broadcast

import (
	"fmt"
	alsa "github.com/Narsil/alsa-go"
	"os"
)

type AlsaOutput struct {
	handle      alsa.Handle
	sampleCount int64
}

func (output *AlsaOutput) Init() error {
	err := output.handle.Open("default", alsa.StreamTypePlayback, alsa.ModeBlock)
	if err != nil {
		return err
	}

	output.handle.SampleFormat = alsa.SampleFormatS16LE
	output.handle.SampleRate = 44100
	output.handle.Channels = 2

	err = output.handle.ApplyHwParams()
	return err
}

func (alsa *AlsaOutput) AudioOut(audio *Audio) {
	pcmBytes := audio.PcmBytes()

	alsaWriteLength, err := alsa.handle.Write(pcmBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't write alsa buffer ", err.Error(), "\n")
		return
	}

	wroteSamples := int64(alsaWriteLength / len(pcmBytes) * audio.sampleCount)
	// Log.Debugf("wrote %d samples in alsa\n", wroteSamples)
	alsa.sampleCount += wroteSamples

	if alsaWriteLength != len(pcmBytes) {
		fmt.Fprintf(os.Stderr, "Did not write whole alsa buffer (Wrote %v, expected %v)\n", alsaWriteLength, len(pcmBytes))
	}

	// Log.Debugf("%v alsa sampleCount : %d\n", time.Now(), alsa.sampleCount)

	// Log.Debugf("wrote %d bytes in alsa\n", alsaWriteLength)
}

func (alsa *AlsaOutput) SampleCount() int64 {
	return alsa.sampleCount
}
