package grpc

import (
	"context"
	"errors"
	"main/internal/domain"
	"main/internal/service"
	user "main/pkg/api"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock для UserService
type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) CreateUser(ctx context.Context, req *service.CreateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) GetUser(ctx context.Context, uuid string) (*domain.User, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) UpdateUser(ctx context.Context, req *service.UpdateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) DeleteUser(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func (m *mockUserService) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}

func (m *mockUserService) AboutMeUser(ctx context.Context, req *user.AboutMeRequest) (*user.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserResponse), args.Error(1)
}

func (m *mockUserService) SetOnlineUser(ctx context.Context, uuid string, ttl time.Duration) error {
	args := m.Called(ctx, uuid, ttl)
	return args.Error(0)
}

func (m *mockUserService) SetOffline(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func (m *mockUserService) IsOnline(ctx context.Context, uuid string) (bool, error) {
	args := m.Called(ctx, uuid)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserService) GetOnlineUsers(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func TestCreateUser_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()

	protoReq := &user.CreateUserRequest{
		Uuid:     "test-uuid",
		Email:    "test@example.com",
		UserName: "Test User",
		Avatar:   "avatar.jpg",
		AboutMe:  "About me",
	}

	expectedDomainUser := &domain.User{
		UUID:      "test-uuid",
		Email:     "test@example.com",
		UserName:  "Test User",
		Avatar:    stringToPtr("avatar.jpg"),
		AboutMe:   stringToPtr("About me"),
		Status:    "offline",
		Friends:   []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService := new(mockUserService)
	serviceReq := &service.CreateUserRequest{
		UUID:     "test-uuid",
		Email:    "test@example.com",
		UserName: "Test User",
		Avatar:   "avatar.jpg",
		AboutMe:  "About me",
	}

	mockService.On("CreateUser", ctx, serviceReq).Return(expectedDomainUser, nil)

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.CreateUser(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-uuid", resp.Uuid)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "Test User", resp.UserName)
	assert.Equal(t, "avatar.jpg", resp.Avatar)
	assert.Equal(t, "About me", resp.AboutMe)

	mockService.AssertExpectations(t)
}

func TestCreateUser_ServiceError(t *testing.T) {
	// Arrange
	ctx := context.Background()

	protoReq := &user.CreateUserRequest{
		Uuid:     "test-uuid",
		Email:    "test@example.com",
		UserName: "Test User",
	}

	mockService := new(mockUserService)
	serviceReq := &service.CreateUserRequest{
		UUID:     "test-uuid",
		Email:    "test@example.com",
		UserName: "Test User",
		Avatar:   "",
		AboutMe:  "",
	}

	mockService.On("CreateUser", ctx, serviceReq).Return(nil, errors.New("database error"))

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.CreateUser(ctx, protoReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Проверяем что это gRPC ошибка
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "database error")

	mockService.AssertExpectations(t)
}

func TestGetUser_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	uuid := "test-uuid"

	protoReq := &user.GetUserRequest{
		Uuid: uuid,
	}

	expectedDomainUser := &domain.User{
		UUID:      uuid,
		Email:     "test@example.com",
		UserName:  "Test User",
		Status:    "online",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService := new(mockUserService)
	mockService.On("GetUser", ctx, uuid).Return(expectedDomainUser, nil)

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.GetUser(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uuid, resp.Uuid)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "Test User", resp.UserName)
	assert.Equal(t, "online", resp.Status)

	mockService.AssertExpectations(t)
}

func TestGetUser_NotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	uuid := "non-existent-uuid"

	protoReq := &user.GetUserRequest{
		Uuid: uuid,
	}

	mockService := new(mockUserService)
	mockService.On("GetUser", ctx, uuid).Return(nil, errors.New("user not found"))

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.GetUser(ctx, protoReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "user not found")

	mockService.AssertExpectations(t)
}

func TestGetUser_OtherError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	uuid := "test-uuid"

	protoReq := &user.GetUserRequest{
		Uuid: uuid,
	}

	mockService := new(mockUserService)
	mockService.On("GetUser", ctx, uuid).Return(nil, errors.New("connection error"))

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.GetUser(ctx, protoReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())

	mockService.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()

	protoReq := &user.UpdateUserRequest{
		Uuid:     "test-uuid",
		Email:    "updated@example.com",
		UserName: "Updated User",
		Avatar:   "new-avatar.jpg",
		AboutMe:  "Updated about me",
	}

	expectedDomainUser := &domain.User{
		UUID:      "test-uuid",
		Email:     "updated@example.com",
		UserName:  "Updated User",
		Avatar:    stringToPtr("new-avatar.jpg"),
		AboutMe:   stringToPtr("Updated about me"),
		Status:    "online",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	mockService := new(mockUserService)
	serviceReq := &service.UpdateUserRequest{
		Uuid:     "test-uuid",
		Email:    "updated@example.com",
		UserName: "Updated User",
		Avatar:   "new-avatar.jpg",
		AboutMe:  "Updated about me",
	}

	mockService.On("UpdateUser", ctx, serviceReq).Return(expectedDomainUser, nil)

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.UpdateUser(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-uuid", resp.Uuid)
	assert.Equal(t, "updated@example.com", resp.Email)
	assert.Equal(t, "Updated User", resp.UserName)
	assert.Equal(t, "new-avatar.jpg", resp.Avatar)
	assert.Equal(t, "Updated about me", resp.AboutMe)

	mockService.AssertExpectations(t)
}

func TestUpdateUser_NotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()

	protoReq := &user.UpdateUserRequest{
		Uuid:  "non-existent-uuid",
		Email: "updated@example.com",
	}

	mockService := new(mockUserService)
	serviceReq := &service.UpdateUserRequest{
		Uuid:     "non-existent-uuid",
		Email:    "updated@example.com",
		UserName: "",
		Avatar:   "",
		AboutMe:  "",
	}

	mockService.On("UpdateUser", ctx, serviceReq).Return(nil, errors.New("user not found"))

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.UpdateUser(ctx, protoReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())

	mockService.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	uuid := "test-uuid"

	protoReq := &user.DeleteUserRequest{
		Uuid: uuid,
	}

	mockService := new(mockUserService)
	mockService.On("DeleteUser", ctx, uuid).Return(nil)

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.DeleteUser(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "user deleted successfully", resp.Message)

	mockService.AssertExpectations(t)
}

func TestDeleteUser_NotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	uuid := "non-existent-uuid"

	protoReq := &user.DeleteUserRequest{
		Uuid: uuid,
	}

	mockService := new(mockUserService)
	mockService.On("DeleteUser", ctx, uuid).Return(errors.New("user not found"))

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.DeleteUser(ctx, protoReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())

	mockService.AssertExpectations(t)
}

