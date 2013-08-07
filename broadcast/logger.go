package broadcast

import (
	"fmt"
	"log/syslog"
	"os"
)

type Logger struct {
	writer *syslog.Writer
}

var Log *Logger = &Logger{}

func (logger *Logger) Writer() *syslog.Writer {
	if logger.writer == nil {
		writer, err := syslog.New(syslog.LOG_DAEMON, "go-broadcast")
		if err != nil {
			fmt.Println("Can't write to syslog")
			os.Exit(1)
		}
		logger.writer = writer
	}
	return logger.writer
}

func (logger *Logger) Debugf(format string, values ...interface{}) {
	logger.Writer().Debug(fmt.Sprintf(format, values...))
}

func (logger *Logger) Printf(format string, values ...interface{}) {
	logger.Writer().Info(fmt.Sprintf(format, values...))
}
