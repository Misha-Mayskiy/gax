package service

import (
	"context"
	"math/rand"
	"time"

	"main/internal/repository"
	api "main/pkg/api"
)

type RoomService interface {
	CreateRoom(ctx context.Context, req *api.CreateRoomRequest) (*api.RoomResponse, error)
	JoinRoom(ctx context.Context, req *api.JoinRoomRequest) (*api.RoomResponse, error)
	SetPlayback(ctx context.Context, req *api.SetPlaybackRequest) (*api.RoomResponse, error)
	GetState(ctx context.Context, req *api.GetStateRequest) (*api.RoomResponse, error)
}

type roomService struct {
	repo repository.RoomRepository
}

func NewRoomService(repo repository.RoomRepository) RoomService {
	return &roomService{repo: repo}
}

func (s *roomService) CreateRoom(ctx context.Context, req *api.CreateRoomRequest) (*api.RoomResponse, error) {
	// Генерируем уникальный ID комнаты
	roomID := generateRoomID()

	room := &api.RoomResponse{
		RoomId:    roomID,
		HostId:    req.HostId,
		TrackUrl:  req.TrackUrl,
		Timestamp: time.Now().Unix(),
		Status:    "pause",
		Users:     []string{req.HostId},
	}

	if err := s.repo.SaveRoom(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *roomService) GetState(ctx context.Context, req *api.GetStateRequest) (*api.RoomResponse, error) {
	return s.repo.GetRoom(ctx, req.RoomId)
}

func (s *roomService) JoinRoom(ctx context.Context, req *api.JoinRoomRequest) (*api.RoomResponse, error) {
	// Получаем комнату
	room, err := s.repo.GetRoom(ctx, req.RoomId)
	if err != nil {
		return nil, err
	}

	// Проверяем, есть ли уже пользователь в комнате
	userExists := false
	for _, user := range room.Users {
		if user == req.UserId {
			userExists = true
			break
		}
	}

	// Добавляем пользователя, если его нет
	if !userExists {
		room.Users = append(room.Users, req.UserId)
		room.Timestamp = time.Now().Unix()

		// Сохраняем обновленную комнату
		if err := s.repo.SaveRoom(ctx, room); err != nil {
			return nil, err
		}
	}

	return room, nil
}

func (s *roomService) SetPlayback(ctx context.Context, req *api.SetPlaybackRequest) (*api.RoomResponse, error) {
	room, err := s.repo.GetRoom(ctx, req.RoomId)
	if err != nil {
		return nil, err
	}
	room.Status = req.Action
	room.Timestamp = req.Timestamp
	if err := s.repo.SaveRoom(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func generateRoomID() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 6)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
