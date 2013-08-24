package broadcast

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestTimedFileOutput_SampleRate(t *testing.T) {
	output := TimedFileOutput{}
	output.sampleRate = 96000

	if output.SampleRate() != output.sampleRate {
		t.Errorf("Wrong SampleRate() value :\n got: %v\nwant: %v", output.SampleRate(), output.sampleRate)
	}
}

func TestTimedFileOutput_SampleRate_default(t *testing.T) {
	output := TimedFileOutput{}
	if output.SampleRate() != 44100 {
		t.Errorf("Wrong default SampleRate() :\n got: %v\nwant: %v", output.SampleRate(), 44100)
	}
}

func TestTimedFileOutput_SetSampleRate(t *testing.T) {
	output := TimedFileOutput{}
	output.SetSampleRate(96000)

	if output.sampleRate != 96000 {
		t.Errorf("SetSampleRate() didn't change sampleRate :\n got: %v\nwant: %v", output.sampleRate, 96000)
	}
}

func TestTimedFileOutput_FileDuration(t *testing.T) {
	output := TimedFileOutput{}
	output.fileDuration = 1 * time.Minute

	if output.FileDuration() != output.fileDuration {
		t.Errorf("Wrong FileDuration() value :\n got: %v\nwant: %v", output.FileDuration(), output.fileDuration)
	}
}

func TestTimedFileOutput_FileDuration_default(t *testing.T) {
	output := TimedFileOutput{}
	defaultFileDuration := 5 * time.Minute

	if output.FileDuration() != defaultFileDuration {
		t.Errorf("Wrong default FileDuration() :\n got: %v\nwant: %v", output.FileDuration(), defaultFileDuration)
	}
}

func TestTimedFileOutput_SetFileDuration(t *testing.T) {
	output := TimedFileOutput{}
	output.SetFileDuration(time.Minute)

	if output.fileDuration != time.Minute {
		t.Errorf(" :\n got: %v\nwant: %v", output.fileDuration, time.Minute)
	}
}

func timeReference() time.Time {
	t, _ := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2006")
	return t
}

func TestTimedFileOutput_fileName(t *testing.T) {
	output := TimedFileOutput{}

	fileName := output.fileName(timeReference(), false)
	expectedFileName := "2006/01-Jan/02-Mon/15h04.wav"

	if fileName != expectedFileName {
		t.Errorf("Wrong file name for %v:\n got: %v\nwant: %v", timeReference(), fileName, expectedFileName)
	}

}

func TestTimedFileOutput_fileName_start(t *testing.T) {
	output := TimedFileOutput{}

	fileName := output.fileName(timeReference(), true)
	expectedFileName := "2006/01-Jan/02-Mon/15h04m05.wav"

	if fileName != expectedFileName {
		t.Errorf("Wrong start file name for %v:\n got: %v\nwant: %v", timeReference(), fileName, expectedFileName)
	}
}

func TestTimedFileOutput_fileName_rootDirectory(t *testing.T) {
	output := TimedFileOutput{RootDirectory: "/srv/pige/records"}

	fileName := output.fileName(timeReference(), false)
	expectedFileName := "/srv/pige/records/2006/01-Jan/02-Mon/15h04.wav"

	if fileName != expectedFileName {
		t.Errorf("Wrong file name:\n got: %v\nwant: %v", fileName, expectedFileName)
	}
}

func tempSndFile() (file *SndFile, err error) {
	tempFile, err := ioutil.TempFile("", "timedfileoutput")
	if err != nil {
		return nil, err
	}
	fileName := tempFile.Name()

	var fileInfo Info
	fileInfo.SetSampleRate(44100)
	fileInfo.SetChannels(2)
	fileInfo.SetFormat(FORMAT_WAV | FORMAT_PCM_16)

	file, err = SndFileOpen(fileName, O_WRONLY, &fileInfo)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func TestTimedFileOutput_closeFile(t *testing.T) {
	file, err := tempSndFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Path())

	output := TimedFileOutput{}
	output.currentFile = file

	err = output.closeFile()

	if err != nil {
		t.Errorf("Should not return an error")
	}
	if !file.IsClosed() {
		t.Errorf("The SndFile should be closed")
	}
	if output.currentFile != nil {
		t.Errorf("The currentFile should be nil")
	}
}

