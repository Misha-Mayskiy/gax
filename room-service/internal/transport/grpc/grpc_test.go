package grpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	api "main/pkg/api"
)

// Простой мок сервиса
type mockService struct {
	createRoomFunc  func(ctx context.Context, req *api.CreateRoomRequest) (*api.RoomResponse, error)
	joinRoomFunc    func(ctx context.Context, req *api.JoinRoomRequest) (*api.RoomResponse, error)
	setPlaybackFunc func(ctx context.Context, req *api.SetPlaybackRequest) (*api.RoomResponse, error)
	getStateFunc    func(ctx context.Context, req *api.GetStateRequest) (*api.RoomResponse, error)
}

func (m *mockService) CreateRoom(ctx context.Context, req *api.CreateRoomRequest) (*api.RoomResponse, error) {
	if m.createRoomFunc != nil {
		return m.createRoomFunc(ctx, req)
	}
	return &api.RoomResponse{}, nil
}

func (m *mockService) JoinRoom(ctx context.Context, req *api.JoinRoomRequest) (*api.RoomResponse, error) {
	if m.joinRoomFunc != nil {
		return m.joinRoomFunc(ctx, req)
	}
	return &api.RoomResponse{}, nil
}

func (m *mockService) SetPlayback(ctx context.Context, req *api.SetPlaybackRequest) (*api.RoomResponse, error) {
	if m.setPlaybackFunc != nil {
		return m.setPlaybackFunc(ctx, req)
	}
	return &api.RoomResponse{}, nil
}

func (m *mockService) GetState(ctx context.Context, req *api.GetStateRequest) (*api.RoomResponse, error) {
	if m.getStateFunc != nil {
		return m.getStateFunc(ctx, req)
	}
	return &api.RoomResponse{}, nil
}

func TestCreateRoomHandler(t *testing.T) {
	// Подготовка
	testRoom := &api.RoomResponse{
		RoomId:    "ABC123",
		HostId:    "user1",
		TrackUrl:  "song.mp3",
		Timestamp: time.Now().Unix(),
		Status:    "pause",
		Users:     []string{"user1"},
	}

	mockSvc := &mockService{
		createRoomFunc: func(ctx context.Context, req *api.CreateRoomRequest) (*api.RoomResponse, error) {
			if req.HostId != "user1" {
				t.Errorf("Ожидался hostId=user1, получил %s", req.HostId)
			}
			if req.TrackUrl != "song.mp3" {
				t.Errorf("Ожидался trackUrl=song.mp3, получил %s", req.TrackUrl)
			}
			return testRoom, nil
		},
	}

	handler := NewRoomHandler(mockSvc, nil)

	// Выполнение
	req := &api.CreateRoomRequest{
		HostId:   "user1",
		TrackUrl: "song.mp3",
	}

	resp, err := handler.CreateRoom(context.Background(), req)

	// Проверка
	if err != nil {
		t.Fatalf("Ошибка не ожидалась: %v", err)
	}

	if resp.RoomId != "ABC123" {
		t.Errorf("Ожидался RoomId=ABC123, получил %s", resp.RoomId)
	}
	if resp.HostId != "user1" {
		t.Errorf("Ожидался HostId=user1, получил %s", resp.HostId)
	}
	if resp.Status != "pause" {
		t.Errorf("Ожидался Status=pause, получил %s", resp.Status)
	}
}

func TestJoinRoomHandler(t *testing.T) {
	// Подготовка
	testRoom := &api.RoomResponse{
		RoomId:    "ABC123",
		HostId:    "host1",
		TrackUrl:  "song.mp3",
		Timestamp: time.Now().Unix(),
		Status:    "play",
		Users:     []string{"host1", "user2"},
	}

	mockSvc := &mockService{
		joinRoomFunc: func(ctx context.Context, req *api.JoinRoomRequest) (*api.RoomResponse, error) {
			if req.RoomId != "ABC123" {
				t.Errorf("Ожидался RoomId=ABC123, получил %s", req.RoomId)
			}
			if req.UserId != "user2" {
				t.Errorf("Ожидался UserId=user2, получил %s", req.UserId)
			}
			return testRoom, nil
		},
	}

	handler := NewRoomHandler(mockSvc, nil)

	// Выполнение
	req := &api.JoinRoomRequest{
		RoomId: "ABC123",
		UserId: "user2",
	}

	resp, err := handler.JoinRoom(context.Background(), req)

	// Проверка
	if err != nil {
		t.Fatalf("Ошибка не ожидалась: %v", err)
	}

	if len(resp.Users) != 2 {
		t.Errorf("Ожидалось 2 пользователя, получил %d", len(resp.Users))
	}
	if resp.Users[1] != "user2" {
		t.Errorf("Ожидался user2 в списке пользователей")
	}
}

