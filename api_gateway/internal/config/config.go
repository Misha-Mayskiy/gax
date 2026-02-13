package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	Redis    RedisConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port    int
	TimeOut time.Duration
}

type ServicesConfig struct {
	AuthServiceAddr   string
	UserServiceAddr   string
	ChatServiceAddr   string
	RoomServiceAddr   string
	SearchServiceAddr string
	MediaServiceAddr  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type DatabaseConfig struct {
	PostgresDSN string
}

type KafkaConfig struct {
	Brokers string
}

type LogConfig struct {
	Level  string
	Pretty bool
}

func ParseConfig(path string) (*Config, error) {
	// Пытаемся загрузить .env файл
	if err := godotenv.Load(path); err != nil {
		// Если файл не найден, используем переменные окружения системы
		if !os.IsNotExist(err) {
			return nil, err
		}
		// Файл не существует - это нормально, продолжаем
	}

	// Парсим конфигурацию с дефолтными значениями
	return &Config{
		Server: ServerConfig{
			Port:    getIntEnv("HTTP_PORT", 8080),
			TimeOut: getDurationEnv("HTTP_TIMEOUT", 30*time.Second),
		},
		Services: ServicesConfig{
			AuthServiceAddr:   getEnv("AUTH_SERVICE_ADDR", "localhost:8081"),
			UserServiceAddr:   getEnv("USER_SERVICE_ADDR", "localhost:8082"),
			ChatServiceAddr:   getEnv("CHAT_SERVICE_ADDR", "localhost:8083"),
			RoomServiceAddr:   getEnv("ROOM_SERVICE_ADDR", "localhost:50053"),
			SearchServiceAddr: getEnv("SEARCH_SERVICE_ADDR", "http://localhost:8085"),
			MediaServiceAddr:  getEnv("MEDIA_SERVICE_ADDR", "http://localhost:8084"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		Database: DatabaseConfig{
			PostgresDSN: getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/users?sslmode=disable"),
		},
		Kafka: KafkaConfig{
			Brokers: getEnv("KAFKA_BROKERS", "localhost:29092"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Pretty: getBoolEnv("LOG_PRETTY", true),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
		// Также поддерживаем строки "true"/"false"
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if dur, err := time.ParseDuration(value); err == nil {
			return dur
		}
	}
	return defaultValue
}

// ConfigInterface интерфейс для конфигурации
type ConfigInterface interface {
	GetServices() ServicesConfig
}

// Добавьте метод GetServices к существующей структуре Config
func (c *Config) GetServices() ServicesConfig {
	return c.Services
}
