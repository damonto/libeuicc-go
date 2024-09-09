package libeuicc

import (
	"log/slog"
	"os"
)

type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, err error, args ...any)
}

const (
	LogDebugLevel slog.Level = slog.LevelDebug
	LogInfoLevel  slog.Level = slog.LevelInfo
	LogWarnLevel  slog.Level = slog.LevelWarn
	LogErrorLevel slog.Level = slog.LevelError
)

type DefaultLogger struct {
	logger *slog.Logger
}

var logger Logger = NewDefaultLogger(LogErrorLevel)

func NewDefaultLogger(level slog.Level) Logger {
	slogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(level),
	}))
	return &DefaultLogger{
		logger: slogger,
	}
}

func (l *DefaultLogger) Debug(format string, args ...any) {
	l.logger.Debug(format, args...)
}

func (l *DefaultLogger) Info(format string, args ...any) {
	l.logger.Info(format, args...)
}

func (l *DefaultLogger) Warn(format string, args ...any) {
	l.logger.Warn(format, args...)
}

func (l *DefaultLogger) Error(format string, err error, args ...any) {
	args = append([]any{"error", err}, args...)
	l.logger.Error(format, args...)
}