func TestTimedFileOutput_newFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "timedfileoutput")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	output := TimedFileOutput{RootDirectory: tempDir}
	output.recording = true

	now, _ := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:05:00 -0700 MST 2006")

	err = output.newFile(now)
	defer output.closeFile()

	if err != nil {
		t.Errorf("Should not return an error")
	}

	if output.currentFile == nil {
		t.Errorf("The currentFile should be defined")
	}
	if output.currentFile.Path() != output.fileName(now, false) {
		t.Errorf("Wrong path :\n got: %v\nwant: %v", output.currentFile.Path(), output.fileName(now, true))
	}
	// pending
	// if output.currentFile.SampleRate() != outout.SampleRate() {
	//   t.Errorf("Wrong sampleRate :\n got: %v\nwant: %v", output.currentFile.SampleRate(), outout.SampleRate())
	// }

	if output.fileSampleCount != 0 {
		t.Errorf("fileSampleCount should zero :\n got: %v\nwant: %v", output.fileSampleCount, 0)
	}
	if output.expectedFileSampleCount != 44100*5*60 {
		t.Errorf("expectedFileSampleCount should be samples for 5 minutes :\n got: %v\nwant: %v", output.fileSampleCount, 44100*5*60)
	}
	if output.nextTimeBound != now.Add(5*time.Minute) {
		t.Errorf("nextTimeBound should be 5 minutes after %v :\n got: %v\nwant: %v", now, output.nextTimeBound, now.Add(5*time.Minute))
	}
}

func TestTimedFileOutput_newFile_error(t *testing.T) {
	output := TimedFileOutput{RootDirectory: "/"}
	err := output.newFile(timeReference())

	if err == nil {
		t.Errorf("Should return an error")
	}
	if output.currentFile != nil {
		t.Errorf("The currentFile should be nil")
	}
}

func TestTimedFileOutput_write(t *testing.T) {
	file, err := tempSndFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Path())

	output := TimedFileOutput{}
	output.currentFile = file

	audio := NewAudio(1024, 2)
	err = output.write(audio)

	if err != nil {
		t.Errorf("Should not return an error")
	}
	if output.fileSampleCount != uint32(audio.SampleCount()) {
		t.Errorf("Wrong fileSampleCount :\n got: %v\nwant: %v", output.fileSampleCount, audio.SampleCount())
	}

	output.closeFile()

	file, err = SndFileOpen(file.Path(), O_RDONLY, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if file.Info().Frames() != int64(audio.SampleCount()) {
		t.Errorf("Wrong sampleCount in file :\n got: %v\nwant: %v", file.Info().Frames(), audio.SampleCount())
	}
}

func TestTimedFileOutput_write_without_file(t *testing.T) {
	output := TimedFileOutput{}

	audio := NewAudio(1024, 2)
	err := output.write(audio)

	// Not a good idea ?
	if err != nil {
		t.Errorf("Should not return an error")
	}
	if output.fileSampleCount != 0 {
		t.Errorf("FileSampleCount should be zero :\n got: %v\nwant: %v", output.fileSampleCount, audio.SampleCount())
	}
}

func TestTimedFileOutput_Write_new_file(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "timedfileoutput")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	output := TimedFileOutput{RootDirectory: tempDir}
	defer output.closeFile()

	audio := NewAudio(1024, 2)
	err = output.Write(audio)

	if err != nil {
		t.Errorf("Should not return an error")
	}

	if output.fileSampleCount != uint32(audio.SampleCount()) {
		t.Errorf("Wrong fileSampleCount :\n got: %v\nwant: %v", output.fileSampleCount, audio.SampleCount())
	}
	if output.currentFile.Path() != output.fileName(audio.Timestamp(), true) {
		t.Errorf("Wrong path :\n got: %v\nwant: %v", output.currentFile.Path(), output.fileName(audio.Timestamp(), true))
	}
	if !output.recording {
		t.Errorf("Should be in recording mode")
	}
}

func TestTimedFileOutput_Write_next_file(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "timedfileoutput")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	output := TimedFileOutput{RootDirectory: tempDir}
	defer output.closeFile()

	audio := NewAudio(1024, 2)
	err = output.Write(audio)
	if err != nil {
		t.Fatal(err)
	}

	audio.SetTimestamp(audio.Timestamp().Add(output.FileDuration()))
	err = output.Write(audio)
	if err != nil {
		t.Errorf("Should not return an error")
	}

	if !output.recording {
		t.Errorf("Should be in recording mode")
	}
	if output.currentFile.Path() != output.fileName(audio.Timestamp(), false) {
		t.Errorf("Wrong path :\n got: %v\nwant: %v", output.currentFile.Path(), output.fileName(audio.Timestamp(), true))
	}
}
