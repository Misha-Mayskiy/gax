package service

import (
	"context"
	"errors"
	"testing"

	api "main/pkg/api"
)

// Тестовый репозиторий для проверки логики
type testRepo struct {
	savedRooms map[string]*api.RoomResponse
}

func newTestRepo() *testRepo {
	return &testRepo{
		savedRooms: make(map[string]*api.RoomResponse),
	}
}

func (t *testRepo) SaveRoom(ctx context.Context, room *api.RoomResponse) error {
	t.savedRooms[room.RoomId] = room
	return nil
}

func (t *testRepo) GetRoom(ctx context.Context, roomID string) (*api.RoomResponse, error) {
	room, exists := t.savedRooms[roomID]
	if !exists {
		return nil, ErrNotFound
	}
	return room, nil
}

// Создаем простую ошибку для тестов
var ErrNotFound = errors.New("not found")

func TestCreateRoom(t *testing.T) {
	// Создаем тестовый репозиторий
	repo := newTestRepo()
	service := NewRoomService(repo)

	// Тест 1: Создание комнаты
	req := &api.CreateRoomRequest{
		HostId:   "user1",
		TrackUrl: "song.mp3",
	}

	room, err := service.CreateRoom(context.Background(), req)
	if err != nil {
		t.Fatalf("Ошибка при создании комнаты: %v", err)
	}

	// Проверяем результат
	if room.RoomId == "" {
		t.Error("ID комнаты не должен быть пустым")
	}
	if room.HostId != "user1" {
		t.Errorf("Ожидался hostId=user1, получил %s", room.HostId)
	}
	if room.TrackUrl != "song.mp3" {
		t.Errorf("Ожидался trackUrl=song.mp3, получил %s", room.TrackUrl)
	}
	if room.Status != "pause" {
		t.Errorf("Ожидался статус pause, получил %s", room.Status)
	}
	if len(room.Users) != 1 || room.Users[0] != "user1" {
		t.Errorf("Ожидался один пользователь user1, получил %v", room.Users)
	}

	// Проверяем, что комната сохранилась
	savedRoom, exists := repo.savedRooms[room.RoomId]
	if !exists {
		t.Error("Комната не сохранилась в репозитории")
	}
	if savedRoom != room {
		t.Error("Сохраненная комната должна быть той же самой")
	}
}

func TestGetState(t *testing.T) {
	repo := newTestRepo()
	service := NewRoomService(repo)

	// Сначала создаем комнату
	createReq := &api.CreateRoomRequest{
		HostId:   "user1",
		TrackUrl: "song.mp3",
	}
	createdRoom, _ := service.CreateRoom(context.Background(), createReq)

	// Тест: Получение состояния существующей комнаты
	getReq := &api.GetStateRequest{
		RoomId: createdRoom.RoomId,
	}

	room, err := service.GetState(context.Background(), getReq)
	if err != nil {
		t.Fatalf("Ошибка при получении состояния: %v", err)
	}

	if room.RoomId != createdRoom.RoomId {
		t.Errorf("Ожидалась комната с ID %s, получил %s", createdRoom.RoomId, room.RoomId)
	}

	// Тест: Попытка получить несуществующую комнату
	badReq := &api.GetStateRequest{
		RoomId: "несуществующий",
	}

	_, err = service.GetState(context.Background(), badReq)
	if err == nil {
		t.Error("Ожидалась ошибка для несуществующей комнаты")
	}
}

