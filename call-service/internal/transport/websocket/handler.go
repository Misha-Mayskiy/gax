package websocket

import (
	"call-service/internal/sfu"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Handler struct {
	manager  *sfu.RoomManager
	upgrader websocket.Upgrader
}

func NewHandler(manager *sfu.RoomManager) *Handler {
	return &Handler{
		manager:  manager,
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	// Доверяем Gateway
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// Fallback для тестов
		userID = r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var currentPeer *sfu.Peer
	var currentRoom *sfu.Room

	log.Printf("New connection from %s", userID)

	for {
		var msg SignalingMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read error:", err)
			break
		}

		switch msg.Type {
		case TypeJoin:
			var payload JoinPayload
			json.Unmarshal(msg.Payload, &payload)

			// Проверка Redis
			ctx := context.Background()
			busy, activeRoom, _ := h.manager.Repo.IsUserBusy(ctx, userID)
			if busy && activeRoom != payload.RoomID {
				log.Printf("User %s is busy", userID)
				return
			}

			h.manager.Repo.SetUserInCall(ctx, userID, payload.RoomID)

			currentPeer = &sfu.Peer{ID: userID, Conn: conn}
			currentRoom = h.manager.GetOrCreateRoom(payload.RoomID)

			if err := currentRoom.Join(currentPeer); err != nil {
				log.Println("Join error:", err)
				return
			}

		case TypeOffer:
			var offer webrtc.SessionDescription
			json.Unmarshal(msg.Payload, &offer)
			currentPeer.PC.SetRemoteDescription(offer)

			answer, _ := currentPeer.PC.CreateAnswer(nil)
			currentPeer.PC.SetLocalDescription(answer)

			currentPeer.SendJSON(map[string]interface{}{"type": "answer", "payload": answer})

		case TypeAnswer:
			var answer webrtc.SessionDescription
			json.Unmarshal(msg.Payload, &answer)
			currentPeer.PC.SetRemoteDescription(answer)

			// Обработка очереди Pending
			currentPeer.Lock.Lock()
			pending := currentPeer.NegotiationPending
			if pending {
				currentPeer.NegotiationPending = false
			}
			currentPeer.Lock.Unlock()

			if pending {
				go func() {
					time.Sleep(time.Millisecond * 50)
					currentRoom.Signal(currentPeer)
				}()
			}

		case TypeCandidate:
			var candidate webrtc.ICECandidateInit
			json.Unmarshal(msg.Payload, &candidate)
			currentPeer.PC.AddICECandidate(candidate)
		}
	}

	if currentRoom != nil && currentPeer != nil {
		currentRoom.RemovePeer(currentPeer.ID)
		h.manager.Repo.RemoveUserFromCall(context.Background(), currentPeer.ID)
	}
}
