package websocket

import "encoding/json"

const (
	TypeJoin      = "join"
	TypeOffer     = "offer"
	TypeAnswer    = "answer"
	TypeCandidate = "candidate"
	TypeUserLeft  = "user_left"
)

// Обертка для сообщений сигнализации WebRTC
type SignalingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Данные для входа (UserID из заголовка)
type JoinPayload struct {
	RoomID string `json:"room_id"`
}
