package broadcast

import (
	"os"
	"path"
	"time"
)

type TimedFileOutput struct {
	RootDirectory string

	fileDuration time.Duration
	sampleRate   int

	recording            bool
	currentFile          *SndFile
	nextTimeBound        time.Time
	writeQuarantineUntil time.Time

	fileSampleCount         uint32
	expectedFileSampleCount uint32
}

func (output *TimedFileOutput) SampleRate() int {
	if output.sampleRate == 0 {
		output.sampleRate = 44100
	}
	return output.sampleRate
}

func (output *TimedFileOutput) SetSampleRate(sampleRate int) {
	output.sampleRate = sampleRate
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

func (output *TimedFileOutput) fileName(now time.Time, firstFile bool) string {
	// Reference : Mon Jan 2 15:04:05 MST 2006
	format := "2006/01-Jan/02-Mon/15h04.wav"
	if firstFile {
		format = "2006/01-Jan/02-Mon/15h04m05.wav"
	}
	return path.Join(output.RootDirectory, now.Format(format))
}

func (output *TimedFileOutput) checkFileSampleCount() {
	if output.fileSampleCount != output.expectedFileSampleCount {
		Log.Printf("Missing samples in file %s : %d", output.currentFile.Path(), int32(output.expectedFileSampleCount)-int32(output.fileSampleCount))
	}
}

func (output *TimedFileOutput) closeFile() (err error) {
	Log.Printf("Close current file (%s)", output.currentFile.Path())
	output.currentFile.Close()
	output.currentFile = nil
	return nil
}

func (output *TimedFileOutput) newFile(now time.Time) error {
	fileName := output.fileName(now, !output.recording)

	var fileInfo Info
	fileInfo.SetSampleRate(output.SampleRate())
	fileInfo.SetChannels(2)
	fileInfo.SetFormat(FORMAT_WAV | FORMAT_PCM_16)

	os.MkdirAll(path.Dir(fileName), 0775)

	file, err := SndFileOpen(fileName, O_WRONLY, &fileInfo)
	if err != nil {
		output.currentFile = nil
		Log.Printf("Can't open new file : %s", fileName)
		return err
	}

	output.fileSampleCount = 0

	truncatedNow := now.Truncate(output.FileDuration())
	output.nextTimeBound = truncatedNow.Add(output.FileDuration())
	output.expectedFileSampleCount = uint32(output.nextTimeBound.Sub(truncatedNow).Seconds() * float64(output.SampleRate()))

	Log.Printf("Opened new file (%s) until %v", fileName, output.nextTimeBound)

	output.currentFile = file
	return nil
}

func (output *TimedFileOutput) write(audio *Audio) (err error) {
	if output.currentFile != nil {
		output.fileSampleCount += uint32(audio.SampleCount())
		output.currentFile.WriteFloat(audio.InterleavedFloats())
	}
	return nil
}

func (output *TimedFileOutput) Write(audio *Audio) (err error) {
	now := audio.Timestamp()

	defer func() {
		output.recording = (err == nil)
	}()

	if output.recording {
		if now.After(output.nextTimeBound) {
			err = output.closeFile()
			if err != nil {
				return err
			}
			err = output.newFile(now)
			if err != nil {
				return err
			}
		}
	} else {
		err = output.newFile(now)
		if err != nil {
			return err
		}
	}

	output.write(audio)
	return nil
}

func (output *TimedFileOutput) AudioOut(audio *Audio) {
	now := audio.Timestamp()

	if now.After(output.writeQuarantineUntil) {
		err := output.Write(audio)
		if err != nil {
			output.writeQuarantineUntil = now.Add(30 * time.Second)
			Log.Printf("Error to write audio : %s. No write until %v", err.Error(), output.writeQuarantineUntil)
		} else {
			output.writeQuarantineUntil = time.Time{}
		}
	}
}
