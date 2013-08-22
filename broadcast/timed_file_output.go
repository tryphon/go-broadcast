package broadcast

import (
	"os"
	"path"
	"time"
)

type TimedFileOutput struct {
	RootDirectory string

	fileDuration  time.Duration
	currentFile   *SndFile
	nextTimeBound time.Time

	fileSampleCount         uint32
	expectedFileSampleCount uint32
}

func (output *TimedFileOutput) SampleRate() int {
	return 44100
}

func (output *TimedFileOutput) FileDuration() time.Duration {
	if output.fileDuration == 0 {
		output.fileDuration = 5 * time.Minute
	}
	return output.fileDuration
}

func (output *TimedFileOutput) SetFileDuration(fileDuration time.Duration) {
	output.fileDuration = fileDuration
}

func (output *TimedFileOutput) fileName(now time.Time, startFile bool) string {
	// Reference : Mon Jan 2 15:04:05 MST 2006
	format := "2006/01-Jan/02-Mon/15h04.wav"
	if startFile {
		format = "2006/01-Jan/02-Mon/15h04m05.wav"
	}
	return path.Join(output.RootDirectory, now.Format(format))
}

func (output *TimedFileOutput) updateNextTimeBound(now time.Time) {
	output.nextTimeBound = now.Truncate(output.FileDuration()).Add(output.FileDuration())
	output.expectedFileSampleCount = uint32(output.nextTimeBound.Sub(now).Seconds() * float64(output.SampleRate()))
}

func (output *TimedFileOutput) AudioOut(audio *Audio) {
	now := time.Now()
	fileName := output.fileName(now, false)

	if output.nextTimeBound.IsZero() {
		output.updateNextTimeBound(now)
	}

	if output.currentFile != nil {
		if now.After(output.nextTimeBound) {
			if output.fileSampleCount != output.expectedFileSampleCount {
				Log.Printf("Missing samples in file %s : %d", output.currentFile.Path(), output.expectedFileSampleCount-output.fileSampleCount)
			}

			output.updateNextTimeBound(now)

			Log.Printf("Close current file (%s)", output.currentFile.Path())

			output.currentFile.Close()
			output.currentFile = nil
		}
	} else {
		fileName = output.fileName(now, true)
	}

	if output.currentFile == nil {
		var fileInfo Info
		fileInfo.SetSampleRate(output.SampleRate())
		fileInfo.SetChannels(2)
		fileInfo.SetFormat(FORMAT_WAV | FORMAT_PCM_16)

		Log.Printf("Open new file (%s) until %v", fileName, output.nextTimeBound)

		os.MkdirAll(path.Dir(fileName), 0775)

		file, err := SndFileOpen(fileName, O_WRONLY, &fileInfo)
		if err != nil {
			Log.Printf("Can't open new file : %s", fileName)
			return
		}
		output.fileSampleCount = 0
		output.currentFile = file
	}

	output.fileSampleCount += uint32(audio.SampleCount())
	output.currentFile.WriteFloat(audio.InterleavedFloats())
}
