package config

import (
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	MongoURI        string
	MongoDB         string
	KafkaBroker     string
	KafkaTopic      string
	UserServiceAddr string
	ChatServicePort string
	LogLevel        zerolog.Level
	LogPretty       bool
}

func New() *Config {
	return &Config{
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:         getEnv("MONGO_DB", "chatdb"),
		KafkaBroker:     getEnv("KAFKA_BROKER", "localhost:29092"),
		KafkaTopic:      getEnv("KAFKA_TOPIC_MESSAGES", "messages"),
		UserServiceAddr: getEnv("USER_SERVICE_ADDR", "localhost:8082"),
		ChatServicePort: getEnv("CHAT_SERVICE_PORT", ":8083"),
		LogLevel:        parseLogLevel(getEnv("LOG_LEVEL", "info")),
		LogPretty:       getEnv("LOG_PRETTY", "true") == "true",
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
