package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// WriterHook is a hook that writes logs of specified LogLevels to specified Writer
// Based on: https://github.com/sirupsen/logrus/issues/678
type WriterHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *WriterHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *WriterHook) Levels() []logrus.Level {
	return hook.LogLevels
}

// SetupLogs initialize logger.
func SetupLogs(levelStr string) {
	if levelStr == "" {
		levelStr = "info"
	}
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		logrus.Fatalf("error parsing log level: %+v", err)
	}
	logrus.SetLevel(level)

	if level != logrus.DebugLevel {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	logrus.SetReportCaller(true)

	logrus.SetOutput(io.Discard) // Send all logs to nowhere by default

	logrus.AddHook(&WriterHook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	})
	logrus.AddHook(&WriterHook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
	})
}
