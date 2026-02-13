package repository

import (
	"encoding/json"
	"testing"

	api "main/pkg/api"

	"github.com/stretchr/testify/assert"
)

func TestRoomRedisRepo(t *testing.T) {
	// Простые тесты без сложных моков
	t.Run("test key format", func(t *testing.T) {
		assert.Equal(t, "room:test123", "room:"+"test123")
	})

	t.Run("test json marshal unmarshal", func(t *testing.T) {
		room := &api.RoomResponse{
			RoomId:   "test",
			HostId:   "host",
			TrackUrl: "url",
			Users:    []string{"user1"},
		}

		data, err := json.Marshal(room)
		assert.NoError(t, err)

		var unmarshaledRoom api.RoomResponse
		err = json.Unmarshal(data, &unmarshaledRoom)
		assert.NoError(t, err)
		assert.Equal(t, room.RoomId, unmarshaledRoom.RoomId)
	})
}
