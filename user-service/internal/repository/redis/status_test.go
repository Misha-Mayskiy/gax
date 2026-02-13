package redis

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Поднимаем mock-redis
	s := miniredis.RunT(t)

	// Тестируем конструктор
	client := New(s.Addr())
	assert.NotNil(t, client)
	assert.NotNil(t, client.rdb)
}

func TestClient_SetOnline(t *testing.T) {
	s := miniredis.RunT(t)
	client := New(s.Addr())
	ctx := context.Background()

	uuid := "user-123"
	ttl := 10 * time.Minute

	// Тест установки статуса
	err := client.SetOnline(ctx, uuid, ttl)
	assert.NoError(t, err)

	// Проверяем прямо в miniredis, что ключ создался
	assert.True(t, s.Exists("online:"+uuid))

	// Проверяем значение
	val, _ := s.Get("online:" + uuid)
	assert.Equal(t, "1", val)

	// Проверяем TTL
	assert.True(t, s.TTL("online:"+uuid) > 0)
}

func TestClient_SetOffline(t *testing.T) {
	s := miniredis.RunT(t)
	client := New(s.Addr())
	ctx := context.Background()

	uuid := "user-delete"

	// Предварительно создаем ключ
	_ = s.Set("online:"+uuid, "1")

	// Удаляем
	err := client.SetOffline(ctx, uuid)
	assert.NoError(t, err)

	// Проверяем, что ключа больше нет
	assert.False(t, s.Exists("online:"+uuid))
}

func TestClient_IsOnline(t *testing.T) {
	s := miniredis.RunT(t)
	client := New(s.Addr())
	ctx := context.Background()

	// Кейс 1: Пользователь онлайн
	uuidOnline := "online-dude"
	_ = s.Set("online:"+uuidOnline, "1")

	isOnline, err := client.IsOnline(ctx, uuidOnline)
	assert.NoError(t, err)
	assert.True(t, isOnline)

	// Кейс 2: Пользователь офлайн
	uuidOffline := "offline-dude"
	isOnline, err = client.IsOnline(ctx, uuidOffline)
	assert.NoError(t, err)
	assert.False(t, isOnline)
}

func TestClient_GetOnlineUsers(t *testing.T) {
	s := miniredis.RunT(t)
	client := New(s.Addr())
	ctx := context.Background()

	// Подготовка данных
	// Эти должны попасть в выборку
	_ = s.Set("online:user1", "1")
	_ = s.Set("online:user2", "1")

	// Этот НЕ должен попасть (другой префикс)
	_ = s.Set("cache:user3", "1")

	// Вызов метода
	users, err := client.GetOnlineUsers(ctx)
	assert.NoError(t, err)

	// Redis не гарантирует порядок ключей, поэтому сортируем перед проверкой
	sort.Strings(users)
	expected := []string{"user1", "user2"}
	sort.Strings(expected)

	assert.Equal(t, expected, users)
}

// Тест на ошибку (например, если контекст отменен или Redis упал)
func TestClient_GetOnlineUsers_Error(t *testing.T) {
	s := miniredis.RunT(t)
	client := New(s.Addr())

	// Закрываем сервер принудительно, чтобы вызвать ошибку соединения
	s.Close()

	_, err := client.GetOnlineUsers(context.Background())
	assert.Error(t, err)
}
