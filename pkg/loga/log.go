package log

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogrusLogger(level string) *logrus.Logger {
	logger := logrus.New()
	var logLevel logrus.Level
	switch level {
	case "debug":
		logLevel = logrus.DebugLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	}
	logger.SetLevel(logLevel)

	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.DateTime,
	})
	//hook := &logFileHook{logFile}
	//logger.AddHook(hook)

	return logger
}

type logFileHook struct {
	logFile string
}

func (l *logFileHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
	}
}

func (l *logFileHook) Fire(entry *logrus.Entry) error {
	w := &lumberjack.Logger{
		Filename: l.logFile,
	}
	f := &logrus.JSONFormatter{
		TimestampFormat: time.DateTime,
	}
	content, err := f.Format(entry)
	if err != nil {
		return err
	}
	_, err = io.Writer(w).Write(content)
	if err != nil {
		return err
	}

	return nil
}
