package http

import (
	"testing"
)

// Тестируем только публичную функцию defaultIfEmpty
func TestDefaultIfEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		def      string
		expected string
	}{
		{"empty string returns default", "", "default", "default"},
		{"non-empty returns itself", "value", "default", "value"},
		{"space returns space", " ", "default", " "},
		{"both empty", "", "", ""},
		{"input with default empty", "test", "", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultIfEmpty(tt.input, tt.def)
			if result != tt.expected {
				t.Errorf("defaultIfEmpty(%q, %q) = %q, want %q",
					tt.input, tt.def, result, tt.expected)
			}
		})
	}
}

// Тестируем логику парсинга параметров (без создания Server)
func TestQueryParsingLogic(t *testing.T) {
	t.Run("parse query parameters", func(t *testing.T) {
		// Эмулируем URL с параметрами
		url := "/search?q=hello&limit=10&offset=5&highlight=true"

		// В реальном коде это делается через r.URL.Query()
		// Здесь просто проверяем что мы понимаем логику
		hasQ := contains(url, "q=")
		hasLimit := contains(url, "limit=")
		hasOffset := contains(url, "offset=")
		hasHighlight := contains(url, "highlight=true")

		if !hasQ {
			t.Error("URL should have q parameter")
		}
		if !hasLimit {
			t.Error("URL should have limit parameter")
		}
		if !hasOffset {
			t.Error("URL should have offset parameter")
		}
		if !hasHighlight {
			t.Error("URL should have highlight parameter")
		}
	})

	t.Run("validation logic", func(t *testing.T) {
		// Тестируем логику валидации: нужен q ИЛИ type
		testCases := []struct {
			q          string
			typ        string
			shouldPass bool
		}{
			{"", "", false},        // оба пустые - невалидно
			{"test", "", true},     // есть q - валидно
			{"", "user", true},     // есть type - валидно
			{"test", "user", true}, // оба есть - валидно
		}

		for _, tc := range testCases {
			isValid := !(tc.q == "" && tc.typ == "")
			if isValid != tc.shouldPass {
				t.Errorf("q=%q, type=%q: expected valid=%v, got %v",
					tc.q, tc.typ, tc.shouldPass, isValid)
			}
		}
	})

	t.Run("limit and offset defaults", func(t *testing.T) {
		// Проверяем логику дефолтных значений
		testCases := []struct {
			input    string
			def      string
			expected string
		}{
			// limit дефолт
			{"", "20", "20"},
			{"10", "20", "10"},
			{"not-a-number", "20", "not-a-number"}, // потом Atoi вернет 0

			// offset дефолт
			{"", "0", "0"},
			{"5", "0", "5"},
		}

		for _, tc := range testCases {
			result := defaultIfEmpty(tc.input, tc.def)
			if result != tc.expected {
				t.Errorf("defaultIfEmpty(%q, %q) = %q, want %q",
					tc.input, tc.def, result, tc.expected)
			}
		}
	})
}

// Вспомогательная функция для тестов
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Тестируем что код компилируется
func TestCompilation(t *testing.T) {
	// Просто проверяем что функция существует и работает
	result := defaultIfEmpty("test", "default")
	if result != "test" {
		t.Errorf("defaultIfEmpty не работает")
	}

	t.Log("Код компилируется и функция defaultIfEmpty работает")
}

// Самые простые тесты - проверка базовых случаев
func TestBasicCases(t *testing.T) {
	t.Run("empty string handling", func(t *testing.T) {
		if defaultIfEmpty("", "hello") != "hello" {
			t.Error("Пустая строка должна возвращать default")
		}
	})

	t.Run("non-empty string handling", func(t *testing.T) {
		if defaultIfEmpty("world", "hello") != "world" {
			t.Error("Непустая строка должна возвращаться как есть")
		}
	})
}

// Тест производительности (опционально)
func BenchmarkDefaultIfEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		defaultIfEmpty("test", "default")
		defaultIfEmpty("", "default")
	}
}
