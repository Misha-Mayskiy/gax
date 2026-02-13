package config

import (
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	DatabaseURL      string
	RedisAddr        string
	KafkaBroker      string
	KafkaTopicUsers  string
	KafkaTopicSearch string
	UserServicePort  string
	LogLevel         zerolog.Level
	LogPretty        bool
}

func Load() *Config {
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/users?sslmode=disable"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBroker:      getEnv("KAFKA_BROKER", "localhost:29092"),
		KafkaTopicUsers:  getEnv("KAFKA_TOPIC_USERS", "users"),
		KafkaTopicSearch: getEnv("KAFKA_TOPIC_SEARCH_EVENTS", "search-events"),
		UserServicePort:  getEnv("USER_SERVICE_PORT", ":8082"),
		LogLevel:         parseLogLevel(getEnv("LOG_LEVEL", "info")),
		LogPretty:        getEnv("LOG_PRETTY", "true") == "true",
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
