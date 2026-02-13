package service

import (
	"context"
	"fmt"

	"api_gateway/client"
	pb "api_gateway/pkg/api/room"
)

// Интерфейс RoomService
type RoomService interface {
	CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.RoomResponse, error)
	JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*pb.RoomResponse, error)
	SetPlayback(ctx context.Context, req *pb.SetPlaybackRequest) (*pb.RoomResponse, error)
	GetState(ctx context.Context, req *pb.GetStateRequest) (*pb.RoomResponse, error)

	// Дополнительные методы для WebSocket
	UserHasAccess(roomID, userID string) (bool, error)
	GetRoomState(roomID string) (*pb.RoomResponse, error)
	UpdatePlaybackState(roomID, action string, position, volume float64) error
}

type roomService struct {
	roomclient pb.RoomServiceClient
}

// CreateRoom implements RoomService.
func (r *roomService) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.RoomResponse, error) {
	if r.roomclient == nil {
		return nil, fmt.Errorf("room service not available")
	}
	return r.roomclient.CreateRoom(ctx, req)
}

// GetState implements RoomService.
func (r *roomService) GetState(ctx context.Context, req *pb.GetStateRequest) (*pb.RoomResponse, error) {
	if r.roomclient == nil {
		return nil, fmt.Errorf("room service not available")
	}
	return r.roomclient.GetState(ctx, req)
}

// JoinRoom implements RoomService.
func (r *roomService) JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*pb.RoomResponse, error) {
	if r.roomclient == nil {
		return nil, fmt.Errorf("room service not available")
	}
	return r.roomclient.JoinRoom(ctx, req)
}

// SetPlayback implements RoomService.
func (r *roomService) SetPlayback(ctx context.Context, req *pb.SetPlaybackRequest) (*pb.RoomResponse, error) {
	if r.roomclient == nil {
		return nil, fmt.Errorf("room service not available")
	}
	return r.roomclient.SetPlayback(ctx, req)
}

// UserHasAccess - проверяет доступ пользователя к комнате
func (r *roomService) UserHasAccess(roomID, userID string) (bool, error) {
	// Для простоты - всегда разрешаем доступ
	// В реальном приложении нужно проверить в базе данных
	return true, nil
}

// GetRoomState - получает состояние комнаты по ID
func (r *roomService) GetRoomState(roomID string) (*pb.RoomResponse, error) {
	req := &pb.GetStateRequest{
		RoomId: roomID,
	}
	return r.GetState(context.Background(), req)
}

// UpdatePlaybackState - обновляет состояние воспроизведения
func (r *roomService) UpdatePlaybackState(roomID, action string, position, volume float64) error {
	req := &pb.SetPlaybackRequest{
		RoomId:    roomID,
		Action:    action,
		Timestamp: int64(position), // Конвертируем position в timestamp
	}
	_, err := r.SetPlayback(context.Background(), req)
	return err
}

func NewRoomService(addr string) RoomService {
	roomclient, err := client.NewRoomClient(addr)
	if err != nil {
		fmt.Printf("Warning: failed to create room client: %v\n", err)
		return &roomService{roomclient: nil}
	}
	return &roomService{
		roomclient: roomclient,
	}
}
