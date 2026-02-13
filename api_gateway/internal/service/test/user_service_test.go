// internal/service/test/user_service_test.go
package test

import (
	"context"
	"testing"

	"api_gateway/internal/service"
	pb "api_gateway/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockUserClient - ПОЛНЫЙ мок для UserServiceClient
type MockUserClient struct {
	mock.Mock
}

// Все методы из интерфейса UserServiceClient:

func (m *MockUserClient) CreateUser(ctx context.Context, in *pb.CreateUserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *MockUserClient) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *MockUserClient) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest, opts ...grpc.CallOption) (*pb.DeleteUserResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.DeleteUserResponse), args.Error(1)
}

func (m *MockUserClient) AboutMeUser(ctx context.Context, in *pb.AboutMeRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *MockUserClient) GetUser(ctx context.Context, in *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *MockUserClient) ListUsers(ctx context.Context, in *pb.ListUsersRequest, opts ...grpc.CallOption) (*pb.ListUsersResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ListUsersResponse), args.Error(1)
}

func (m *MockUserClient) SetOnline(ctx context.Context, in *pb.SetOnlineRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.StatusResponse), args.Error(1)
}

func (m *MockUserClient) SetOffline(ctx context.Context, in *pb.SetOfflineRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.StatusResponse), args.Error(1)
}

func (m *MockUserClient) IsOnline(ctx context.Context, in *pb.IsOnlineRequest, opts ...grpc.CallOption) (*pb.IsOnlineResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.IsOnlineResponse), args.Error(1)
}

func (m *MockUserClient) GetOnlineUsers(ctx context.Context, in *pb.GetOnlineUsersRequest, opts ...grpc.CallOption) (*pb.GetOnlineUsersResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.GetOnlineUsersResponse), args.Error(1)
}

// Теперь тесты
func TestUserService_Create(t *testing.T) {
	t.Run("Create успешно", func(t *testing.T) {
		mockClient := new(MockUserClient)
		userService := service.NewTestUserService(mockClient)

		uuid := "user-123"
		email := "test@example.com"
		userName := "Test User"
		avatar := "avatar.jpg"
		aboutMe := "Test user description"
		friends := []string{"user-456", "user-789"}

		expectedResponse := &pb.UserResponse{
			Uuid:     uuid,
			Email:    email,
			UserName: userName,
			Avatar:   avatar,
			AboutMe:  aboutMe,
			Friends:  friends,
		}

		mockClient.On("CreateUser", mock.Anything, &pb.CreateUserRequest{
			Uuid:     uuid,
			Email:    email,
			UserName: userName,
			Avatar:   avatar,
			AboutMe:  aboutMe,
			Friends:  friends,
		}, mock.Anything).Return(expectedResponse, nil)

		resp, err := userService.Create(context.Background(), uuid, email, userName, avatar, aboutMe, friends)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})

	t.Run("Create ошибка валидации - пустой email", func(t *testing.T) {
		mockClient := new(MockUserClient)
		userService := service.NewTestUserService(mockClient)

		resp, err := userService.Create(context.Background(), "user-123", "", "Test User", "", "", nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run("Create ошибка gRPC", func(t *testing.T) {
		mockClient := new(MockUserClient)
		userService := service.NewTestUserService(mockClient)

		mockClient.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).
			Return((*pb.UserResponse)(nil), status.Error(codes.Internal, "internal error"))

		resp, err := userService.Create(context.Background(), "user-123", "test@example.com", "Test User", "", "", nil)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "internal error")
	})
}

func TestUserService_Get(t *testing.T) {
	t.Run("Get успешно", func(t *testing.T) {
		mockClient := new(MockUserClient)
		userService := service.NewTestUserService(mockClient)

		uuid := "user-123"

		expectedResponse := &pb.UserResponse{
			Uuid:     uuid,
			Email:    "test@example.com",
			UserName: "Test User",
		}

		mockClient.On("AboutMeUser", mock.Anything, &pb.AboutMeRequest{Uuid: uuid}, mock.Anything).
			Return(expectedResponse, nil)

		resp, err := userService.Get(context.Background(), uuid)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})

	t.Run("Get ошибка - пустой uuid", func(t *testing.T) {
		mockClient := new(MockUserClient)
		userService := service.NewTestUserService(mockClient)

		resp, err := userService.Get(context.Background(), "")

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "uuid is required")
	})
}

func TestUserService_Update(t *testing.T) {
	t.Run("Update успешно", func(t *testing.T) {
		mockClient := new(MockUserClient)
		userService := service.NewTestUserService(mockClient)

		uuid := "user-123"
		email := "updated@example.com"
		userName := "Updated Name"

		expectedResponse := &pb.UserResponse{
			Uuid:     uuid,
			Email:    email,
			UserName: userName,
		}

		mockClient.On("UpdateUser", mock.Anything, &pb.UpdateUserRequest{
			Uuid:     uuid,
			Email:    email,
			UserName: userName,
		}, mock.Anything).Return(expectedResponse, nil)

		resp, err := userService.Update(context.Background(), uuid, email, userName, "", "", nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestUserService_Delete(t *testing.T) {
	t.Run("Delete успешно", func(t *testing.T) {
		mockClient := new(MockUserClient)
		userService := service.NewTestUserService(mockClient)

		uuid := "user-123"

		expectedResponse := &pb.DeleteUserResponse{
			// Заполните реальными полями из вашего protobuf
		}

		mockClient.On("DeleteUser", mock.Anything, &pb.DeleteUserRequest{Uuid: uuid}, mock.Anything).
			Return(expectedResponse, nil)

		resp, err := userService.Delete(context.Background(), uuid)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

// Дополнительные тесты
func TestUserService_SetOnline(t *testing.T) {
	t.Run("SetOnline успешно", func(t *testing.T) {
		mockClient := new(MockUserClient)
		userService := service.NewTestUserService(mockClient)

		uuid := "user-123"
		ttlSeconds := int32(300)

		expectedResponse := &pb.StatusResponse{
			Success: true,
		}

		mockClient.On("SetOnline", mock.Anything, &pb.SetOnlineRequest{
			Uuid:       uuid,
			TtlSeconds: ttlSeconds,
		}, mock.Anything).Return(expectedResponse, nil)

		resp, err := userService.SetOnline(context.Background(), uuid, ttlSeconds)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestUserService_IsOnline(t *testing.T) {
	t.Run("IsOnline успешно", func(t *testing.T) {
		mockClient := new(MockUserClient)
		userService := service.NewTestUserService(mockClient)

		uuid := "user-123"

		expectedResponse := &pb.IsOnlineResponse{
			Online: true,
		}

		mockClient.On("IsOnline", mock.Anything, &pb.IsOnlineRequest{Uuid: uuid}, mock.Anything).
			Return(expectedResponse, nil)

		resp, err := userService.IsOnline(context.Background(), uuid)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestUserService_GetOnlineUsers(t *testing.T) {
	t.Run("GetOnlineUsers успешно", func(t *testing.T) {
		mockClient := new(MockUserClient)
		userService := service.NewTestUserService(mockClient)

		expectedResponse := &pb.GetOnlineUsersResponse{
			// Заполните реальными полями из вашего protobuf
		}

		mockClient.On("GetOnlineUsers", mock.Anything, &pb.GetOnlineUsersRequest{}, mock.Anything).
			Return(expectedResponse, nil)

		resp, err := userService.GetOnlineUsers(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}
