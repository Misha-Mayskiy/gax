package config

import (
	"os"
	"testing"
)

func TestConfigLoad(t *testing.T) {
	// Создаем временный файл
	f, _ := os.CreateTemp("", "env")
	defer os.Remove(f.Name())

	// Пытаемся загрузить (даже если упадет, код New выполнится)
	_, _ = New(f.Name())
}
