package logger

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestInit(t *testing.T) {
	// Сохраняем текущий уровень
	originalLevel := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(originalLevel)

	tests := []struct {
		name   string
		level  zerolog.Level
		pretty bool
	}{
		{
			name:   "JSON format with Info level",
			level:  zerolog.InfoLevel,
			pretty: false,
		},
		{
			name:   "Pretty format with Debug level",
			level:  zerolog.DebugLevel,
			pretty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.level, tt.pretty)
			if zerolog.GlobalLevel() != tt.level {
				t.Errorf("Expected global level %v, got %v", tt.level, zerolog.GlobalLevel())
			}
			GetLogger()
		})
	}
}

func TestWrappers(t *testing.T) {
	// Инициализируем логгер перед использованием оберток
	Init(zerolog.DebugLevel, false)

	// Тестируем GetLogger
	l := GetLogger()
	l.Info().Msg("Test message from GetLogger")

	if e := Debug(); e == nil {
		t.Error("Debug() returned nil")
	} else {
		e.Msg("debug msg")
	}

	if e := Info(); e == nil {
		t.Error("Info() returned nil")
	} else {
		e.Msg("info msg")
	}

	if e := Warn(); e == nil {
		t.Error("Warn() returned nil")
	} else {
		e.Msg("warn msg")
	}

	if e := Error(); e == nil {
		t.Error("Error() returned nil")
	} else {
		e.Msg("error msg")
	}

	if e := Fatal(); e == nil {
		t.Error("Fatal() returned nil")
	}

	if e := Panic(); e == nil {
		t.Error("Panic() returned nil")
	}
}

// Дополнительный тест, для покрытия внутри zerolog
func TestPanicExecution(t *testing.T) {
	Init(zerolog.DebugLevel, false)

	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic")
		}
	}()

	Panic().Msg("This should panic")
}
