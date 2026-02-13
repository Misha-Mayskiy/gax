package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

// Init инициализирует глобальный логгер
func Init(level zerolog.Level, pretty bool) {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(level)

	if pretty {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

// GetLogger возвращает инициализированный логгер
func GetLogger() zerolog.Logger {
	return logger
}

// Convenience functions для упрощенного использования
func Debug() *zerolog.Event {
	return logger.Debug()
}

func Info() *zerolog.Event {
	return logger.Info()
}

func Warn() *zerolog.Event {
	return logger.Warn()
}

func Error() *zerolog.Event {
	return logger.Error()
}

func Fatal() *zerolog.Event {
	return logger.Fatal()
}

func Panic() *zerolog.Event {
	return logger.Panic()
}
