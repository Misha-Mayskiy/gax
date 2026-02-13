package config

import (
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	PostgresDSN string
	KafkaBroker string
	ElasticURL  string
	HTTPPort    string
	LogLevel    zerolog.Level
	LogPretty   bool
}

func New() *Config {
	return &Config{
		PostgresDSN: getEnv("POSTGRES_DSN", ""),
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:29092"),
		ElasticURL:  getEnv("ELASTIC_URL", "http://localhost:9200"),
		HTTPPort:    getEnv("HTTP_PORT", ":8085"),
		LogLevel:    parseLogLevel(getEnv("LOG_LEVEL", "info")),
		LogPretty:   getEnv("LOG_PRETTY", "true") == "true",
	}
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}
