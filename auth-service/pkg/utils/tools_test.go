package utils

import (
	"testing"
)

func TestGeneratePasswordHash(t *testing.T) {
	password := "secret123"

	hash, err := GeneratePasswordHash(password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not be equal to password")
	}
}

func TestComparePassword(t *testing.T) {
	password := "secret123"
	wrongPassword := "wrong123"

	hash, _ := GeneratePasswordHash(password)

	// Тест 1: Правильный пароль
	err := ComparePassword(hash, password)
	if err != nil {
		t.Errorf("Expected match, got error: %v", err)
	}

	// Тест 2: Неправильный пароль
	err = ComparePassword(hash, wrongPassword)
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}

	// Тест 3: Сравнение строки (функция Compare)
	err = Compare(hash, password)
	if err != nil {
		t.Error("Compare function failed")
	}
}
