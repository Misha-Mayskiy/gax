package websocket

import (
	"encoding/json"
	"testing"
)

func TestSignalingMessage(t *testing.T) {
	// Проверяем, что JSON маршалится правильно
	msg := SignalingMessage{
		Type: TypeJoin,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		t.Error(err)
	}

	var decoded SignalingMessage
	err = json.Unmarshal(bytes, &decoded)
	if err != nil {
		t.Error(err)
	}

	if decoded.Type != TypeJoin {
		t.Error("Type mismatch")
	}
}
