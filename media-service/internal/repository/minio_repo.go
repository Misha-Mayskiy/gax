package repository

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioRepo struct {
	client *minio.Client
}

func NewMinioRepo(endpoint, access, secret string) (*MinioRepo, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(access, secret, ""),
		Secure: false, // Внутри докера обычно HTTP, SSL не нужен
	})
	if err != nil {
		return nil, err
	}
	return &MinioRepo{client: client}, nil
}

// Upload загружает поток данных в бакет
func (m *MinioRepo) Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) error {
	// Проверка существования бакета (lazy init)
	exists, err := m.client.BucketExists(ctx, bucket)
	if err == nil && !exists {
		_ = m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}

	_, err = m.client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// Download возвращает объект (читатель)
func (m *MinioRepo) Download(ctx context.Context, bucket, objectName string) (*minio.Object, error) {
	return m.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
}

// Delete удаляет файл (нужен для отката транзакции, если БД упала)
func (m *MinioRepo) Delete(ctx context.Context, bucket, objectName string) error {
	return m.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}
