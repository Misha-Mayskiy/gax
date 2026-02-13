package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(addr string, password string) *RedisRepo {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &RedisRepo{client: rdb}
}

// Помечает юзера как "В звонке"
func (r *RedisRepo) SetUserInCall(ctx context.Context, userID string, roomID string) error {
	return r.client.Set(ctx, "call:status:"+userID, roomID, 24*time.Hour).Err()
}

// Удаляет статус юзера "В звонке"
func (r *RedisRepo) RemoveUserFromCall(ctx context.Context, userID string) error {
	return r.client.Del(ctx, "call:status:"+userID).Err()
}

// Проверяет, находится ли юзер "В звонке"
func (r *RedisRepo) IsUserBusy(ctx context.Context, userID string) (bool, string, error) {
	val, err := r.client.Get(ctx, "call:status:"+userID).Result()
	if err == redis.Nil {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, val, nil
}
