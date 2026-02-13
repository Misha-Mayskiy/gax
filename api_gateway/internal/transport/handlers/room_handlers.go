package handlers

import (
	pb "api_gateway/pkg/api/room"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // В production заменить на проверку origin
	},
}

type RoomWebSocketClient struct {
	conn   *websocket.Conn
	userID string
	roomID string
	send   chan []byte
}

type RoomWebSocketHub struct {
	rooms      map[string]map[*RoomWebSocketClient]bool // roomID -> clients
	register   chan *RoomWebSocketClient
	unregister chan *RoomWebSocketClient
	broadcast  chan WebSocketMessage
	mutex      sync.RWMutex
}

type WebSocketMessage struct {
	Type    string          `json:"type"`
	RoomID  string          `json:"room_id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func (h *RoomWebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			if _, exists := h.rooms[client.roomID]; !exists {
				h.rooms[client.roomID] = make(map[*RoomWebSocketClient]bool)
			}
			h.rooms[client.roomID][client] = true
			h.mutex.Unlock()

			// Уведомляем других участников о подключении
			joinMsg, _ := json.Marshal(map[string]string{"user_id": client.userID})
			h.broadcast <- WebSocketMessage{
				Type:    MessageTypeJoin,
				RoomID:  client.roomID,
				Payload: joinMsg,
			}

		case client := <-h.unregister:
			h.mutex.Lock()
			if room, exists := h.rooms[client.roomID]; exists {
				if _, clientExists := room[client]; clientExists {
					delete(room, client)
					close(client.send)
					if len(room) == 0 {
						delete(h.rooms, client.roomID)
					}
				}
			}
			h.mutex.Unlock()

			// Уведомляем об отключении
			leaveMsg, _ := json.Marshal(map[string]string{"user_id": client.userID})
			h.broadcast <- WebSocketMessage{
				Type:    MessageTypeLeave,
				RoomID:  client.roomID,
				Payload: leaveMsg,
			}

		case message := <-h.broadcast:
			h.mutex.RLock()
			if room, exists := h.rooms[message.RoomID]; exists {
				msgBytes, _ := json.Marshal(message)
				for client := range room {
					select {
					case client.send <- msgBytes:
					default:
						close(client.send)
						delete(room, client)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// Типы сообщений
const (
	MessageTypeJoin        = "join"
	MessageTypeLeave       = "leave"
	MessageTypePlayback    = "playback"
	MessageTypeStateUpdate = "state_update"
	MessageTypeChat        = "chat"
	MessageTypeError       = "error"
)

func NewRoomWebSocketHub() *RoomWebSocketHub {
	return &RoomWebSocketHub{
		rooms:      make(map[string]map[*RoomWebSocketClient]bool),
		register:   make(chan *RoomWebSocketClient),
		unregister: make(chan *RoomWebSocketClient),
		broadcast:  make(chan WebSocketMessage),
	}
}

// Существующие REST handlers остаются без изменений
func (h *HandlerFacade) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req pb.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := h.roomService.CreateRoom(h.ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlerFacade) JoinRoom(w http.ResponseWriter, r *http.Request) {
	var req pb.JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := h.roomService.JoinRoom(h.ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlerFacade) SetPlayback(w http.ResponseWriter, r *http.Request) {
	var req pb.SetPlaybackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := h.roomService.SetPlayback(h.ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlerFacade) GetState(w http.ResponseWriter, r *http.Request) {
	var req pb.GetStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := h.roomService.GetState(h.ctx, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// WebSocket handler с упрощенной логикой
func (h *HandlerFacade) RoomWebSocket(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры
	userID := r.URL.Query().Get("user_id")
	roomID := r.URL.Query().Get("room_id")
	token := r.URL.Query().Get("token")

	// Упрощенная проверка
	if userID == "" || roomID == "" {
		http.Error(w, "Missing user_id or room_id", http.StatusBadRequest)
		return
	}

	// Проверка токена
	if !h.authService.ValidateToken(h.ctx, token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Обновление до WebSocket соединения
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &RoomWebSocketClient{
		conn:   conn,
		userID: userID,
		roomID: roomID,
		send:   make(chan []byte, 256),
	}

	// Регистрируем клиента
	h.roomWebSocketHub.register <- client

	// Запускаем горутины для чтения/записи
	go h.writePump(client)
	go h.readPump(client)
}

func (h *HandlerFacade) writePump(client *RoomWebSocketClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
		h.roomWebSocketHub.unregister <- client
	}()

	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			// Ping для поддержания соединения
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *HandlerFacade) readPump(client *RoomWebSocketClient) {
	defer func() {
		h.roomWebSocketHub.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadLimit(512 * 1024) // 512KB
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		h.handleWebSocketMessage(client, message)
	}
}

func (h *HandlerFacade) handleWebSocketMessage(client *RoomWebSocketClient, message []byte) {
	var msg WebSocketMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid WebSocket message: %v", err)
		return
	}

	// Устанавливаем RoomID если не указан
	if msg.RoomID == "" {
		msg.RoomID = client.roomID
	}

	switch msg.Type {
	case MessageTypePlayback:
		// Обработка управления воспроизведением
		h.handlePlaybackControl(client.roomID, msg.Payload)
		// Рассылаем обновление всем участникам
		h.roomWebSocketHub.broadcast <- msg

	case MessageTypeChat:
		// Просто пересылаем сообщение чата всем участникам
		h.roomWebSocketHub.broadcast <- msg

	case "sync_request":
		// Запрос текущего состояния комнаты
		state, err := h.roomService.GetRoomState(client.roomID)
		if err != nil {
			log.Printf("Failed to get room state: %v", err)
			// Отправляем ошибку клиенту
			errorMsg := WebSocketMessage{
				Type:    MessageTypeError,
				RoomID:  client.roomID,
				Payload: json.RawMessage(`{"error":"` + err.Error() + `"}`),
			}
			errorBytes, _ := json.Marshal(errorMsg)
			client.send <- errorBytes
			return
		}
		stateJSON, _ := json.Marshal(state)
		response := WebSocketMessage{
			Type:    MessageTypeStateUpdate,
			RoomID:  client.roomID,
			Payload: stateJSON,
		}
		responseBytes, _ := json.Marshal(response)
		client.send <- responseBytes

	default:
		// Для других типов сообщений просто рассылаем всем
		h.roomWebSocketHub.broadcast <- msg
	}
}

// Вспомогательный метод для обработки playback control
func (h *HandlerFacade) handlePlaybackControl(roomID string, payload json.RawMessage) {
	var control struct {
		Action    string  `json:"action"` // play, pause, seek
		Position  float64 `json:"position,omitempty"`
		Timestamp int64   `json:"timestamp"`
		UserID    string  `json:"user_id"`
	}

	if err := json.Unmarshal(payload, &control); err != nil {
		log.Printf("Failed to unmarshal playback control: %v", err)
		return
	}

	// Обновляем состояние через roomService
	err := h.roomService.UpdatePlaybackState(roomID, control.Action, control.Position, 0)
	if err != nil {
		log.Printf("Failed to update playback state: %v", err)
	}
}