func TestSetPlaybackHandler(t *testing.T) {
	// Подготовка
	testRoom := &api.RoomResponse{
		RoomId:    "ABC123",
		HostId:    "host1",
		TrackUrl:  "song.mp3",
		Timestamp: 150,
		Status:    "play",
		Users:     []string{"host1"},
	}

	mockSvc := &mockService{
		setPlaybackFunc: func(ctx context.Context, req *api.SetPlaybackRequest) (*api.RoomResponse, error) {
			if req.RoomId != "ABC123" {
				t.Errorf("Ожидался RoomId=ABC123, получил %s", req.RoomId)
			}
			if req.Action != "play" {
				t.Errorf("Ожидался Action=play, получил %s", req.Action)
			}
			if req.Timestamp != 150 {
				t.Errorf("Ожидался Timestamp=150, получил %d", req.Timestamp)
			}
			return testRoom, nil
		},
	}

	handler := NewRoomHandler(mockSvc, nil)

	// Выполнение
	req := &api.SetPlaybackRequest{
		RoomId:    "ABC123",
		Action:    "play",
		Timestamp: 150,
	}

	resp, err := handler.SetPlayback(context.Background(), req)

	// Проверка
	if err != nil {
		t.Fatalf("Ошибка не ожидалась: %v", err)
	}

	if resp.Status != "play" {
		t.Errorf("Ожидался Status=play, получил %s", resp.Status)
	}
	if resp.Timestamp != 150 {
		t.Errorf("Ожидался Timestamp=150, получил %d", resp.Timestamp)
	}
}

func TestGetStateHandler(t *testing.T) {
	// Подготовка
	testRoom := &api.RoomResponse{
		RoomId:    "ABC123",
		HostId:    "host1",
		TrackUrl:  "song.mp3",
		Timestamp: 100,
		Status:    "pause",
		Users:     []string{"host1"},
	}

	mockSvc := &mockService{
		getStateFunc: func(ctx context.Context, req *api.GetStateRequest) (*api.RoomResponse, error) {
			if req.RoomId != "ABC123" {
				t.Errorf("Ожидался RoomId=ABC123, получил %s", req.RoomId)
			}
			return testRoom, nil
		},
	}

	handler := NewRoomHandler(mockSvc, nil)

	// Выполнение
	req := &api.GetStateRequest{
		RoomId: "ABC123",
	}

	resp, err := handler.GetState(context.Background(), req)

	// Проверка
	if err != nil {
		t.Fatalf("Ошибка не ожидалась: %v", err)
	}

	if resp.RoomId != "ABC123" {
		t.Errorf("Ожидался RoomId=ABC123, получил %s", resp.RoomId)
	}
	if resp.Status != "pause" {
		t.Errorf("Ожидался Status=pause, получил %s", resp.Status)
	}
}

func TestHandlerErrorPropagation(t *testing.T) {
	// Проверяем, что ошибки из сервиса правильно передаются
	mockSvc := &mockService{
		getStateFunc: func(ctx context.Context, req *api.GetStateRequest) (*api.RoomResponse, error) {
			return nil, fmt.Errorf("room not found")
		},
	}

	handler := NewRoomHandler(mockSvc, nil)

	// Выполнение
	req := &api.GetStateRequest{
		RoomId: "NONEXISTENT",
	}

	_, err := handler.GetState(context.Background(), req)

	// Проверка
	if err == nil {
		t.Error("Ожидалась ошибка, но её нет")
	}
}

// func TestNilRequestHandling(t *testing.T) {
// 	mockSvc := &mockService{}
// 	handler := NewRoomHandler(mockSvc, nil)

