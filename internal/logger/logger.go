package logger

import (
	"fmt"
	"log"
	"os"
)

const (
	logDir  = "/var/log/zep"
	logPerm = 0667
)

// debug , info , warn , error
type Logger struct {
	debug   *log.Logger
	warn    *log.Logger
	info    *log.Logger
	error   *log.Logger
	LogPath string
}

func setupDir() error {
	err := os.MkdirAll(logDir, 0755)

	if err != nil {
		return fmt.Errorf("failed to created log dir:%v", err)
	}

	return nil
}

func NewLogger(logFileName string) (*Logger, error) {
	err := setupDir()

	if err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(fmt.Sprintf("%s/%s", logDir, logFileName), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file:%v", err)
	}

	return &Logger{
		debug:   log.New(logFile, "DEBUG:", log.LstdFlags),
		warn:    log.New(logFile, "WARN:", log.LstdFlags),
		info:    log.New(logFile, "INFO:", log.LstdFlags),
		error:   log.New(logFile, "ERROR:", log.LstdFlags),
		LogPath: fmt.Sprintf("%s/%s", logDir, logFileName),
	}, nil
}

func (log *Logger) Debug(v ...interface{}) {
	log.debug.Println(v...)
}

func (log *Logger) Warn(v ...interface{}) {
	log.warn.Println(v...)
}

func (log *Logger) Info(v ...interface{}) {
	log.info.Println(v...)
}

func (log *Logger) Error(v ...interface{}) {
	log.error.Println(v...)
}