func TestJoinRoom(t *testing.T) {
	repo := newTestRepo()
	service := NewRoomService(repo)

	// Создаем комнату
	createReq := &api.CreateRoomRequest{
		HostId:   "host",
		TrackUrl: "song.mp3",
	}
	createdRoom, _ := service.CreateRoom(context.Background(), createReq)

	// Тест 1: Новый пользователь присоединяется
	joinReq1 := &api.JoinRoomRequest{
		RoomId: createdRoom.RoomId,
		UserId: "user1",
	}

	room1, err := service.JoinRoom(context.Background(), joinReq1)
	if err != nil {
		t.Fatalf("Ошибка при присоединении: %v", err)
	}

	// Проверяем, что пользователь добавился
	found := false
	for _, user := range room1.Users {
		if user == "user1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Пользователь user1 не добавился в комнату")
	}
	if len(room1.Users) != 2 {
		t.Errorf("Ожидалось 2 пользователя, получил %d", len(room1.Users))
	}

	// Тест 2: Существующий пользователь снова присоединяется
	room2, err := service.JoinRoom(context.Background(), joinReq1)
	if err != nil {
		t.Fatalf("Ошибка при повторном присоединении: %v", err)
	}

	// Должен остаться тот же список пользователей
	if len(room2.Users) != 2 {
		t.Errorf("Количество пользователей не должно меняться при повторном присоединении")
	}

	// Тест 3: Присоединение к несуществующей комнате
	badReq := &api.JoinRoomRequest{
		RoomId: "несуществующий",
		UserId: "user2",
	}

	_, err = service.JoinRoom(context.Background(), badReq)
	if err == nil {
		t.Error("Ожидалась ошибка для несуществующей комнаты")
	}
}

func TestSetPlayback(t *testing.T) {
	repo := newTestRepo()
	service := NewRoomService(repo)

	// Создаем комнату
	createReq := &api.CreateRoomRequest{
		HostId:   "host",
		TrackUrl: "song.mp3",
	}
	createdRoom, _ := service.CreateRoom(context.Background(), createReq)

	// Тест: Устанавливаем воспроизведение
	setReq := &api.SetPlaybackRequest{
		RoomId:    createdRoom.RoomId,
		Action:    "play",
		Timestamp: 100,
	}

	room, err := service.SetPlayback(context.Background(), setReq)
	if err != nil {
		t.Fatalf("Ошибка при установке воспроизведения: %v", err)
	}

	if room.Status != "play" {
		t.Errorf("Ожидался статус play, получил %s", room.Status)
	}
	if room.Timestamp != 100 {
		t.Errorf("Ожидался timestamp 100, получил %d", room.Timestamp)
	}

	// Тест: Устанавливаем паузу
	setReq2 := &api.SetPlaybackRequest{
		RoomId:    createdRoom.RoomId,
		Action:    "pause",
		Timestamp: 150,
	}

	room2, err := service.SetPlayback(context.Background(), setReq2)
	if err != nil {
		t.Fatalf("Ошибка при установке паузы: %v", err)
	}

	if room2.Status != "pause" {
		t.Errorf("Ожидался статус pause, получил %s", room2.Status)
	}
	if room2.Timestamp != 150 {
		t.Errorf("Ожидался timestamp 150, получил %d", room2.Timestamp)
	}

	// Тест: Попытка изменить несуществующую комнату
	badReq := &api.SetPlaybackRequest{
		RoomId:    "несуществующий",
		Action:    "play",
		Timestamp: 100,
	}

	_, err = service.SetPlayback(context.Background(), badReq)
	if err == nil {
		t.Error("Ожидалась ошибка для несуществующей комнаты")
	}
}

func TestGenerateRoomID(t *testing.T) {
	// Генерируем несколько ID
	ids := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		id := generateRoomID()

		// Проверяем длину
		if len(id) != 6 {
			t.Errorf("ID должен быть длиной 6 символов, получил %s длиной %d", id, len(id))
		}

		// Проверяем символы
		for _, char := range id {
			if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				t.Errorf("ID должен содержать только буквы A-Z и цифры 0-9, получил %s", id)
			}
		}

		// Проверяем уникальность
		if ids[id] {
			t.Errorf("Повторяющийся ID: %s", id)
		}
		ids[id] = true
	}
}

// Простой тест на проверку работы интерфейса
func TestServiceInterface(t *testing.T) {
	var _ RoomService = &roomService{}
}
