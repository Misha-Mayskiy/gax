package service

import (
	"bytes"
	"context"
	"media-service/internal/domain"
	"media-service/tests/mocks"
	"mime/multipart"
	"testing"
)

func TestUploadFlow(t *testing.T) {
	// 1. Подготовка (Arrange)
	mockPg := &mocks.PgRepo{
		SaveFunc: func(ctx context.Context, f *domain.FileMeta) error {
			if f.Filename != "avatar.jpg" {
				t.Errorf("Expected filename avatar.jpg, got %s", f.Filename)
			}
			return nil
		},
	}

	mockMinio := &mocks.MinioRepo{
		UploadFunc: func() error { return nil },
	}

	mockKafka := &mocks.Kafka{}

	svc := NewMediaService(mockPg, mockMinio, mockKafka)

	// Создаем файл для загрузки
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "avatar.jpg")
	part.Write([]byte("image data"))
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(1024)
	fileHeader := form.File["file"][0]
	file, _ := fileHeader.Open()
	defer file.Close()

	// 2. Действие (Act)
	// ИСПРАВЛЕНИЕ: Добавляем недостающие аргументы
	userID := "user-123"
	desc := "test description"
	chatID := "chat-1"

	result, err := svc.UploadFile(context.Background(), file, fileHeader, userID, desc, chatID)

	// 3. Проверка (Assert)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID == "" {
		t.Error("Generated ID should not be empty")
	}

	if result.Bucket != "files" {
		t.Errorf("Expected bucket 'files', got %s", result.Bucket)
	}

	// Можно дополнительно проверить, что новые поля проставились
	if result.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, result.UserID)
	}
}

func TestDeleteFlow(t *testing.T) {
	// 1. Arrange
	targetID := "file-123"

	mockPg := &mocks.PgRepo{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.FileMeta, error) {
			return &domain.FileMeta{
				ID:         targetID,
				Bucket:     "files",
				ObjectName: targetID,
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id string) error {
			if id != targetID {
				t.Errorf("Expected delete ID %s, got %s", targetID, id)
			}
			return nil
		},
	}

	mockMinio := &mocks.MinioRepo{
		DeleteFunc: func() error {
			return nil
		},
	}

	mockKafka := &mocks.Kafka{}

	svc := NewMediaService(mockPg, mockMinio, mockKafka)

	// 2. Act
	err := svc.DeleteFile(context.Background(), targetID)

	// 3. Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
