package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"media-service/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	svc *service.MediaService
}

func NewHandler(svc *service.MediaService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	// Ограничение 50 МБ
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		log.Printf("UPLOAD ERROR: Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// ПОЛУЧАЕМ ПОЛЯ ИЗ ФОРМЫ
	userID := r.FormValue("user_id")
	description := r.FormValue("description")
	chatID := r.FormValue("chat_id")

	log.Printf("UPLOAD HANDLER: Form values - user_id: %s, description: %s, chat_id: %s",
		userID, description, chatID)

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("UPLOAD ERROR: Invalid file: %v", err)
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("UPLOAD HANDLER: Processing file: %s, size: %d", header.Filename, header.Size)

	// ВЫЗЫВАЕМ ИСПРАВЛЕННЫЙ МЕТОД С ПАРАМЕТРАМИ
	meta, err := h.svc.UploadFile(r.Context(), file, header, userID, description, chatID)
	if err != nil {
		log.Printf("UPLOAD ERROR: Service error: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	// ВАЖНО: Устанавливаем правильный Content-Type
	w.Header().Set("Content-Type", "application/json")

	log.Printf("UPLOAD HANDLER: Upload successful - ID: %s, UserID: %s", meta.ID, meta.UserID)

	// Формируем полный ответ
	response := map[string]interface{}{
		"id":           meta.ID,
		"filename":     meta.Filename,
		"bucket":       meta.Bucket,
		"object_name":  meta.ObjectName,
		"content_type": meta.ContentType,
		"size":         meta.Size,
		"created_at":   meta.CreatedAt.Format(time.RFC3339),
		"user_id":      meta.UserID, // ← ВАЖНО: возвращаем user_id!
		"description":  meta.Description,
		"chat_id":      meta.ChatID,
		"success":      true,
	}

	json.NewEncoder(w).Encode(response)
}
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	meta, stream, err := h.svc.DownloadFile(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	defer stream.Close()

	// Заголовки для браузера
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", meta.Filename))
	w.Header().Set("Content-Type", meta.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", meta.Size))

	io.Copy(w, stream)
}

// GET /media/{id} — Получить инфо о файле (JSON)
func (h *Handler) GetFileMeta(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	meta, err := h.svc.GetMeta(r.Context(), id)
	if err != nil {
		http.Error(w, "File not found", 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}

// DELETE /media/delete/{id}
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.svc.DeleteFile(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"deleted"}`))
}

// ListFiles - получение списка файлов пользователя
func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter is required", http.StatusBadRequest)
		return
	}

	// Параметры пагинации
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Получаем список файлов
	fileList, err := h.svc.ListUserFiles(r.Context(), userID, limit, offset)
	if err != nil {
		log.Printf("Error listing files: %v", err)
		// Возвращаем пустой список вместо ошибки
		response := map[string]interface{}{
			"files":  []interface{}{},
			"total":  0,
			"limit":  limit,
			"offset": offset,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Преобразуем в JSON
	var files []map[string]interface{}
	for _, file := range fileList.Files {
		files = append(files, map[string]interface{}{
			"id":           file.ID,
			"filename":     file.Filename,
			"content_type": file.ContentType,
			"size":         file.Size,
			"uploaded_at":  file.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"user_id":      file.UserID,
			"description":  file.Description,
			"chat_id":      file.ChatID,
		})
	}

	response := map[string]interface{}{
		"files":  files,
		"total":  fileList.Total,
		"limit":  fileList.Limit,
		"offset": fileList.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
