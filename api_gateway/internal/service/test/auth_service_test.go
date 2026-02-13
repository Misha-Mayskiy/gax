// internal/service/test/auth_service_test.go
package test

import (
	"context"
	"testing"

	"api_gateway/internal/service"
	pb "api_gateway/pkg/api/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockAuthClient реализация мока для AuthServiceClient
type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Register(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.RegisterResponse), args.Error(1)
}

func (m *MockAuthClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.LoginResponse), args.Error(1)
}

func (m *MockAuthClient) PasswordChange(ctx context.Context, in *pb.PasswordChangeRequest, opts ...grpc.CallOption) (*pb.PasswordChangeResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.PasswordChangeResponse), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	t.Run("Register успешный", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		authService := service.NewAuthServiceWithClient(mockClient)

		req := &pb.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		expectedResponse := &pb.RegisterResponse{
			Uuid:    "user-123",
			Message: "Registration successful",
		}

		mockClient.On("Register", mock.Anything, req).Return(expectedResponse, nil)

		resp, err := authService.Register(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})

	t.Run("Register когда сервис недоступен", func(t *testing.T) {
		// Создаем сервис без клиента
		// mockClient := new(MockAuthClient)
		authService := service.NewAuthServiceWithClient(nil)

		req := &pb.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		resp, err := authService.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "auth service not available")
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("Login успешный", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		authService := service.NewAuthServiceWithClient(mockClient)

		req := &pb.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		expectedResponse := &pb.LoginResponse{
			Uuid:    "user-123",
			Success: true,
			Message: "Login successful",
		}

		mockClient.On("Login", mock.Anything, req).Return(expectedResponse, nil)

		resp, err := authService.Login(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})

	t.Run("Login с неверными данными", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		authService := service.NewAuthServiceWithClient(mockClient)

		req := &pb.LoginRequest{
			Email:    "wrong@example.com",
			Password: "wrongpass",
		}

		expectedResponse := &pb.LoginResponse{
			Uuid:    "",
			Success: false,
			Message: "Invalid credentials",
		}

		mockClient.On("Login", mock.Anything, req).Return(expectedResponse, nil)

		resp, err := authService.Login(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestAuthService_PasswordChange(t *testing.T) {
	t.Run("PasswordChange успешный", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		authService := service.NewAuthServiceWithClient(mockClient)

		req := &pb.PasswordChangeRequest{
			Uuid:        "user-123",
			OldPassword: "oldpass",
			NewPassword: "newpass",
		}

		expectedResponse := &pb.PasswordChangeResponse{
			Success: true,
			Message: "Password changed successfully",
		}

		mockClient.On("PasswordChange", mock.Anything, req).Return(expectedResponse, nil)

		resp, err := authService.PasswordChange(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	t.Run("ValidateToken валидный токен", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		authService := service.NewAuthServiceWithClient(mockClient)

		result := authService.ValidateToken(context.Background(), "valid-token")
		assert.True(t, result)
	})

	t.Run("ValidateToken пустой токен", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		authService := service.NewAuthServiceWithClient(mockClient)

		result := authService.ValidateToken(context.Background(), "")
		assert.False(t, result)
	})

	t.Run("ValidateToken невалидный токен", func(t *testing.T) {
		mockClient := new(MockAuthClient)
		authService := service.NewAuthServiceWithClient(mockClient)

		result := authService.ValidateToken(context.Background(), "invalid")
		assert.False(t, result)
	})
}
