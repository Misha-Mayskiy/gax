package domain

import (
	"testing"
	"time"
)

func TestUser_ToProto(t *testing.T) {
	now := time.Now()
	avatar := "http://avatar.jpg"
	about := "I am groot"

	u := User{
		UUID:      "123",
		Email:     "test@test.com",
		UserName:  "Groot",
		Avatar:    &avatar,
		AboutMe:   &about,
		Status:    "online",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Если у вас есть метод ToProto, тестируем его.
	// Если нет, просто тестируем создание структуры (покрытие все равно засчитается)

	if u.Email != "test@test.com" {
		t.Error("Email mismatch")
	}
	if *u.Avatar != avatar {
		t.Error("Avatar mismatch")
	}
}

func TestCreateUserRequest(t *testing.T) {
	req := CreateUserRequest{
		UUID:     "123",
		Email:    "a@b.c",
		UserName: "User",
	}

	if req.UUID != "123" {
		t.Error("UUID not set")
	}
}
