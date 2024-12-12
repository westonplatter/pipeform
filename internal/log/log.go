package log

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

type Level string

const (
	LevelTrace Level = "trace"
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

func PossibleLevels() []Level {
	return []Level{LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError}
}

type Logger struct {
	hclog.Logger
	f *os.File
}

func NewLogger(level Level, path string) (*Logger, error) {
	if level == "" || path == "" {
		return &Logger{Logger: hclog.NewNullLogger()}, nil
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}

	hclogger := hclog.New(&hclog.LoggerOptions{
		Name:   "pipeform",
		Level:  hclog.LevelFromString(string(level)),
		Output: f,
	})

	return &Logger{
		Logger: hclogger,
		f:      f,
	}, nil
}

func (l *Logger) Close() error {
	if l.f == nil {
		return nil
	}
	if err := l.f.Close(); err != nil {
		return err
	}
	l.f = nil
	return nil
}
