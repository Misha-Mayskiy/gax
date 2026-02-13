package domain

type ReadInfo struct {
	UserID string `bson:"user_id"`
	ReadAt int64  `bson:"read_at"`
}

type SavedInfo struct {
	UserID  string `bson:"user_id"`
	SavedAt int64  `bson:"saved_at"`
}

// --- Чаты ---

type ChatKind string

const (
	ChatKindDirect ChatKind = "direct"
	ChatKindGroup  ChatKind = "group"
)

type Chat struct {
	ID        string   `bson:"id"`
	Kind      ChatKind `bson:"kind"`
	MemberIDs []string `bson:"member_ids"`
	Title     string   `bson:"title"`
	CreatedBy string   `bson:"created_by"`
	CreatedAt int64    `bson:"created_at"`
}

// --- Медиа ---

type Media struct {
	ID        string `bson:"id"`
	Type      string `bson:"type"`
	URL       string `bson:"url"`
	Mime      string `bson:"mime"`
	SizeBytes int64  `bson:"size_bytes"`
	AuthorID  string `bson:"author_id"`
}

// --- Сообщения ---
type Message struct {
	ID        string  `bson:"id"`
	ChatID    string  `bson:"chat_id"`
	AuthorID  string  `bson:"author_id"`
	Text      string  `bson:"text"`
	Media     []Media `bson:"media,omitempty"`
	CreatedAt int64   `bson:"created_at"`
	UpdatedAt int64   `bson:"updated_at,omitempty"`
	Deleted   bool    `bson:"deleted"`
	DeletedAt int64   `bson:"deleted_at,omitempty"`

	// Встроенные поля для оптимизации
	ReadBy  []ReadInfo  `bson:"read_by,omitempty"`  // Кто прочитал
	SavedBy []SavedInfo `bson:"saved_by,omitempty"` // Кто сохранил
}

// --- Пользователь ---

type User struct {
	ID        string   `bson:"id"` // UUID
	Email     string   `bson:"email"`
	UserName  string   `bson:"username"`
	AvatarURL string   `bson:"avatar_url"`
	AboutMe   string   `bson:"about_me"`
	Friends   []string `bson:"friends"` // UUID друзей
	CreatedAt int64    `bson:"created_at"`
	UpdatedAt int64    `bson:"updated_at"`
}

// --- Отметка о прочтении ---

type ReadReceipt struct {
	MessageID string `bson:"message_id"`
	UserID    string `bson:"user_id"`
	ReadAt    int64  `bson:"read_at"`
}

// --- Избранное сообщение ---

type SavedMessage struct {
	UserID    string `bson:"user_id"`
	MessageID string `bson:"message_id"`
	SavedAt   int64  `bson:"saved_at"`
}

// --- Событие для Kafka ---

type NewMessageEvent struct {
	MessageID string `bson:"message_id"`
	ChatID    string `bson:"chat_id"`
	AuthorID  string `bson:"author_id"`
	Text      string `bson:"text"`
	Timestamp int64  `bson:"timestamp"`
}

type SearchEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
