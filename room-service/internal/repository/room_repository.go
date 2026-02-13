package repository

import (
	"context"
	api "main/pkg/api"
)

type RoomRepository interface {
	SaveRoom(ctx context.Context, room *api.RoomResponse) error
	GetRoom(ctx context.Context, roomID string) (*api.RoomResponse, error)
}
