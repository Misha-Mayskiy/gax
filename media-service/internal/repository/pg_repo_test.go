package repository

import (
	"context"
	"testing"
	"time"

	"media-service/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPgRepo_Save(t *testing.T) {
	// 1. Создаем фейковую БД
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPgRepo(db)

	// 2. Данные для теста
	file := &domain.FileMeta{
		ID:          "uuid-123",
		Filename:    "test.txt",
		Bucket:      "files",
		ObjectName:  "obj-123",
		ContentType: "text/plain",
		Size:        1024,
		CreatedAt:   time.Now(),
		UserID:      "user-1",
		Description: "test desc",
		ChatID:      "chat-1",
	}

	// 3. Ожидание (Expectation)
	mock.ExpectExec("INSERT INTO files").
		WithArgs(
			file.ID,
			file.Filename,
			file.Bucket,
			file.ObjectName,
			file.ContentType,
			file.Size,
			file.CreatedAt,
			file.UserID,      // <-- Добавлено
			file.Description, // <-- Добавлено
			file.ChatID,      // <-- Добавлено
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 4. Действие
	if err := repo.Save(context.Background(), file); err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// 5. Проверка, что все ожидания сбылись
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPgRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewPgRepo(db)
	id := "uuid-123"

	// Настраиваем, что БД должна вернуть строки (Rows)
	rows := sqlmock.NewRows([]string{
		"id", "filename", "bucket", "object_name", "content_type", "size", "created_at",
		"user_id", "description", "chat_id", // <-- Новые колонки
	}).
		AddRow(
			"uuid-123", "test.txt", "files", "obj-123", "text/plain", 1024, time.Now(),
			"user-1", "desc", "chat-1", // <-- Значения для новых колонок
		)

	mock.ExpectQuery("SELECT id, filename, bucket"). // sqlmock проверяет по частичному совпадению, этого достаточно
								WithArgs(id).
								WillReturnRows(rows)

	res, err := repo.GetByID(context.Background(), id)
	if err != nil {
		t.Errorf("error not expected: %s", err)
	}

	if res.ID != id {
		t.Errorf("expected id %s, got %s", id, res.ID)
	}
	// Можно также проверить новые поля, если нужно
	if res.UserID != "user-1" {
		t.Errorf("expected user_id user-1, got %s", res.UserID)
	}
}

func TestPgRepo_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewPgRepo(db)
	id := "uuid-123"

	// Ожидаем DELETE запрос (тут ничего не менялось, так как удаляем по ID)
	mock.ExpectExec("DELETE FROM files WHERE id = \\$1").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	err = repo.Delete(context.Background(), id)
	if err != nil {
		t.Errorf("error not expected: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
