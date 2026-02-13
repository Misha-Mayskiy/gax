package config

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func TestLoad_Defaults(t *testing.T) {
	keys := []string{
		"DATABASE_URL", "REDIS_ADDR", "KAFKA_BROKER",
		"KAFKA_TOPIC_USERS", "KAFKA_TOPIC_SEARCH_EVENTS",
		"USER_SERVICE_PORT", "LOG_LEVEL", "LOG_PRETTY",
	}

	for _, key := range keys {
		t.Setenv(key, "")
	}

	cfg := Load()

	// Проверяем дефолтные значения (hardcoded в config.go)
	if cfg.DatabaseURL != "postgres://postgres:postgres@localhost:5433/users?sslmode=disable" {
		t.Errorf("expected default DatabaseURL, got %s", cfg.DatabaseURL)
	}
	if cfg.RedisAddr != "localhost:6379" {
		t.Errorf("expected default RedisAddr, got %s", cfg.RedisAddr)
	}
	if cfg.KafkaBroker != "localhost:29092" {
		t.Errorf("expected default KafkaBroker, got %s", cfg.KafkaBroker)
	}
	if cfg.LogLevel != zerolog.InfoLevel {
		t.Errorf("expected default LogLevel Info, got %v", cfg.LogLevel)
	}
	if !cfg.LogPretty {
		t.Errorf("expected default LogPretty true, got %v", cfg.LogPretty)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	// Устанавливаем кастомные значения
	t.Setenv("DATABASE_URL", "custom-db-url")
	t.Setenv("REDIS_ADDR", "192.168.0.1:6379")
	t.Setenv("KAFKA_BROKER", "kafka:9092")
	t.Setenv("KAFKA_TOPIC_USERS", "my-users")
	t.Setenv("KAFKA_TOPIC_SEARCH_EVENTS", "my-search")
	t.Setenv("USER_SERVICE_PORT", ":9999")
	t.Setenv("LOG_LEVEL", "error")
	t.Setenv("LOG_PRETTY", "false")

	cfg := Load()

	if cfg.DatabaseURL != "custom-db-url" {
		t.Errorf("expected custom DatabaseURL, got %s", cfg.DatabaseURL)
	}
	if cfg.RedisAddr != "192.168.0.1:6379" {
		t.Errorf("expected custom RedisAddr, got %s", cfg.RedisAddr)
	}
	if cfg.KafkaBroker != "kafka:9092" {
		t.Errorf("expected custom KafkaBroker, got %s", cfg.KafkaBroker)
	}
	if cfg.KafkaTopicUsers != "my-users" {
		t.Errorf("expected custom KafkaTopicUsers, got %s", cfg.KafkaTopicUsers)
	}
	if cfg.UserServicePort != ":9999" {
		t.Errorf("expected custom UserServicePort, got %s", cfg.UserServicePort)
	}
	if cfg.LogLevel != zerolog.ErrorLevel {
		t.Errorf("expected custom LogLevel Error, got %v", cfg.LogLevel)
	}
	if cfg.LogPretty {
		t.Errorf("expected custom LogPretty false, got %v", cfg.LogPretty)
	}
}

func Test_parseLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  zerolog.Level
	}{
		{"debug", "debug", zerolog.DebugLevel},
		{"warn", "warn", zerolog.WarnLevel},
		{"error", "error", zerolog.ErrorLevel},
		{"fatal", "fatal", zerolog.FatalLevel},
		{"info", "info", zerolog.InfoLevel},
		{"empty", "", zerolog.InfoLevel},      // default case
		{"unknown", "foo", zerolog.InfoLevel}, // default case
		{"upper", "DEBUG", zerolog.InfoLevel}, // Ваш switch case чувствителен к регистру (case "debug"), поэтому DEBUG вернет default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLogLevel(tt.input)
			if got != tt.want {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func Test_getEnv(t *testing.T) {
	// Тест приватной функции getEnv
	key := "TEST_ENV_KEY"
	val := "some_value"
	def := "default_value"

	// 1. Переменная не задана
	os.Unsetenv(key)
	if got := getEnv(key, def); got != def {
		t.Errorf("getEnv without env var = %s, want %s", got, def)
	}

	// 2. Переменная задана
	t.Setenv(key, val)
	if got := getEnv(key, def); got != val {
		t.Errorf("getEnv with env var = %s, want %s", got, val)
	}

	// 3. Переменная пустая (должен вернуться дефолт, согласно логике if val != "")
	t.Setenv(key, "")
	if got := getEnv(key, def); got != def {
		t.Errorf("getEnv with empty env var = %s, want %s", got, def)
	}
}
