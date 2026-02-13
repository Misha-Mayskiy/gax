package repository

import (
	"context"
	"encoding/json"

	api "main/pkg/api"

	"github.com/redis/go-redis/v9"
)

type RoomRedisRepo struct {
	client *redis.Client
}

func NewRoomRedisRepo(addr string) *RoomRedisRepo {
	return &RoomRedisRepo{
		client: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (r *RoomRedisRepo) SaveRoom(ctx context.Context, room *api.RoomResponse) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, "room:"+room.RoomId, data, 0).Err()
}

func (r *RoomRedisRepo) GetRoom(ctx context.Context, roomID string) (*api.RoomResponse, error) {
	val, err := r.client.Get(ctx, "room:"+roomID).Result()
	if err != nil {
		return nil, err
	}
	var room api.RoomResponse
	if err := json.Unmarshal([]byte(val), &room); err != nil {
		return nil, err
	}
	return &room, nil
}
