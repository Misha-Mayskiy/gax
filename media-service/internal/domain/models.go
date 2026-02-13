package domain

import "time"

// FileMeta - метаданные файла
type FileMeta struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	Bucket      string    `json:"bucket"`
	ObjectName  string    `json:"object_name"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      string    `json:"user_id"`           // Добавить это поле
	Description string    `json:"description"`       // Добавить это поле
	ChatID      string    `json:"chat_id,omitempty"` // Добавить это поле
}

// FileList - список файлов с пагинацией
type FileList struct {
	Files  []*FileMeta `json:"files"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}