func TestListUsers_ReturnsEmpty(t *testing.T) {
	// Arrange
	ctx := context.Background()

	protoReq := &user.ListUsersRequest{
		Limit:  10,
		Offset: 0,
	}

	handler := &userHandler{
		userService: nil, // ListUsers пока не реализован
	}

	// Act
	resp, err := handler.ListUsers(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Users)
	assert.Equal(t, int32(0), resp.Total)
}

func TestAboutMeUser_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	uuid := "test-uuid"

	protoReq := &user.AboutMeRequest{
		Uuid: uuid,
	}

	expectedProtoResp := &user.UserResponse{
		Uuid:      uuid,
		Email:     "test@example.com",
		UserName:  "Test User",
		Status:    "online",
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	mockService := new(mockUserService)
	mockService.On("AboutMeUser", ctx, protoReq).Return(expectedProtoResp, nil)

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.AboutMeUser(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uuid, resp.Uuid)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "Test User", resp.UserName)

	mockService.AssertExpectations(t)
}

func TestSetOnline_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()

	protoReq := &user.SetOnlineRequest{
		Uuid:       "test-uuid",
		TtlSeconds: 30,
	}

	mockService := new(mockUserService)
	mockService.On("SetOnlineUser", ctx, "test-uuid", 30*time.Second).Return(nil)

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.SetOnline(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "user set online", resp.Message)

	mockService.AssertExpectations(t)
}

func TestSetOffline_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()

	protoReq := &user.SetOfflineRequest{
		Uuid: "test-uuid",
	}

	mockService := new(mockUserService)
	mockService.On("SetOffline", ctx, "test-uuid").Return(nil)

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.SetOffline(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "user set offline", resp.Message)

	mockService.AssertExpectations(t)
}

func TestIsOnline_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	uuid := "test-uuid"

	protoReq := &user.IsOnlineRequest{
		Uuid: uuid,
	}

	mockService := new(mockUserService)
	mockService.On("IsOnline", ctx, uuid).Return(true, nil)

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.IsOnline(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uuid, resp.Uuid)
	assert.True(t, resp.Online)

	mockService.AssertExpectations(t)
}

func TestIsOnline_ServiceError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	uuid := "test-uuid"

	protoReq := &user.IsOnlineRequest{
		Uuid: uuid,
	}

	mockService := new(mockUserService)
	mockService.On("IsOnline", ctx, uuid).Return(false, errors.New("redis error"))

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.IsOnline(ctx, protoReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())

	mockService.AssertExpectations(t)
}

func TestGetOnlineUsers_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()

	protoReq := &user.GetOnlineUsersRequest{}

	expectedUsers := []string{"user1", "user2", "user3"}

	mockService := new(mockUserService)
	mockService.On("GetOnlineUsers", ctx).Return(expectedUsers, nil)

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.GetOnlineUsers(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 3, len(resp.Uuids))
	assert.Equal(t, []string{"user1", "user2", "user3"}, resp.Uuids)

	mockService.AssertExpectations(t)
}

func TestGetOnlineUsers_Empty(t *testing.T) {
	// Arrange
	ctx := context.Background()

	protoReq := &user.GetOnlineUsersRequest{}

	expectedUsers := []string{}

	mockService := new(mockUserService)
	mockService.On("GetOnlineUsers", ctx).Return(expectedUsers, nil)

	handler := &userHandler{
		userService: mockService,
	}

	// Act
	resp, err := handler.GetOnlineUsers(ctx, protoReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Uuids)

	mockService.AssertExpectations(t)
}

func TestNewUserHandler(t *testing.T) {
	// Arrange
	mockService := new(mockUserService)

	// Act
	handler := NewUserHandler(mockService)

	// Assert
	assert.NotNil(t, handler)

	// Проверяем что handler реализует интерфейс
	var _ user.UserServiceServer = handler
}

// Вспомогательная функция
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
