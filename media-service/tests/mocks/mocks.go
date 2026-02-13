package mocks

import (
	"context"
	"io"
	"media-service/internal/domain"

	"github.com/minio/minio-go/v7"
)

// --- Mock Postgres ---
type PgRepo struct {
	SaveFunc    func(ctx context.Context, f *domain.FileMeta) error
	GetByIDFunc func(ctx context.Context, id string) (*domain.FileMeta, error)
	DeleteFunc  func(ctx context.Context, id string) error

	// НОВЫЕ ПОЛЯ для реализации полного интерфейса:
	GetFilesByUserFunc      func(ctx context.Context, userID string, limit, offset int) ([]*domain.FileMeta, error)
	GetTotalFilesByUserFunc func(ctx context.Context, userID string) (int, error)
}

func (m *PgRepo) Save(ctx context.Context, f *domain.FileMeta) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, f)
	}
	return nil
}

func (m *PgRepo) GetByID(ctx context.Context, id string) (*domain.FileMeta, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *PgRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// НОВЫЕ МЕТОДЫ:

func (m *PgRepo) GetFilesByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.FileMeta, error) {
	if m.GetFilesByUserFunc != nil {
		return m.GetFilesByUserFunc(ctx, userID, limit, offset)
	}
	return nil, nil
}

func (m *PgRepo) GetTotalFilesByUser(ctx context.Context, userID string) (int, error) {
	if m.GetTotalFilesByUserFunc != nil {
		return m.GetTotalFilesByUserFunc(ctx, userID)
	}
	return 0, nil
}

// --- Mock MinIO ---
type MinioRepo struct {
	UploadFunc func() error
	DeleteFunc func() error
}

func (m *MinioRepo) Upload(ctx context.Context, bucket, name string, r io.Reader, size int64, cType string) error {
	if m.UploadFunc != nil {
		return m.UploadFunc()
	}
	return nil
}
func (m *MinioRepo) Download(ctx context.Context, bucket, name string) (*minio.Object, error) {
	return nil, nil
}

func (m *MinioRepo) Delete(ctx context.Context, bucket, name string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc()
	}
	return nil
}

// --- Mock Kafka ---
type Kafka struct{}

func (m *Kafka) SendFileUploaded(meta *domain.FileMeta) error { return nil }
