package repository

import (
	"bytes"
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
)

// setupFakeS3 поднимает in-memory S3 сервер и возвращает URL и backend для проверок
func setupFakeS3() (*httptest.Server, *s3mem.Backend) {
	// Создаем backend в памяти
	backend := s3mem.New()
	faker := gofakes3.New(backend)

	// Запускаем HTTP сервер, который притворяется S3
	ts := httptest.NewServer(faker.Server())
	return ts, backend
}

func TestMinioRepo_Upload(t *testing.T) {
	// 1. Arrange
	ts, backend := setupFakeS3()
	defer ts.Close()

	// Парсим URL тестового сервера, чтобы достать host:port
	u, _ := url.Parse(ts.URL)
	endpoint := u.Host // Minio клиент требует endpoint без http://

	// Инициализируем наш репозиторий, направляя его на фейковый сервер
	repo, err := NewMinioRepo(endpoint, "accessKey", "secretKey")
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	bucket := "test-bucket"
	objectName := "test.txt"
	content := []byte("hello world")
	size := int64(len(content))

	// 2. Act
	// Тестируем Upload. Внутри должна сработать логика создания бакета (MakeBucket), т.к. его нет
	err = repo.Upload(context.Background(), bucket, objectName, bytes.NewReader(content), size, "text/plain")

	// 3. Assert
	if err != nil {
		t.Errorf("Upload returned error: %v", err)
	}

	// Проверяем "напрямую" в бэкенде, появился ли файл
	obj, err := backend.HeadObject(bucket, objectName)
	if err != nil {
		t.Fatalf("Object was not uploaded to backend: %v", err)
	}
	if obj.Size != size {
		t.Errorf("Expected size %d, got %d", size, obj.Size)
	}
}

func TestMinioRepo_Delete(t *testing.T) {
	// 1. Arrange
	ts, backend := setupFakeS3()
	defer ts.Close()
	u, _ := url.Parse(ts.URL)

	repo, _ := NewMinioRepo(u.Host, "accessKey", "secretKey")

	bucket := "files"
	objectName := "todelete.jpg"

	// Кладем файл
	backend.CreateBucket(bucket)
	backend.PutObject(bucket, objectName, map[string]string{}, bytes.NewReader([]byte{1, 2, 3}), int64(3), nil)

	// Убеждаемся, что он там
	if _, err := backend.HeadObject(bucket, objectName); err != nil {
		t.Fatal("Setup failed: object not found")
	}

	// 2. Act
	err := repo.Delete(context.Background(), bucket, objectName)
	if err != nil {
		t.Errorf("Delete returned error: %v", err)
	}

	// 3. Assert
	// Проверяем, что файл исчез
	if _, err := backend.HeadObject(bucket, objectName); err == nil {
		t.Error("Object should be deleted, but it still exists")
	}
}

func TestNewMinioRepo_Error(t *testing.T) {
	// Тест на некорректный эндпоинт (сложно заставить minio.New вернуть ошибку,
	// но можно проверить, что он возвращает инициализированную структуру)

	repo, err := NewMinioRepo("localhost:9000", "u", "p")
	if err != nil {
		t.Fatalf("NewMinioRepo returned error: %v", err)
	}
	if repo == nil || repo.client == nil {
		t.Error("Repo or client is nil")
	}
}
