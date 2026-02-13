package service

import (
	"context"
	"fmt"

	"api_gateway/client"
	pb "api_gateway/pkg/api/auth"
)

type AuthService interface {
	Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error)
	Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
	PasswordChange(ctx context.Context, req *pb.PasswordChangeRequest) (*pb.PasswordChangeResponse, error)
	ValidateToken(ctx context.Context, token string) bool
}

type authService struct {
	authClient pb.AuthServiceClient
}

func NewAuthService(addr string) AuthService {
	// Создаем gRPC клиент для auth-service
	authClient, err := client.NewAuthClient(addr)
	if err != nil {
		fmt.Printf("Warning: failed to create auth client: %v\n", err)
		return nil
	}

	return &authService{
		authClient: authClient,
	}
}

func (s *authService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if s.authClient == nil {
		return nil, fmt.Errorf("auth service not available")
	}
	return s.authClient.Register(ctx, req)
}

func (s *authService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if s.authClient == nil {
		return nil, fmt.Errorf("auth service not available")
	}
	return s.authClient.Login(ctx, req)
}

func (s *authService) PasswordChange(ctx context.Context, req *pb.PasswordChangeRequest) (*pb.PasswordChangeResponse, error) {
	if s.authClient == nil {
		return nil, fmt.Errorf("auth service not available")
	}
	return s.authClient.PasswordChange(ctx, req)
}
func (s *authService) ValidateToken(ctx context.Context, token string) bool {
	// Временная реализация
	return token != "" && token != "invalid"
}
