package sfu

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// ВАЖНО:
// Я предполагаю, что структура Peer выглядит примерно так (на основе использования в room.go):
// Если у вас в peer.go поля называются иначе, поправьте создание Peer в тестах.
/*
type Peer struct {
	ID                 string
	Conn               *websocket.Conn
	PC                 *webrtc.PeerConnection
	Lock               sync.Mutex
	StreamIDs          []string
	NegotiationPending bool
}
func (p *Peer) Close() { ... }
func (p *Peer) SendJSON(v interface{}) error { ... }
*/

// helper для создания реального WS соединения
func createWSConnection(t *testing.T) (*websocket.Conn, *websocket.Conn, func()) {
	// Создаем тестовый сервер, который просто апгрейдит соединение
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

	// Канал, чтобы передать серверное соединение обратно в тест
	serverConnChan := make(chan *websocket.Conn, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		serverConnChan <- conn
	}))

	// Клиент подключается к серверу
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}

	// Получаем серверную сторону соединения
	serverConn := <-serverConnChan

	cleanup := func() {
		clientConn.Close()
		serverConn.Close()
		srv.Close()
	}

	return clientConn, serverConn, cleanup
}

func TestRoom_JoinAndSignal(t *testing.T) {
	// 1. Подготовка сети (реальный WebSocket через loopback)
	clientConn, serverConn, cleanup := createWSConnection(t)
	defer cleanup()

	// 2. Создаем комнату
	room := NewRoom("test-room-1")

	// 3. Создаем Peer (используем серверную часть соединения)
	// ВАЖНО: Тут мы вручную создаем структуру Peer. Убедитесь, что поля совпадают с вашим peer.go
	peer := &Peer{
		ID:        "user-1",
		Conn:      serverConn,
		StreamIDs: make([]string, 0),
		// PC инициализируется внутри Room.Join
	}

	// 4. Тестируем Join
	// Этот метод создаст PeerConnection, добавит Peer в комнату и вызовет Signal
	err := room.Join(peer)
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}

	// Проверяем, что Peer добавлен
	room.Lock.RLock()
	if _, ok := room.Peers["user-1"]; !ok {
		t.Error("Peer was not added to room map")
	}
	room.Lock.RUnlock()

	// Проверяем, что PC создан
	if peer.PC == nil {
		t.Error("PeerConnection was not initialized")
	}

	// 5. Тестируем Signal (проверяем, что клиент получил Offer)
	// Signal вызывается внутри Join, поэтому мы ожидаем сообщение в вебсокете

	// Читаем сообщение со стороны клиента
	var msg map[string]interface{}
	if err := clientConn.ReadJSON(&msg); err != nil {
		t.Fatalf("Client read error: %v", err)
	}

	if msg["type"] != "offer" {
		t.Errorf("Expected message type 'offer', got %v", msg["type"])
	}
	// Проверяем наличие SDP payload
	if payload, ok := msg["payload"].(map[string]interface{}); ok {
		if _, ok := payload["sdp"]; !ok {
			t.Error("Payload missing SDP")
		}
	} else {
		t.Error("Payload format invalid")
	}
}

func TestRoom_RemovePeer_And_Notify(t *testing.T) {
	// Тест сценария: User 1 и User 2 в комнате. User 1 уходит -> User 2 получает уведомление.

	// --- Setup User 1 ---
	_, sConn1, cleanup1 := createWSConnection(t)
	defer cleanup1()
	peer1 := &Peer{ID: "user-1", Conn: sConn1, StreamIDs: []string{"stream-A"}}

	// --- Setup User 2 ---
	cConn2, sConn2, cleanup2 := createWSConnection(t)
	defer cleanup2()
	peer2 := &Peer{ID: "user-2", Conn: sConn2}

	room := NewRoom("test-room-multi")

	// Добавляем обоих вручную (без Join, чтобы не поднимать WebRTC стек для этого теста,
	// так как нам нужна проверка логики уведомлений, а Join слишком тяжелый)
	// Но нам нужно, чтобы методы Peer (Close, SendJSON) работали.

	// ВАЖНО: Если Join обязателен для инициализации чего-то критичного, используйте его.
	// Здесь я использую AddPeer напрямую для простоты теста удаления.
	room.AddPeer(peer1)
	room.AddPeer(peer2)

	// Эмулируем, что у комнаты есть трек от user-1 (для теста очистки треков)
	// Mocking tracks is hard without pion internals, so we verify logic flow mainly.

	// --- Action: Remove User 1 ---
	room.RemovePeer("user-1")

	// --- Verification ---

	// 1. User 1 должен исчезнуть из мапы
	room.Lock.RLock()
	if _, ok := room.Peers["user-1"]; ok {
		t.Error("User 1 should be removed")
	}
	if _, ok := room.Peers["user-2"]; !ok {
		t.Error("User 2 should stay")
	}
	room.Lock.RUnlock()

	// 2. User 2 должен получить уведомление "user_left"
	var msg map[string]interface{}

	// Ставим таймаут на чтение, чтобы тест не завис
	cConn2.SetReadDeadline(time.Now().Add(1 * time.Second))
	err := cConn2.ReadJSON(&msg)
	if err != nil {
		t.Fatalf("User 2 did not receive notification: %v", err)
	}

	if msg["type"] != "user_left" {
		t.Errorf("Expected 'user_left', got %v", msg["type"])
	}

	payload := msg["payload"].(map[string]interface{})
	if payload["user_id"] != "user-1" {
		t.Errorf("Expected user_id 'user-1', got %v", payload["user_id"])
	}
}

func TestRoom_RemovePeer_NotFound(t *testing.T) {
	// Простой тест на удаление несуществующего пира (чтобы не было паники)
	room := NewRoom("empty")
	room.RemovePeer("ghost")
	// Если не упало - тест пройден
}