// 	// Все методы должны корректно обрабатывать nil
// 	methods := []struct {
// 		name string
// 		call func() error
// 	}{
// 		{
// 			name: "CreateRoom",
// 			call: func() error {
// 				_, err := handler.CreateRoom(context.Background(), nil)
// 				return err
// 			},
// 		},
// 		{
// 			name: "JoinRoom",
// 			call: func() error {
// 				_, err := handler.JoinRoom(context.Background(), nil)
// 				return err
// 			},
// 		},
// 		{
// 			name: "SetPlayback",
// 			call: func() error {
// 				_, err := handler.SetPlayback(context.Background(), nil)
// 				return err
// 			},
// 		},
// 		{
// 			name: "GetState",
// 			call: func() error {
// 				_, err := handler.GetState(context.Background(), nil)
// 				return err
// 			},
// 		},
// 	}

// 	for _, method := range methods {
// 		t.Run(method.name+" with nil request", func(t *testing.T) {
// 			err := method.call()
// 			if err == nil {
// 				t.Errorf("%s должен возвращать ошибку при nil запросе", method.name)
// 			}
// 		})
// 	}
// }

func TestHandlerInitialization(t *testing.T) {
	// Проверяем, что обработчик создается корректно
	mockSvc := &mockService{}

	// Создаем обработчик
	handler := NewRoomHandler(mockSvc, nil)

	if handler == nil {
		t.Fatal("Обработчик не должен быть nil")
	}

	if handler.svc != mockSvc {
		t.Error("Сервис должен быть установлен в обработчике")
	}
}

// Простой тест контекста
func TestContextPropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем контекст сразу

	mockSvc := &mockService{
		createRoomFunc: func(ctx context.Context, req *api.CreateRoomRequest) (*api.RoomResponse, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return &api.RoomResponse{}, nil
			}
		},
	}

	handler := NewRoomHandler(mockSvc, nil)
	req := &api.CreateRoomRequest{
		HostId:   "test",
		TrackUrl: "test",
	}

	// Вызываем с отмененным контекстом
	_, err := handler.CreateRoom(ctx, req)

	if err == nil {
		t.Error("Ожидалась ошибка отмены контекста")
	}
}

// Простой тест проверки интерфейса
func TestHandlerImplementsInterface(t *testing.T) {
	var _ api.RoomServiceServer = &RoomHandler{}
}

// Дополнительный тест: проверка всех методов вместе
func TestAllMethodsTogether(t *testing.T) {
	// Создаем тестовую комнату
	roomID := "TEST123"
	hostID := "HOST001"

	// Мок, который сохраняет состояние между вызовами
	roomState := &api.RoomResponse{}

	mockSvc := &mockService{
		createRoomFunc: func(ctx context.Context, req *api.CreateRoomRequest) (*api.RoomResponse, error) {
			roomState = &api.RoomResponse{
				RoomId:    roomID,
				HostId:    req.HostId,
				TrackUrl:  req.TrackUrl,
				Timestamp: time.Now().Unix(),
				Status:    "pause",
				Users:     []string{req.HostId},
			}
			return roomState, nil
		},
		getStateFunc: func(ctx context.Context, req *api.GetStateRequest) (*api.RoomResponse, error) {
			if req.RoomId == roomID {
				return roomState, nil
			}
			return nil, fmt.Errorf("not found")
		},
	}

	handler := NewRoomHandler(mockSvc, nil)

	// 1. Создаем комнату
	createResp, err := handler.CreateRoom(context.Background(), &api.CreateRoomRequest{
		HostId:   hostID,
		TrackUrl: "test_song.mp3",
	})

	if err != nil {
		t.Fatalf("Не удалось создать комнату: %v", err)
	}

	if createResp.RoomId != roomID {
		t.Errorf("Неверный ID комнаты: %s", createResp.RoomId)
	}

	// 2. Получаем состояние
	getResp, err := handler.GetState(context.Background(), &api.GetStateRequest{
		RoomId: roomID,
	})

	if err != nil {
		t.Fatalf("Не удалось получить состояние: %v", err)
	}

	if getResp.HostId != hostID {
		t.Errorf("Неверный host ID: %s", getResp.HostId)
	}
}

// Самый простой тест: просто проверяем, что код работает
func TestSimpleHandler(t *testing.T) {
	// Самый базовый тест без сложной логики
	t.Run("basic handler test", func(t *testing.T) {
		handler := NewRoomHandler(nil, nil)

		if handler == nil {
			t.Error("Handler should not be nil")
		}
	})
}
