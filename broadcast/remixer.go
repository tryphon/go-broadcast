package broadcast

import (
	"strconv"
	"strings"
)

type Remixer struct {
	OutputChannels []RemixerChannel
	Output         AudioHandler
}

type RemixerChannel struct {
	InputChannels []int
}

func (channel *RemixerChannel) Mix(audio *Audio) []float32 {
	samples := make([]float32, audio.SampleCount())

	if len(channel.InputChannels) > 0 {
		mixLevel := 1.0 / float32(len(channel.InputChannels))

		for _, inputChannel := range channel.InputChannels {
			inputSamples := audio.Samples(inputChannel)
			for samplePosition := 0; samplePosition < audio.SampleCount(); samplePosition++ {
				samples[samplePosition] += inputSamples[samplePosition] * mixLevel
			}
		}
	}

	return samples
}

func (channel *RemixerChannel) String() string {
	outputSpec := ""

	if len(channel.InputChannels) > 0 {
		for index, inputChannel := range channel.InputChannels {
			if index > 0 {
				outputSpec += ","
			}
			outputSpec += strconv.Itoa(inputChannel + 1)
		}
	} else {
		outputSpec += "0"
	}

	return outputSpec
}

func (remixer *Remixer) OutputChannelCount() int {
	return len(remixer.OutputChannels)
}

func (remixer *Remixer) AudioOut(audio *Audio) {
	if remixer.Output == nil {
		return
	}

	remixedAudio := NewAudio(audio.SampleCount(), remixer.OutputChannelCount())

	for index, outputChannel := range remixer.OutputChannels {
		remixedAudio.SetSamples(index, outputChannel.Mix(audio))
	}

	remixer.Output.AudioOut(remixedAudio)
}

func (remixer *Remixer) String() string {
	outputSpecs := ""

	for index, channel := range remixer.OutputChannels {
		if index > 0 {
			outputSpecs += ":"
		}
		outputSpecs += channel.String()
	}

	return outputSpecs
}

func (channel *RemixerChannel) parseSpec(outputSpec string) {
	if outputSpec == "0" {
		channel.InputChannels = make([]int, 0)
		return
	}

	channelSpecs := strings.Split(outputSpec, ",")
	channel.InputChannels = make([]int, len(channelSpecs))

	for index, channelSpec := range channelSpecs {
		userChannel, _ := strconv.Atoi(channelSpec)
		channel.InputChannels[index] = userChannel - 1
	}
}

func NewRemixer(definition string) *Remixer {
	remixer := Remixer{}

	outputSpecs := strings.Split(definition, ":")
	remixer.OutputChannels = make([]RemixerChannel, len(outputSpecs))

	for index, outputSpec := range outputSpecs {
		remixer.OutputChannels[index].parseSpec(outputSpec)
	}

	return &remixer
}
