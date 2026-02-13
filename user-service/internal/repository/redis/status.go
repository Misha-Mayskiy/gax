package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func New(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // если без пароля
		DB:       0,
	})
	return &Client{rdb: rdb}
}

// Добавить пользователя в онлайн (с TTL)
func (c *Client) SetOnline(ctx context.Context, uuid string, ttl time.Duration) error {
	return c.rdb.Set(ctx, "online:"+uuid, "1", ttl).Err()
}

// Удалить пользователя из онлайн
func (c *Client) SetOffline(ctx context.Context, uuid string) error {
	return c.rdb.Del(ctx, "online:"+uuid).Err()
}

// Проверить онлайн‑статус
func (c *Client) IsOnline(ctx context.Context, uuid string) (bool, error) {
	exists, err := c.rdb.Exists(ctx, "online:"+uuid).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// Получить список всех онлайн‑пользователей
func (c *Client) GetOnlineUsers(ctx context.Context) ([]string, error) {
	keys, err := c.rdb.Keys(ctx, "online:*").Result()
	if err != nil {
		return nil, err
	}

	uuids := make([]string, 0, len(keys))
	for _, k := range keys {
		// ключи вида "online:<uuid>"
		uuids = append(uuids, k[len("online:"):])
	}
	return uuids, nil
}
