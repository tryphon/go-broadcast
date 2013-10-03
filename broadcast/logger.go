package broadcast

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
)

type Logger struct {
	Syslog bool
	Debug  bool

	writer LogWriter
}

type LogWriter interface {
	Debug(message string) error
	Info(message string) error
	Emerg(message string) error
}

type StdWriter struct {
	Out io.Writer
}

func (writer *StdWriter) Debug(message string) error {
	fmt.Fprintln(writer.Out, message)
	return nil
}
func (writer *StdWriter) Info(message string) error {
	fmt.Fprintln(writer.Out, message)
	return nil
}
func (writer *StdWriter) Emerg(message string) error {
	fmt.Fprintln(writer.Out, message)
	return nil
}

var Log *Logger = &Logger{}

func (logger *Logger) Writer() LogWriter {
	if logger.writer == nil {
		if logger.Syslog {
			syslogWriter, err := syslog.New(syslog.LOG_DAEMON, "go-broadcast")
			if err != nil {
				panic("Can't write to syslog")
			}
			logger.writer = syslogWriter
		} else {
			logger.writer = &StdWriter{Out: os.Stderr}
		}
	}
	return logger.writer
}

func (logger *Logger) Debugf(format string, values ...interface{}) {
	if logger.Debug {
		logger.Writer().Debug(fmt.Sprintf(format, values...))
	}
}

func (logger *Logger) Printf(format string, values ...interface{}) {
	logger.Writer().Info(fmt.Sprintf(format, values...))
}

func (logger *Logger) Panicf(format string, values ...interface{}) {
	message := fmt.Sprintf(format, values...)
	logger.Writer().Emerg(message)
	panic(message)
}
