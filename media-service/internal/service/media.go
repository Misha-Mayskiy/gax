package service

import (
	"context"
	"io"
	"log"
	"media-service/internal/domain"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

const (
	BucketName = "files"
	KafkaTopic = "file.events"
)

type MediaService struct {
	pg    FileMetadataRepository
	minio BlobStorage
	kafka EventProducer
}

func NewMediaService(pg FileMetadataRepository, minio BlobStorage, kafka EventProducer) *MediaService {
	return &MediaService{pg: pg, minio: minio, kafka: kafka}
}

func (s *MediaService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader,
	userID, description, chatID string) (*domain.FileMeta, error) {

	log.Printf("SERVICE: Uploading file for user: %s, description: %s", userID, description)

	id := uuid.New().String()

	meta := &domain.FileMeta{
		ID:          id,
		Filename:    header.Filename,
		Bucket:      BucketName,
		ObjectName:  id,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		CreatedAt:   time.Now(),
		UserID:      userID,      // ← ТЕПЕРЬ ЗАПОЛНЯЕТСЯ!
		Description: description, // ← ТЕПЕРЬ ЗАПОЛНЯЕТСЯ!
		ChatID:      chatID,      // ← ТЕПЕРЬ ЗАПОЛНЯЕТСЯ!
	}

	log.Printf("SERVICE: Created metadata - ID: %s, UserID: %s", meta.ID, meta.UserID)

	// 1. MinIO
	if err := s.minio.Upload(ctx, meta.Bucket, meta.ObjectName, file, meta.Size, meta.ContentType); err != nil {
		return nil, err
	}

	// 2. Postgres
	if err := s.pg.Save(ctx, meta); err != nil {
		log.Printf("SERVICE: DB Save error: %v", err)
		return nil, err
	}

	// 3. Kafka Event (Async)
	if s.kafka != nil {
		go func() {
			err := s.kafka.SendFileUploaded(meta)
			if err != nil {
				log.Printf("Kafka send error: %v", err)
			}
		}()
	}

	return meta, nil
}

func (s *MediaService) DownloadFile(ctx context.Context, id string) (*domain.FileMeta, io.ReadCloser, error) {
	// 1. Метаданные
	meta, err := s.pg.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// 2. Стрим
	stream, err := s.minio.Download(ctx, meta.Bucket, meta.ObjectName)
	if err != nil {
		return nil, nil, err
	}

	return meta, stream, nil
}

func (s *MediaService) GetMeta(ctx context.Context, id string) (*domain.FileMeta, error) {
	return s.pg.GetByID(ctx, id)
}

func (s *MediaService) DeleteFile(ctx context.Context, id string) error {
	meta, err := s.pg.GetByID(ctx, id)
	if err != nil {
		return err
	}

	err = s.minio.Delete(ctx, meta.Bucket, meta.ObjectName)
	if err != nil {
		return err
	}

	return s.pg.Delete(ctx, id)
}
func (s *MediaService) ListUserFiles(ctx context.Context, userID string, limit, offset int) (*domain.FileList, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	files, err := s.pg.GetFilesByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.pg.GetTotalFilesByUser(ctx, userID)
	if err != nil {
		// Если ошибка при получении total, используем количество файлов как total
		total = len(files)
	}

	return &domain.FileList{
		Files:  files,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
