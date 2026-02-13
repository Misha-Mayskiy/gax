// api_gateway/internal/service/media_service.go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"api_gateway/pkg/logger"

	"go.uber.org/zap"
)

// MediaService - структура для работы с медиа-сервисом
type MediaService struct {
	addr   string
	client *http.Client
}

// NewMediaService создает новый MediaService
func NewMediaService(addr string) *MediaService {
	return &MediaService{
		addr: addr,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewMediaServiceForTest создает MediaService для тестов
func NewMediaServiceForTest(addr string, client *http.Client) *MediaService {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return &MediaService{
		addr:   addr,
		client: client,
	}
}

func (s *MediaService) UploadFile(ctx context.Context, file io.Reader, filename, contentType, userID, chatID string) (map[string]any, error) {
	// Создаем multipart запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавляем файл
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	io.Copy(part, file)

	// Добавляем дополнительные поля
	if userID != "" {
		writer.WriteField("user_id", userID)
	}
	if chatID != "" {
		writer.WriteField("chat_id", chatID)
	}

	writer.Close()

	// Отправляем запрос к Media Service
	url := fmt.Sprintf("%s/upload", s.addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "MediaService upload request failed", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *MediaService) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, string, error) {
	url := fmt.Sprintf("%s/download/%s", s.addr, fileID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "MediaService download request failed", zap.Error(err))
		return nil, "", err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	filename := resp.Header.Get("Content-Disposition")
	return resp.Body, filename, nil
}

func (s *MediaService) GetFileMeta(ctx context.Context, fileID string) (map[string]any, error) {
	url := fmt.Sprintf("%s/media/%s", s.addr, fileID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "MediaService get meta request failed", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get meta failed with status: %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *MediaService) DeleteFile(ctx context.Context, fileID, userID string) error {
	url := fmt.Sprintf("%s/media/delete/%s", s.addr, fileID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	// Если нужно передать userID для проверки прав
	if userID != "" {
		q := req.URL.Query()
		q.Add("user_id", userID)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "MediaService delete request failed", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (s *MediaService) ListUserFiles(ctx context.Context, userID string, limit, offset int) (map[string]any, error) {
	if s.addr == "" {
		// Если адрес медиа-сервиса не указан, возвращаем пустой список
		return map[string]any{
			"files":  []interface{}{},
			"total":  0,
			"limit":  limit,
			"offset": offset,
		}, nil
	}

	url := fmt.Sprintf("%s/media/list?user_id=%s&limit=%d&offset=%d",
		s.addr, userID, limit, offset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "MediaService list files request failed", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	// Если медиа-сервис возвращает 404 (файлов нет), это не ошибка
	if resp.StatusCode == http.StatusNotFound {
		return map[string]any{
			"files":  []interface{}{},
			"total":  0,
			"limit":  limit,
			"offset": offset,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list files failed with status: %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
