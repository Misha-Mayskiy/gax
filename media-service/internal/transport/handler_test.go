package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"media-service/internal/domain"
	"media-service/internal/service"

	"media-service/tests/mocks"

	"github.com/gorilla/mux"
)

func TestUploadHandler(t *testing.T) {
	// 1. Инициализация (Arrange)

	// Используем моки из tests/mocks/mocks.go
	mockPg := &mocks.PgRepo{
		SaveFunc: func(ctx context.Context, f *domain.FileMeta) error { return nil },
	}
	mockMinio := &mocks.MinioRepo{
		UploadFunc: func() error { return nil },
	}
	mockKafka := &mocks.Kafka{}

	svc := service.NewMediaService(mockPg, mockMinio, mockKafka)

	// 2. Создаем Хендлер (Act)
	handler := NewHandler(svc)

	// 3. Тестируем (Assert)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "cat.png")
	part.Write([]byte("meow content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/media/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	// Вызываем метод
	handler.Upload(rr, req)

	// Проверки
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var resp domain.FileMeta
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal("Response is not valid JSON")
	}

	if resp.Filename != "cat.png" {
		t.Errorf("Expected filename cat.png, got %s", resp.Filename)
	}
}

func TestDeleteHandler(t *testing.T) {
	// 1. Arrange
	mockPg := &mocks.PgRepo{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.FileMeta, error) {
			return &domain.FileMeta{ID: id, Bucket: "files", ObjectName: id}, nil
		},
		DeleteFunc: func(ctx context.Context, id string) error { return nil },
	}
	mockMinio := &mocks.MinioRepo{
		DeleteFunc: func() error { return nil },
	}
	mockKafka := &mocks.Kafka{}

	svc := service.NewMediaService(mockPg, mockMinio, mockKafka)
	handler := NewHandler(svc)

	// 2. Act
	req := httptest.NewRequest("DELETE", "/media/delete/123", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/media/delete/{id}", handler.DeleteFile)

	router.ServeHTTP(rr, req)

	// 3. Assert
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
}
