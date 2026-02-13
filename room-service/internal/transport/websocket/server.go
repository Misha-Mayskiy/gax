package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Локальные структуры для WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	RoomID    string      `json:"room_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	UserName  string      `json:"user_name,omitempty"`
	Action    string      `json:"action,omitempty"`
	Position  float64     `json:"position,omitempty"`
	Users     []User      `json:"users,omitempty"`
	User      *User       `json:"user,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsHost   bool   `json:"is_host"`
	JoinedAt int64  `json:"joined_at"`
}

type WebSocketServer struct {
	upgrader   websocket.Upgrader
	rooms      map[string]*Room
	roomsMutex sync.RWMutex
}

type Room struct {
	ID        string
	HostID    string
	Clients   map[*Client]bool
	broadcast chan *WebSocketMessage
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
	room *Room
	user *User
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		rooms: make(map[string]*Room),
	}
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Парсим параметры
	roomID := r.URL.Query().Get("room_id")
	userID := r.URL.Query().Get("user_id")
	userName := r.URL.Query().Get("user_name")

	if roomID == "" || userID == "" {
		conn.WriteMessage(websocket.CloseMessage, []byte("Missing room_id or user_id"))
		return
	}

	// Создаем или получаем комнату
	s.roomsMutex.Lock()
	room, exists := s.rooms[roomID]
	if !exists {
		room = &Room{
			ID:        roomID,
			Clients:   make(map[*Client]bool),
			broadcast: make(chan *WebSocketMessage, 100),
		}
		s.rooms[roomID] = room
		go room.runBroadcaster()
	}
	s.roomsMutex.Unlock()

	// Создаем клиента
	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
		room: room,
		user: &User{
			ID:       userID,
			Name:     userName,
			IsHost:   false, // TODO: проверять хост
			JoinedAt: time.Now().Unix(),
		},
	}

	room.Clients[client] = true

	// Отправляем информацию о комнате новому пользователю
	s.sendRoomInfo(client)

	// Уведомляем о присоединении
	joinMessage := &WebSocketMessage{
		Type:      "user_joined",
		RoomID:    roomID,
		UserID:    userID,
		UserName:  userName,
		User:      client.user,
		Timestamp: time.Now().Unix(),
	}
	room.broadcast <- joinMessage

	// Запускаем обработчики
	go client.writePump()
	client.readPump()

	// Удаляем клиента при отключении
	delete(room.Clients, client)
	close(client.send)

	// Уведомляем об отключении
	leaveMessage := &WebSocketMessage{
		Type:      "user_left",
		RoomID:    roomID,
		UserID:    userID,
		UserName:  userName,
		User:      client.user,
		Timestamp: time.Now().Unix(),
	}
	room.broadcast <- leaveMessage

	// Если комната пуста - удаляем
	if len(room.Clients) == 0 {
		s.roomsMutex.Lock()
		delete(s.rooms, roomID)
		s.roomsMutex.Unlock()
	}
}

func (s *WebSocketServer) sendRoomInfo(client *Client) {
	// Собираем список пользователей в комнате
	users := make([]User, 0, len(client.room.Clients))
	for c := range client.room.Clients {
		users = append(users, *c.user)
	}

	// Отправляем информацию о комнате
	roomInfo := &WebSocketMessage{
		Type:      "room_info",
		RoomID:    client.room.ID,
		Users:     users,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(roomInfo)
	if err != nil {
		return
	}

	client.send <- data
}

func (r *Room) runBroadcaster() {
	for message := range r.broadcast {
		data, err := json.Marshal(message)
		if err != nil {
			continue
		}

		for client := range r.Clients {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(r.Clients, client)
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// Обработка сообщений от клиента
		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Обработка сообщений
		switch msg.Type {
		case "control":
			// Пересылаем управление всем клиентам
			controlMsg := &WebSocketMessage{
				Type:      msg.Action, // "play", "pause", "seek"
				RoomID:    c.room.ID,
				UserID:    c.user.ID,
				UserName:  c.user.Name,
				Position:  msg.Position,
				Timestamp: time.Now().Unix(),
			}
			c.room.broadcast <- controlMsg

		case "user_join":
			// Уже обработано при подключении

		case "user_leave":
			// Уже обработано при отключении
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
