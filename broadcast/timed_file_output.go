package broadcast

import (
	"os"
	"path"
	"time"
)

type TimedFileOutput struct {
	RootDirectory string

	currentFile   *SndFile
	nextTimeBound time.Time
}

func (output *TimedFileOutput) FileDuration() time.Duration {
	return 5 * time.Minute
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
}

func (output *TimedFileOutput) AudioOut(audio *Audio) {
	now := time.Now()
	fileName := output.fileName(now, false)

	if output.nextTimeBound.IsZero() {
		output.updateNextTimeBound(now)
	}

	if output.currentFile != nil {
		if now.After(output.nextTimeBound) {
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
		fileInfo.SetSampleRate(44100)
		fileInfo.SetChannels(2)
		fileInfo.SetFormat(FORMAT_WAV | FORMAT_PCM_16)

		Log.Printf("Open new file (%s) until %v", fileName, output.nextTimeBound)

		os.MkdirAll(path.Dir(fileName), 0775)

		file, err := SndFileOpen(fileName, O_WRONLY, &fileInfo)
		if err != nil {
			Log.Printf("Can't open new file : %s", fileName)
			return
		}
		output.currentFile = file
	}

	output.currentFile.WriteFloat(audio.InterleavedFloats())
}
