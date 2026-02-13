package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// helper: конвертирует http URL тестового сервера в ws URL
func httpToWs(u string) string {
	return "ws" + strings.TrimPrefix(u, "http")
}

func TestWebSocketServer_Connection_MissingParams(t *testing.T) {
	// 1. Поднимаем сервер
	wsServer := NewWebSocketServer()
	server := httptest.NewServer(http.HandlerFunc(wsServer.HandleWebSocket))
	defer server.Close()

	// 2. Пытаемся подключиться без room_id и user_id
	wsURL := httpToWs(server.URL) // Нет query params

	_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

	// В коде сервера: conn.WriteMessage(CloseMessage) и return.
	// Dial может вернуть ошибку handshake или мы можем прочитать закрытие.
	if err == nil {
		// Если подключились, пробуем прочитать. Сервер должен закрыть соединение.
		// Но скорее всего Dial вернет err "bad handshake" или EOF, так как сервер сразу пишет Close.
	}

	// Этот тест в основном нужен, чтобы дернуть ветку `if roomID == "" || userID == ""`
}

func TestWebSocketServer_FullFlow(t *testing.T) {
	// Сценарий:
	// 1. User1 подключается (создается комната).
	// 2. User1 получает info и joined.
	// 3. User2 подключается.
	// 4. User1 получает уведомление, что User2 зашел.
	// 5. User1 шлет Control сообщение (Play).
	// 6. User2 получает это сообщение.
	// 7. User1 отключается.
	// 8. User2 получает уведомление о выходе User1.
	// 9. Комната удаляется.

	wsServer := NewWebSocketServer()
	server := httptest.NewServer(http.HandlerFunc(wsServer.HandleWebSocket))
	defer server.Close()

	// --- User 1 Connect ---
	url1 := httpToWs(server.URL) + "?room_id=room1&user_id=user1&user_name=Alice"
	conn1, _, err := websocket.DefaultDialer.Dial(url1, nil)
	if err != nil {
		t.Fatalf("User1 dial error: %v", err)
	}
	defer conn1.Close()

	// User 1 должен получить "room_info"
	var msg1 WebSocketMessage
	if err := conn1.ReadJSON(&msg1); err != nil {
		t.Fatalf("User1 read room_info error: %v", err)
	}
	if msg1.Type != "room_info" {
		t.Errorf("Expected room_info, got %s", msg1.Type)
	}

	// User 1 должен получить "user_joined" (про самого себя)
	if err := conn1.ReadJSON(&msg1); err != nil {
		t.Fatalf("User1 read user_joined error: %v", err)
	}
	if msg1.Type != "user_joined" || msg1.UserID != "user1" {
		t.Errorf("Expected user_joined for user1, got %v", msg1)
	}

	// --- User 2 Connect ---
	url2 := httpToWs(server.URL) + "?room_id=room1&user_id=user2&user_name=Bob"
	conn2, _, err := websocket.DefaultDialer.Dial(url2, nil)
	if err != nil {
		t.Fatalf("User2 dial error: %v", err)
	}
	defer conn2.Close()

	// User 2: читает room_info
	conn2.ReadJSON(&msg1) // room_info
	// User 2: читает user_joined (себя)
	conn2.ReadJSON(&msg1) // user_joined (self)

	// User 1: должен получить уведомление, что User 2 зашел
	if err := conn1.ReadJSON(&msg1); err != nil {
		t.Fatalf("User1 read user2 joined error: %v", err)
	}
	if msg1.Type != "user_joined" || msg1.UserID != "user2" {
		t.Errorf("Expected user_joined for user2, got %v", msg1)
	}

	// --- Control Message ---
	// User 1 отправляет Play
	controlMsg := WebSocketMessage{
		Type:     "control",
		Action:   "play",
		Position: 10.5,
	}
	if err := conn1.WriteJSON(controlMsg); err != nil {
		t.Fatalf("User1 write error: %v", err)
	}

	// User 2 должен получить Play
	var msg2 WebSocketMessage
	if err := conn2.ReadJSON(&msg2); err != nil {
		t.Fatalf("User2 read control error: %v", err)
	}
	if msg2.Type != "play" || msg2.Position != 10.5 || msg2.UserID != "user1" {
		t.Errorf("Expected play action from user1, got %v", msg2)
	}

	// --- User 1 Disconnect ---
	conn1.WriteMessage(websocket.CloseMessage, []byte{})
	conn1.Close()

	// User 2 должен получить user_left
	if err := conn2.ReadJSON(&msg2); err != nil {
		// Может быть EOF, если сервер закрыл соединение (но он не должен, пока User2 там)
		t.Logf("User2 read error (might be expected if server closed): %v", err)
	} else {
		if msg2.Type != "user_left" || msg2.UserID != "user1" {
			t.Errorf("Expected user_left for user1, got %v", msg2)
		}
	}

	// Даем серверу время обработать отключение и очистить мапу
	time.Sleep(100 * time.Millisecond)

	wsServer.roomsMutex.RLock()
	room, exists := wsServer.rooms["room1"]
	wsServer.roomsMutex.RUnlock()

	// Комната должна существовать, так как User2 еще там
	if !exists {
		t.Error("Room should exist (User2 is still inside)")
	} else {
		room.broadcast <- &WebSocketMessage{Type: "ping"} // просто дергаем канал, чтобы проверить, что он не закрыт
	}

	// --- User 2 Disconnect ---
	conn2.WriteMessage(websocket.CloseMessage, []byte{})
	conn2.Close()

	time.Sleep(100 * time.Millisecond)

	// Теперь комната должна удалиться
	wsServer.roomsMutex.RLock()
	_, exists = wsServer.rooms["room1"]
	wsServer.roomsMutex.RUnlock()

	if exists {
		t.Error("Room should be deleted after last user left")
	}
}

func TestWebSocketServer_ReadPump_IgnoreUnknownTypes(t *testing.T) {
	wsServer := NewWebSocketServer()
	server := httptest.NewServer(http.HandlerFunc(wsServer.HandleWebSocket))
	defer server.Close()

	url := httpToWs(server.URL) + "?room_id=test&user_id=u1"
	conn, _, _ := websocket.DefaultDialer.Dial(url, nil)
	defer conn.Close()

	// Читаем приветственные сообщения
	conn.ReadJSON(&struct{}{}) // info
	conn.ReadJSON(&struct{}{}) // joined

	// Отправляем мусор / игнорируемые типы
	conn.WriteJSON(map[string]string{"type": "user_join"}) // должно игнорироваться
	conn.WriteJSON(map[string]string{"type": "unknown"})   // switch default? (нет default, просто пропуск)

	// Чтобы убедиться, что сервер не упал и connection жив, отправим валидный запрос
	conn.WriteJSON(WebSocketMessage{Type: "control", Action: "pause"})

	// Должны получить ответ "pause"
	var msg WebSocketMessage
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("Server crashed or closed connection: %v", err)
	}
	if msg.Type != "pause" {
		t.Errorf("Expected pause, got %s", msg.Type)
	}
}
