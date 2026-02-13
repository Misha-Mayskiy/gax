package service

import (
	"context"
	"io"
	"media-service/internal/domain"

	"github.com/minio/minio-go/v7"
)

// Интерфейс для работы с Метаданными (БД)
type FileMetadataRepository interface {
	Save(ctx context.Context, f *domain.FileMeta) error
	GetByID(ctx context.Context, id string) (*domain.FileMeta, error)
	Delete(ctx context.Context, id string) error
	GetFilesByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.FileMeta, error)
	GetTotalFilesByUser(ctx context.Context, userID string) (int, error)
}

// Интерфейс для работы с Байтами (S3)
type BlobStorage interface {
	Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, bucket, objectName string) (*minio.Object, error)
	Delete(ctx context.Context, bucket, objectName string) error
}

// Интерфейс для Kafka
type EventProducer interface {
	SendFileUploaded(meta *domain.FileMeta) error
}
