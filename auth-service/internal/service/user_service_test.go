package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository - мок для репозитория
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, uuid, username, email, password string) error {
	args := m.Called(ctx, uuid, username, email, password)
	return args.Error(0)
}

func (m *MockUserRepository) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(uuid, newPassword string) error {
	args := m.Called(uuid, newPassword)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo) // Создаем РЕАЛЬНЫЙ сервис с мок-репозиторием

	ctx := context.Background()
	username := "testuser"
	email := "test@example.com"
	password := "12345"

	// Настраиваем ожидание: Create должен быть вызван
	mockRepo.On("Create", ctx, mock.AnythingOfType("string"), username, email, mock.AnythingOfType("string")).
		Return(nil)

	uuid, err := svc.Register(ctx, username, email, password)

	assert.NoError(t, err)
	assert.NotEmpty(t, uuid)
	mockRepo.AssertExpectations(t)
}

func TestLogin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)

	ctx := context.Background()
	email := "test@example.com"
	password := "12345"
	expectedUUID := "user-uuid-123"

	mockRepo.On("Login", ctx, email, password).Return(expectedUUID, nil)

	uuid, err := svc.Login(ctx, email, password)

	assert.NoError(t, err)
	assert.Equal(t, expectedUUID, uuid)
}

func TestChangePassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)

	ctx := context.Background()
	uuid := "user-uuid"
	newPass := "newpass"

	// Сервис хеширует пароль, поэтому в репозиторий придет хеш (строка)
	mockRepo.On("UpdatePassword", uuid, mock.AnythingOfType("string")).Return(nil)

	err := svc.ChangePassword(ctx, uuid, newPass)

	assert.NoError(t, err)
}
