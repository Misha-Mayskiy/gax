package mocks

import (
	"context"
	"errors"
	"media-service/internal/domain"
	"testing"
)

func TestPgRepo_Coverage(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("test error")

	t.Run("Save", func(t *testing.T) {
		m := &PgRepo{}
		// 1. Default (nil)
		if err := m.Save(ctx, nil); err != nil {
			t.Error("Expected nil error when func is nil")
		}
		// 2. Custom
		m.SaveFunc = func(_ context.Context, _ *domain.FileMeta) error { return testErr }
		if err := m.Save(ctx, nil); err != testErr {
			t.Error("Expected custom error")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		m := &PgRepo{}
		// 1. Default
		res, err := m.GetByID(ctx, "id")
		if res != nil || err != nil {
			t.Error("Expected nil result and nil error")
		}
		// 2. Custom
		m.GetByIDFunc = func(_ context.Context, _ string) (*domain.FileMeta, error) {
			return &domain.FileMeta{ID: "123"}, nil
		}
		res, _ = m.GetByID(ctx, "id")
		if res.ID != "123" {
			t.Error("Expected custom object")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		m := &PgRepo{}
		// 1. Default
		if err := m.Delete(ctx, "id"); err != nil {
			t.Error("Expected nil error")
		}
		// 2. Custom
		m.DeleteFunc = func(_ context.Context, _ string) error { return testErr }
		if err := m.Delete(ctx, "id"); err != testErr {
			t.Error("Expected custom error")
		}
	})

	t.Run("GetFilesByUser", func(t *testing.T) {
		m := &PgRepo{}
		// 1. Default
		res, err := m.GetFilesByUser(ctx, "u", 10, 0)
		if res != nil || err != nil {
			t.Error("Expected nil result")
		}
		// 2. Custom
		m.GetFilesByUserFunc = func(_ context.Context, _ string, _, _ int) ([]*domain.FileMeta, error) {
			return []*domain.FileMeta{{ID: "1"}}, nil
		}
		res, _ = m.GetFilesByUser(ctx, "u", 10, 0)
		if len(res) != 1 {
			t.Error("Expected 1 file")
		}
	})

	t.Run("GetTotalFilesByUser", func(t *testing.T) {
		m := &PgRepo{}
		// 1. Default
		count, err := m.GetTotalFilesByUser(ctx, "u")
		if count != 0 || err != nil {
			t.Error("Expected 0 count")
		}
		// 2. Custom
		m.GetTotalFilesByUserFunc = func(_ context.Context, _ string) (int, error) {
			return 5, nil
		}
		count, _ = m.GetTotalFilesByUser(ctx, "u")
		if count != 5 {
			t.Error("Expected 5 count")
		}
	})
}

func TestMinioRepo_Coverage(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("minio error")

	t.Run("Upload", func(t *testing.T) {
		m := &MinioRepo{}
		// 1. Default
		if err := m.Upload(ctx, "b", "n", nil, 0, ""); err != nil {
			t.Error("Expected nil error")
		}
		// 2. Custom
		m.UploadFunc = func() error { return testErr }
		if err := m.Upload(ctx, "b", "n", nil, 0, ""); err != testErr {
			t.Error("Expected custom error")
		}
	})

	t.Run("Download", func(t *testing.T) {
		m := &MinioRepo{}
		// Download обычно заглушен жестко
		obj, err := m.Download(ctx, "b", "n")
		if obj != nil || err != nil {
			t.Error("Download should always return nil, nil in current mock impl")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		m := &MinioRepo{}
		// 1. Default
		if err := m.Delete(ctx, "b", "n"); err != nil {
			t.Error("Expected nil error")
		}
		// 2. Custom
		m.DeleteFunc = func() error { return testErr }
		if err := m.Delete(ctx, "b", "n"); err != testErr {
			t.Error("Expected custom error")
		}
	})
}

func TestKafka_Coverage(t *testing.T) {
	m := &Kafka{}
	// Kafka мок жестко заглушен
	if err := m.SendFileUploaded(nil); err != nil {
		t.Error("Expected nil error")
	}
}
