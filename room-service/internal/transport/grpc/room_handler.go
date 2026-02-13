package grpc

import (
	"context"
	"main/internal/service"
	"main/internal/transport/websocket"
	api "main/pkg/api"
)

type RoomHandler struct {
	svc service.RoomService
	api.UnimplementedRoomServiceServer
}

func NewRoomHandler(svc service.RoomService, wsServer *websocket.WebSocketServer) *RoomHandler {
	return &RoomHandler{svc: svc}
}

func (h *RoomHandler) CreateRoom(ctx context.Context, req *api.CreateRoomRequest) (*api.RoomResponse, error) {
	return h.svc.CreateRoom(ctx, req)
}

func (h *RoomHandler) JoinRoom(ctx context.Context, req *api.JoinRoomRequest) (*api.RoomResponse, error) {
	return h.svc.JoinRoom(ctx, req)
}

func (h *RoomHandler) SetPlayback(ctx context.Context, req *api.SetPlaybackRequest) (*api.RoomResponse, error) {
	return h.svc.SetPlayback(ctx, req)
}

func (h *RoomHandler) GetState(ctx context.Context, req *api.GetStateRequest) (*api.RoomResponse, error) {
	return h.svc.GetState(ctx, req)
}
