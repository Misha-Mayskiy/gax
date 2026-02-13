package v1

import (
	"auth-service/internal/service"
	"auth-service/pkg/api"
	"context"
)

type GRPCServer struct {
	service *service.UserService
	api.AuthServiceServer
}

func NewGRPCServer(svc *service.UserService) *GRPCServer {
	return &GRPCServer{
		service: svc,
	}
}

func (gs *GRPCServer) Register(ctx context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	userUuid, err := gs.service.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return &api.RegisterResponse{
			Uuid:    "",
			Message: err.Error(),
		}, err
	}
	return &api.RegisterResponse{
		Uuid:    userUuid,
		Message: "Registration successful",
	}, nil
}
func (gs *GRPCServer) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	userUuid, err := gs.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return &api.LoginResponse{
			Uuid:    "",
			Success: false,
			Message: err.Error(),
		}, err
	}
	return &api.LoginResponse{
		Uuid:    userUuid,
		Success: true,
		Message: "Login successful",
	}, nil
}
func (gs *GRPCServer) PasswordChange(ctx context.Context, req *api.PasswordChangeRequest) (*api.PasswordChangeResponse, error) {
	err := gs.service.ChangePassword(ctx, req.Uuid, req.NewPassword)
	if err != nil {
		return &api.PasswordChangeResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}
	return &api.PasswordChangeResponse{
		Success: true,
		Message: "Password changed successfully",
	}, nil
}
