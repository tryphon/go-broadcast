package broadcast

import (
	alsa "github.com/tryphon/alsa-go"
	// "time"
)

type AlsaOutput struct {
	handle     alsa.Handle
	Device     string
	SampleRate int

	sampleCount int64
}

func (output *AlsaOutput) Init() error {
	if output.Device == "" {
		output.Device = "default"
	}

	err := output.handle.Open(output.Device, alsa.StreamTypePlayback, alsa.ModeBlock)
	if err != nil {
		return err
	}

	if output.SampleRate == 0 {
		output.SampleRate = 44100
	}

	output.handle.SampleFormat = alsa.SampleFormatS16LE
	output.handle.SampleRate = output.SampleRate
	output.handle.Channels = 2

	err = output.handle.ApplyHwParams()
	return err
}

func (alsa *AlsaOutput) AudioOut(audio *Audio) {
	alsa.handle.AvailUpdate()

	// for alsa.Delay() < 0 {
	// 	Log.Debugf("Alsa delay is negative (%d), waiting for better conditions", alsa.Delay())
	// 	time.Sleep(time.Second)
	// }

	pcmBytes := audio.PcmBytes()

	alsaWriteLength, err := alsa.handle.Write(pcmBytes)
	if err != nil {
		Log.Debugf("Can't write alsa buffer: %v", err.Error())
		return
	}

	wroteSamples := int64(alsaWriteLength / len(pcmBytes) * audio.sampleCount)
	alsa.sampleCount += wroteSamples

	if alsaWriteLength != len(pcmBytes) {
		Log.Debugf("Did not write whole alsa buffer (Wrote %v, expected %v)", alsaWriteLength, len(pcmBytes))
	}
}

func (alsa *AlsaOutput) SampleCount() int64 {
	return alsa.sampleCount
}

func (alsa *AlsaOutput) Delay() (delay int) {
	delay, _ = alsa.handle.Delay()
	return delay
}