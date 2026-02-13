package domain

import "time"

type SearchType string

const (
	SearchTypeUser SearchType = "user"
	SearchTypeChat SearchType = "chat"
	SearchTypeFile SearchType = "file"
)

// Объекты для индексации (агрегированный вид)
type UserIndex struct {
	ID        string   `json:"id"` // новый идентификатор
	UUID      string   `json:"uuid"`
	UserName  string   `json:"username"`
	Email     string   `json:"email"`
	AboutMe   string   `json:"about_me"`
	Friends   []string `json:"friends"`
	UpdatedAt int64    `json:"updated_at"`
}

type ChatIndex struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	MemberIDs []string `json:"member_ids"`
	Kind      string   `json:"kind"`
	UpdatedAt int64    `json:"updated_at"`
}

type FileIndex struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Mime      string `json:"mime"`
	Type      string `json:"type"`
	SizeBytes int64  `json:"size_bytes"`
	AuthorID  string `json:"author_id"`
	ChatID    string `json:"chat_id"`
	UpdatedAt int64  `json:"updated_at"`
}

// Результат поиска (универсальный формат)
type SearchResult struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"` // user | chat | file
	Title    string   `json:"title,omitempty"`
	Snippet  string   `json:"snippet,omitempty"` // highlight
	ExtraIDs []string `json:"extra_ids,omitempty"`
}

// Параметры запроса
type SearchQuery struct {
	Q         string
	Type      string // "", "user", "chat", "file"
	Limit     int
	Offset    int
	Highlight bool
}

type Chat struct {
	ID        string   `bson:"id"`
	Kind      string   `bson:"kind"`
	MemberIDs []string `bson:"member_ids"`
	Title     string   `bson:"title"`
	CreatedBy string   `bson:"created_by"`
	CreatedAt int64    `bson:"created_at"`
}

type Message struct {
	ID        string `json:"id"`
	ChatID    string `json:"chat_id"`
	AuthorID  string `json:"author_id"`
	Text      string `json:"text"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	Deleted   bool   `json:"deleted"`
}
type Media struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	Bucket      string    `json:"bucket"`
	ObjectName  string    `json:"object_name"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}

// MediaUploadRequest запрос на загрузку файла
type MediaUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	// Для загрузки через multipart/form-data
}

// MediaUploadResponse ответ при загрузке файла
type MediaUploadResponse struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	URL         string `json:"url,omitempty"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Message     string `json:"message"`
}

// MediaListResponse ответ со списком файлов
type MediaListResponse struct {
	Total   int     `json:"total"`
	Limit   int     `json:"limit"`
	Offset  int     `json:"offset"`
	Items   []Media `json:"items"`
	Message string  `json:"message"`
}

// MediaDownloadRequest запрос на скачивание файла
type MediaDownloadRequest struct {
	ID string `json:"id"`
}

// MediaDeleteRequest запрос на удаление файла
type MediaDeleteRequest struct {
	ID string `json:"id"`
}

// MediaListRequest запрос на получение списка файлов
type MediaListRequest struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// MediaMetaRequest запрос на получение метаданных файла
type MediaMetaRequest struct {
	ID string `json:"id"`
}
type SearchError struct {
	Message string
}

func (e *SearchError) Error() string {
	return e.Message
}
