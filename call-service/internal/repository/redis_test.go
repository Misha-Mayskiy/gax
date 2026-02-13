package repository

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestRedisRepo(t *testing.T) {
	// Запускаем мини-редис
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	repo := NewRedisRepo(s.Addr(), "")
	ctx := context.Background()
	userID := "user-123"
	roomID := "room-1"

	// Тест 1: SetUserInCall
	err = repo.SetUserInCall(ctx, userID, roomID)
	assert.NoError(t, err)

	// Тест 2: IsUserBusy
	busy, activeRoom, err := repo.IsUserBusy(ctx, userID)
	assert.NoError(t, err)
	assert.True(t, busy)
	assert.Equal(t, roomID, activeRoom)

	// Тест 3: RemoveUserFromCall
	err = repo.RemoveUserFromCall(ctx, userID)
	assert.NoError(t, err)

	busy, _, _ = repo.IsUserBusy(ctx, userID)
	assert.False(t, busy)
}
