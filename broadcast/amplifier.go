package broadcast

type Amplifier struct {
	Amplification float32
	Output        AudioHandler
}

func (amplifier *Amplifier) AudioOut(audio *Audio) {
	if amplifier.Output == nil {
		return
	}

	if amplifier.Amplification != 0 {
		audio.Process(amplifier.amplify)
	}

	amplifier.Output.AudioOut(audio)
}

func (amplifier *Amplifier) amplify(_ int, _ int, sample float32) float32 {
	return sample * (amplifier.Amplification + 1)
}
