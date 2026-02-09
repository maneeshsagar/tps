package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)
}

type ZeroLogger struct {
	level  zerolog.Level
	logger zerolog.Logger
}

func NewZeroLogger(level string) *ZeroLogger {
	lvl := parseLevel(level)
	zerolog.SetGlobalLevel(lvl)

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()

	log.Logger = logger

	return &ZeroLogger{
		level:  lvl,
		logger: logger,
	}
}

// Default returns a logger with info level for early initialization
func Default() *ZeroLogger {
	return NewZeroLogger("info")
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func (z *ZeroLogger) Debug(msg string, fields ...any) {
	event := z.logger.Debug()
	if len(fields) > 0 {
		event = event.Fields(fields)
	}
	event.Msg(msg)
}

func (z *ZeroLogger) Info(msg string, fields ...any) {
	event := z.logger.Info()
	if len(fields) > 0 {
		event = event.Fields(fields)
	}
	event.Msg(msg)
}

func (z *ZeroLogger) Warn(msg string, fields ...any) {
	event := z.logger.Warn()
	if len(fields) > 0 {
		event = event.Fields(fields)
	}
	event.Msg(msg)
}

func (z *ZeroLogger) Error(msg string, fields ...any) {
	event := z.logger.Error()
	if len(fields) > 0 {
		event = event.Fields(fields)
	}
	event.Msg(msg)
}

func (z *ZeroLogger) Fatal(msg string, fields ...any) {
	event := z.logger.Fatal()
	if len(fields) > 0 {
		event = event.Fields(fields)
	}
	event.Msg(msg) // exits with os.Exit(1)
}
