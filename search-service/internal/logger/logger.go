package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(level zerolog.Level, pretty bool) {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(level)
	if pretty {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	} else {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

func GetLogger() zerolog.Logger {
	return log.Logger
}
