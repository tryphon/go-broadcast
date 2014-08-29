package broadcast

type FileInput struct {
	File string

	sndFile *SndFile
}

func (input *FileInput) Init() error {
	file, err := SndFileOpen(input.File, O_RDONLY, nil)
	if err != nil {
		input.sndFile = nil
		Log.Printf("Can't open new file : %s", input.File)
		return err
	}

	input.sndFile = file
	return nil
}

func (input *FileInput) SampleRate() int {
	return input.sndFile.Info().SampleRate()
}

func (input *FileInput) ChannelCount() int {
	return input.sndFile.Info().Channels()
}

func (input *FileInput) Read() *Audio {
	if input.sndFile == nil {
		return nil
	}

	samples := make([]float32, 1024*input.ChannelCount())
	readCount := input.sndFile.ReadFloat(samples)
	samplesCount := int(readCount) / input.ChannelCount()

	audio := NewAudio(samplesCount, input.ChannelCount())
	audio.LoadInterleavedFloats(samples, samplesCount, input.ChannelCount())

	if int(readCount) != len(samples) {
		input.sndFile = nil
	}

	return audio
}

func (input *FileInput) Close() {
	if input.sndFile == nil {
		return
	}

	input.sndFile.Close()
	input.sndFile = nil
}
