package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"api_gateway/pkg/logger"

	"go.uber.org/zap"
)

func (h *HandlerFacade) UploadFile(w http.ResponseWriter, r *http.Request) {
	log.Printf("======= UPLOAD DEBUG START =======")
	log.Printf("Proxying upload to media-service")

	// Парсим форму для получения данных
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("user_id")
	log.Printf("User ID from form: %s", userID)

	// 1. Сначала проверим, что media-service доступен
	mediaURL := os.Getenv("MEDIA_SERVICE_URL")
	if mediaURL == "" {
		mediaURL = "http://media-service:8084"
	}

	log.Printf("Media service URL: %s", mediaURL)

	// Проверяем доступность через /health
	client := &http.Client{Timeout: 5 * time.Second}
	healthResp, err := client.Get(mediaURL + "/health")
	if err != nil {
		log.Printf("Health check FAILED: %v", err)

		// В режиме разработки возвращаем тестовый ответ
		log.Printf("Media service unavailable, returning test response")

		w.Header().Set("Content-Type", "application/json")

		file, header, _ := r.FormFile("file")
		if file != nil {
			defer file.Close()
		}

		description := r.FormValue("description")
		chatID := r.FormValue("chat_id")

		var filename string
		var fileSize int64
		var contentType string

		if header != nil {
			filename = header.Filename
			fileSize = header.Size
			contentType = header.Header.Get("Content-Type")
		} else {
			filename = "test-file.txt"
			fileSize = 1024
			contentType = "text/plain"
		}

		response := map[string]interface{}{
			"id":           fmt.Sprintf("test-file-%d", time.Now().UnixNano()),
			"filename":     filename,
			"size":         fileSize,
			"content_type": contentType,
			"user_id":      userID,
			"description":  description,
			"chat_id":      chatID,
			"uploaded_at":  time.Now().Format(time.RFC3339),
			"url":          fmt.Sprintf("/uploads/%s", filename),
			"success":      true,
			"debug":        "media-service unavailable, using test mode",
		}

		json.NewEncoder(w).Encode(response)
		return
	}
	healthResp.Body.Close()
	log.Printf("Health check OK: %d", healthResp.StatusCode)

	// 2. Создаем прокси-запрос к media-service
	// Сначала получаем файл
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Failed to get file from form: %v", err)
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("File received: %s, size: %d", header.Filename, header.Size)

	// Создаем multipart запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		log.Printf("Failed to create form file: %v", err)
		http.Error(w, "Failed to create form file", http.StatusInternalServerError)
		return
	}

	// Копируем файл
	bytesCopied, err := io.Copy(part, file)
	if err != nil {
		log.Printf("Failed to copy file: %v", err)
		http.Error(w, "Failed to copy file", http.StatusInternalServerError)
		return
	}
	log.Printf("Copied %d bytes to buffer", bytesCopied)

	// Добавляем дополнительные поля
	for key, values := range r.Form {
		for _, value := range values {
			writer.WriteField(key, value)
			log.Printf("Added form field: %s=%s", key, value)
		}
	}

	err = writer.Close()
	if err != nil {
		log.Printf("Failed to close writer: %v", err)
		http.Error(w, "Failed to close writer", http.StatusInternalServerError)
		return
	}

	log.Printf("Buffer size: %d bytes", body.Len())
	log.Printf("Content-Type boundary: %s", writer.FormDataContentType())

	// Создаем прокси запрос
	targetURL := mediaURL + "/upload"
	log.Printf("Target URL: %s", targetURL)

	proxyReq, err := http.NewRequest("POST", targetURL, body)
	if err != nil {
		log.Printf("Failed to create proxy request: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	proxyReq.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем запрос
	client.Timeout = 30 * time.Second
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Failed to proxy to media-service: %v", err)
		http.Error(w, "Media service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	log.Printf("Proxy response status: %d", resp.StatusCode)

	// 3. Проксируем ответ напрямую
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	// Копируем тело ответа
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Failed to copy response body: %v", err)
	}

	log.Printf("======= UPLOAD DEBUG END =======")
}

// DownloadFile - скачивание файла
func (h *HandlerFacade) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	stream, filename, err := h.mediaService.DownloadFile(r.Context(), fileID)
	if err != nil {
		logger.GetLoggerFromCtx(h.ctx).Error(h.ctx, "Download failed", zap.Error(err))
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer stream.Close()

	// Устанавливаем заголовки
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")

	// Копируем файл в ответ
	io.Copy(w, stream)
}

// GetFileMeta - получение метаданных файла
func (h *HandlerFacade) GetFileMeta(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	meta, err := h.mediaService.GetFileMeta(r.Context(), fileID)
	if err != nil {
		logger.GetLoggerFromCtx(h.ctx).Error(h.ctx, "Get metadata failed", zap.Error(err))
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}

// DeleteFile - удаление файла
func (h *HandlerFacade) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("id")
	userID := r.URL.Query().Get("user_id") // Для проверки прав

	if fileID == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	err := h.mediaService.DeleteFile(r.Context(), fileID, userID)
	if err != nil {
		logger.GetLoggerFromCtx(h.ctx).Error(h.ctx, "Delete failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// ListUserFiles - список файлов пользователя
// ListUserFiles - список файлов пользователя (прямой прокси в media-service)
func (h *HandlerFacade) ListUserFiles(w http.ResponseWriter, r *http.Request) {
	log.Printf("======= LIST FILES PROXY START =======")

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	log.Printf("List files for user: %s", userID)

	// 1. Создаем прокси-запрос к media-service
	mediaURL := os.Getenv("MEDIA_SERVICE_URL")
	if mediaURL == "" {
		mediaURL = "http://media-service:8084"
	}

	// Сохраняем оригинальные параметры запроса
	query := r.URL.Query()
	targetURL := mediaURL + "/media/list?" + query.Encode()

	log.Printf("Target URL: %s", targetURL)

	// 2. Создаем прокси-запрос
	proxyReq, err := http.NewRequest(r.Method, targetURL, nil)
	if err != nil {
		log.Printf("Failed to create proxy request: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Копируем оригинальные заголовки
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	log.Printf("Proxy request to: %s", targetURL)

	// 3. Отправляем запрос
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Failed to proxy to media-service: %v", err)

		// Возвращаем пустой список вместо ошибки
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		if limit == 0 {
			limit = 50
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"files":  []interface{}{},
			"total":  0,
			"limit":  limit,
			"offset": offset,
		})
		return
	}
	defer resp.Body.Close()

	log.Printf("Proxy response status: %d", resp.StatusCode)

	// 4. Проксируем ответ
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	// Копируем тело ответа
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	// Логируем тело ответа для отладки
	log.Printf("Response body: %s", string(bodyBytes))

	// Проверяем, что это валидный JSON
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonResponse); err != nil {
		log.Printf("Invalid JSON response: %v", err)
		// Возвращаем пустой список
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		if limit == 0 {
			limit = 50
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"files":  []interface{}{},
			"total":  0,
			"limit":  limit,
			"offset": offset,
		})
		return
	}

	// Записываем тело ответа
	w.Write(bodyBytes)

	log.Printf("======= LIST FILES PROXY END =======")
}
