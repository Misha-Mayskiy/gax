package grpc

import (
	"context"
	"main/internal/domain"
	chatpb "main/pkg/api"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ==================== МОКИ ====================

// MockChatService - мок для service.ChatServiceInterface
type MockChatService struct {
	mock.Mock
}

func (m *MockChatService) CreateDirect(ctx context.Context, userID, peerID string) (domain.Chat, error) {
	args := m.Called(ctx, userID, peerID)
	return args.Get(0).(domain.Chat), args.Error(1)
}

func (m *MockChatService) CreateGroup(ctx context.Context, creatorID string, members []string, title string) (domain.Chat, error) {
	args := m.Called(ctx, creatorID, members, title)
	return args.Get(0).(domain.Chat), args.Error(1)
}

func (m *MockChatService) SendMessage(ctx context.Context, msg domain.Message) (domain.Message, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).(domain.Message), args.Error(1)
}

func (m *MockChatService) UpdateMessage(ctx context.Context, messageID, authorID string, text *string, media *[]domain.Media) (domain.Message, error) {
	args := m.Called(ctx, messageID, authorID, text, media)
	return args.Get(0).(domain.Message), args.Error(1)
}

func (m *MockChatService) DeleteMessage(messageIDs []string, hard bool, requesterID string) ([]domain.Message, error) {
	args := m.Called(messageIDs, hard, requesterID)
	return args.Get(0).([]domain.Message), args.Error(1)
}

func (m *MockChatService) ListMessages(ctx context.Context, chatID string, limit int, cursor string) ([]domain.Message, string, error) {
	args := m.Called(ctx, chatID, limit, cursor)
	return args.Get(0).([]domain.Message), args.String(1), args.Error(2)
}

func (m *MockChatService) MarkRead(ctx context.Context, chatID, userID, messageID string) error {
	args := m.Called(ctx, chatID, userID, messageID)
	return args.Error(0)
}

func (m *MockChatService) ToggleSaved(ctx context.Context, userID, messageID string, saved bool) error {
	args := m.Called(ctx, userID, messageID, saved)
	return args.Error(0)
}

func (m *MockChatService) ListSaved(ctx context.Context, userID string, limit int, cursor string) ([]domain.Message, string, error) {
	args := m.Called(ctx, userID, limit, cursor)
	return args.Get(0).([]domain.Message), args.String(1), args.Error(2)
}

func (m *MockChatService) ListChats(ctx context.Context, userID string, limit int, cursor string) ([]domain.Chat, string, error) {
	args := m.Called(ctx, userID, limit, cursor)
	return args.Get(0).([]domain.Chat), args.String(1), args.Error(2)
}

func (m *MockChatService) GetChat(ctx context.Context, chatID string) (domain.Chat, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).(domain.Chat), args.Error(1)
}

func (m *MockChatService) UpdateGroupChat(ctx context.Context, req *chatpb.UpdateGroupChatRequest) (domain.Chat, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(domain.Chat), args.Error(1)
}

func (m *MockChatService) ListReadMessages(ctx context.Context, userID, chatID string, limit int) ([]domain.Message, error) {
	args := m.Called(ctx, userID, chatID, limit)
	return args.Get(0).([]domain.Message), args.Error(1)
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

// createTestServer создает тестовый gRPC сервер с моком сервиса
func createTestServer() (*ChatServer, *MockChatService) {
	mockService := &MockChatService{}
	// MockChatService реализует service.ChatServiceInterface
	return NewChatServer(mockService), mockService
}

// ==================== ТЕСТЫ ====================

func TestChatServer_SendMessage_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.SendMessageRequest{
		ChatId:   "chat1",
		AuthorId: "user1",
		Text:     "Hello World",
	}

	expectedMsg := createTestMessage("msg-123", "chat1", "user1", "Hello World")

	// Настройка моков
	mockService.On("SendMessage", ctx, mock.AnythingOfType("domain.Message")).Return(expectedMsg, nil)

	// Выполнение
	resp, err := server.SendMessage(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Message)
	assert.Equal(t, "msg-123", resp.Message.Id)
	assert.Equal(t, "chat1", resp.Message.ChatId)
	assert.Equal(t, "user1", resp.Message.AuthorId)
	assert.Equal(t, "Hello World", resp.Message.Text)

	mockService.AssertExpectations(t)
}

func TestChatServer_SendMessage_ServiceError(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.SendMessageRequest{
		ChatId:   "chat1",
		AuthorId: "user1",
		Text:     "Hello World",
	}

	// Настройка моков - сервис возвращает ошибку
	mockService.On("SendMessage", ctx, mock.Anything).Return(domain.Message{}, assert.AnError)

	// Выполнение
	resp, err := server.SendMessage(ctx, req)

	// Проверки
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcErr.Code())
	assert.Contains(t, grpcErr.Message(), "failed to send message")

	mockService.AssertExpectations(t)
}

func TestChatServer_UpdateMessage_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.UpdateMessageRequest{
		MessageId: "msg-123",
		AuthorId:  "user1",
		Text:      "Updated text",
	}

	expectedMsg := createTestMessage("msg-123", "chat1", "user1", "Updated text")

	// Настройка моков
	mockService.On("UpdateMessage", ctx, "msg-123", "user1", &req.Text, (*[]domain.Media)(nil)).Return(expectedMsg, nil)

	// Выполнение
	resp, err := server.UpdateMessage(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Message)
	assert.Equal(t, "Updated text", resp.Message.Text)

	mockService.AssertExpectations(t)
}

func TestChatServer_DeleteMessage_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.DeleteMessageRequest{
		MessageIds:  []string{"msg-1", "msg-2"},
		HardDelete:  false,
		RequesterId: "user1",
	}

	// Настройка моков
	mockService.On("DeleteMessage", []string{"msg-1", "msg-2"}, false, "user1").
		Return([]domain.Message{createTestMessage("msg-1", "chat1", "user1", "Message 1")}, nil)

	// Выполнение
	resp, err := server.DeleteMessage(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "deleted", resp.Message)

	mockService.AssertExpectations(t)
}

func TestChatServer_DeleteMessage_ServiceError(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.DeleteMessageRequest{
		MessageIds:  []string{"msg-1"},
		HardDelete:  false,
		RequesterId: "user1",
	}

	// Настройка моков - сервис возвращает ошибку
	mockService.On("DeleteMessage", []string{"msg-1"}, false, "user1").
		Return([]domain.Message{}, assert.AnError)

	// Выполнение
	resp, err := server.DeleteMessage(ctx, req)

	// Проверки
	assert.NoError(t, err) // gRPC метод DeleteMessage не возвращает ошибку в прото
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "failed to delete messages")

	mockService.AssertExpectations(t)
}

func TestChatServer_ListMessages_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.ListMessagesRequest{
		ChatId: "chat1",
		Limit:  10,
		Cursor: "",
	}

	expectedMessages := []domain.Message{
		createTestMessage("msg-1", "chat1", "user1", "Message 1"),
		createTestMessage("msg-2", "chat1", "user2", "Message 2"),
	}

	// Настройка моков
	mockService.On("ListMessages", ctx, "chat1", 10, "").Return(expectedMessages, "msg-2", nil)

	// Выполнение
	resp, err := server.ListMessages(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Messages, 2)
	assert.Equal(t, "msg-2", resp.NextCursor)
	assert.Equal(t, "Message 1", resp.Messages[0].Text)

	mockService.AssertExpectations(t)
}

func TestChatServer_MarkRead_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.MarkReadRequest{
		ChatId:    "chat1",
		UserId:    "user1",
		MessageId: "msg-123",
	}

	// Настройка моков
	mockService.On("MarkRead", ctx, "chat1", "user1", "msg-123").Return(nil)

	// Выполнение
	resp, err := server.MarkRead(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)

	mockService.AssertExpectations(t)
}

func TestChatServer_MarkRead_ServiceError(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.MarkReadRequest{
		ChatId:    "chat1",
		UserId:    "user1",
		MessageId: "msg-123",
	}

	// Настройка моков - сервис возвращает ошибку
	mockService.On("MarkRead", ctx, "chat1", "user1", "msg-123").Return(assert.AnError)

	// Выполнение
	resp, err := server.MarkRead(ctx, req)

	// Проверки
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcErr.Code())
	assert.Contains(t, grpcErr.Message(), "failed to mark read")

	mockService.AssertExpectations(t)
}

func TestChatServer_ToggleSaved_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.ToggleSavedRequest{
		UserId:    "user1",
		MessageId: "msg-123",
		Saved:     true,
	}

	// Настройка моков
	mockService.On("ToggleSaved", ctx, "user1", "msg-123", true).Return(nil)

	// Выполнение
	resp, err := server.ToggleSaved(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)

	mockService.AssertExpectations(t)
}

func TestChatServer_ListSaved_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.ListSavedRequest{
		UserId: "user1",
		Limit:  10,
		Cursor: "",
	}

	expectedMessages := []domain.Message{
		createTestMessage("msg-1", "chat1", "user1", "Saved message 1"),
		createTestMessage("msg-2", "chat2", "user2", "Saved message 2"),
	}

	// Настройка моков
	mockService.On("ListSaved", ctx, "user1", 10, "").Return(expectedMessages, "msg-2", nil)

	// Выполнение
	resp, err := server.ListSaved(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Messages, 2)
	assert.Equal(t, "msg-2", resp.NextCursor)

	mockService.AssertExpectations(t)
}

func TestChatServer_CreateDirectChat_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.CreateDirectChatRequest{
		UserId: "user1",
		PeerId: "user2",
	}

	expectedChat := createTestChat("user1_user2", domain.ChatKindDirect)

	// Настройка моков
	mockService.On("CreateDirect", ctx, "user1", "user2").Return(expectedChat, nil)

	// Выполнение
	resp, err := server.CreateDirectChat(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Chat)
	assert.Equal(t, "user1_user2", resp.Chat.Id)
	assert.Equal(t, "direct", resp.Chat.Kind)
	assert.Contains(t, resp.Chat.MemberIds, "user1")
	assert.Contains(t, resp.Chat.MemberIds, "user2")

	mockService.AssertExpectations(t)
}

func TestChatServer_CreateDirectChat_NotFound(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.CreateDirectChatRequest{
		UserId: "user1",
		PeerId: "user2",
	}

	// Настройка моков - сервис возвращает ошибку "not found"
	mockService.On("CreateDirect", ctx, "user1", "user2").Return(domain.Chat{}, status.Error(codes.NotFound, "user not found"))

	// Выполнение
	resp, err := server.CreateDirectChat(ctx, req)

	// Проверки
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, grpcErr.Code())
	assert.Contains(t, grpcErr.Message(), "failed to create direct chat")

	mockService.AssertExpectations(t)
}

func TestChatServer_CreateGroupChat_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.CreateGroupChatRequest{
		UserId:    "creator1",
		MemberIds: []string{"user1", "user2"},
		Title:     "Test Group",
	}

	expectedChat := createTestChat("group-123", domain.ChatKindGroup)
	expectedChat.Title = "Test Group"
	expectedChat.MemberIDs = []string{"creator1", "user1", "user2"}

	// Настройка моков
	mockService.On("CreateGroup", ctx, "creator1", []string{"user1", "user2"}, "Test Group").Return(expectedChat, nil)

	// Выполнение
	resp, err := server.CreateGroupChat(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Chat)
	assert.Equal(t, "group", resp.Chat.Kind)
	assert.Equal(t, "Test Group", resp.Chat.Title)
	assert.Len(t, resp.Chat.MemberIds, 3)

	mockService.AssertExpectations(t)
}

func TestChatServer_ListChats_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.ListChatsRequest{
		UserId: "user1",
		Limit:  10,
		Cursor: "",
	}

	expectedChats := []domain.Chat{
		createTestChat("chat1", domain.ChatKindDirect),
		createTestChat("chat2", domain.ChatKindGroup),
	}

	// Настройка моков
	mockService.On("ListChats", ctx, "user1", 10, "").Return(expectedChats, "chat2", nil)

	// Выполнение
	resp, err := server.ListChats(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Chats, 2)
	assert.Equal(t, "chat2", resp.NextCursor)

	mockService.AssertExpectations(t)
}

func TestChatServer_GetChat_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.GetChatRequest{
		ChatId: "chat1",
	}

	expectedChat := createTestChat("chat1", domain.ChatKindDirect)

	// Настройка моков
	mockService.On("GetChat", ctx, "chat1").Return(expectedChat, nil)

	// Выполнение
	resp, err := server.GetChat(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Chat)
	assert.Equal(t, "chat1", resp.Chat.Id)
	assert.Equal(t, "direct", resp.Chat.Kind)

	mockService.AssertExpectations(t)
}

func TestChatServer_GetChat_ServiceError(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.GetChatRequest{
		ChatId: "chat1",
	}

	// Настройка моков - сервис возвращает ошибку
	mockService.On("GetChat", ctx, "chat1").Return(domain.Chat{}, assert.AnError)

	// Выполнение
	resp, err := server.GetChat(ctx, req)

	// Проверки
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcErr.Code())
	assert.Contains(t, grpcErr.Message(), "failed to get chat")

	mockService.AssertExpectations(t)
}

func TestChatServer_UpdateGroupChat_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.UpdateGroupChatRequest{
		ChatId:          "group-123",
		Title:           "New Title",
		AddMemberIds:    []string{"user3"},
		RemoveMemberIds: []string{"user4"},
		RequesterId:     "creator1",
	}

	expectedChat := createTestChat("group-123", domain.ChatKindGroup)
	expectedChat.Title = "New Title"

	// Настройка моков
	mockService.On("UpdateGroupChat", ctx, req).Return(expectedChat, nil)

	// Выполнение
	resp, err := server.UpdateGroupChat(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Chat)
	assert.Equal(t, "New Title", resp.Chat.Title)

	mockService.AssertExpectations(t)
}

func TestChatServer_UpdateGroupChat_ServiceError(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.UpdateGroupChatRequest{
		ChatId:          "group-123",
		Title:           "New Title",
		AddMemberIds:    []string{"user3"},
		RemoveMemberIds: []string{},
		RequesterId:     "creator1",
	}

	// Настройка моков - сервис возвращает ошибку
	mockService.On("UpdateGroupChat", ctx, req).Return(domain.Chat{}, assert.AnError)

	// Выполнение
	resp, err := server.UpdateGroupChat(ctx, req)

	// Проверки
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcErr.Code())
	assert.Contains(t, grpcErr.Message(), "failed to update group chat")

	mockService.AssertExpectations(t)
}

func TestChatServer_ListReadMessages_Success(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.ListReadMessagesRequest{
		UserId: "user1",
		ChatId: "chat1",
		Limit:  10,
	}

	expectedMessages := []domain.Message{
		createTestMessage("msg-1", "chat1", "user1", "Read message 1"),
		createTestMessage("msg-2", "chat1", "user2", "Read message 2"),
	}

	// Настройка моков
	mockService.On("ListReadMessages", ctx, "user1", "chat1", 10).Return(expectedMessages, nil)

	// Выполнение
	resp, err := server.ListReadMessages(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Messages, 2)
	assert.Equal(t, "Read message 1", resp.Messages[0].Text)

	mockService.AssertExpectations(t)
}

func TestChatServer_ListReadMessages_EmptyChatID(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.ListReadMessagesRequest{
		UserId: "user1",
		ChatId: "", // Пустой chat_id - получаем все прочитанные сообщения
		Limit:  10,
	}

	expectedMessages := []domain.Message{
		createTestMessage("msg-1", "chat1", "user1", "Message 1"),
		createTestMessage("msg-2", "chat2", "user1", "Message 2"),
	}

	// Настройка моков
	mockService.On("ListReadMessages", ctx, "user1", "", 10).Return(expectedMessages, nil)

	// Выполнение
	resp, err := server.ListReadMessages(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Messages, 2)

	mockService.AssertExpectations(t)
}

func TestChatServer_ListReadMessages_ServiceError(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.ListReadMessagesRequest{
		UserId: "user1",
		ChatId: "chat1",
		Limit:  10,
	}

	// Настройка моков - сервис возвращает ошибку
	mockService.On("ListReadMessages", ctx, "user1", "chat1", 10).Return([]domain.Message{}, assert.AnError)

	// Выполнение
	resp, err := server.ListReadMessages(ctx, req)

	// Проверки
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcErr.Code())
	assert.Contains(t, grpcErr.Message(), "failed to list read messages")

	mockService.AssertExpectations(t)
}

func TestChatServer_ListMessages_ServiceError(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.ListMessagesRequest{
		ChatId: "chat1",
		Limit:  10,
		Cursor: "",
	}

	// Настройка моков - сервис возвращает ошибку
	mockService.On("ListMessages", ctx, "chat1", 10, "").Return([]domain.Message{}, "", assert.AnError)

	// Выполнение
	resp, err := server.ListMessages(ctx, req)

	// Проверки
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcErr.Code())
	assert.Contains(t, grpcErr.Message(), "failed to list messages")

	mockService.AssertExpectations(t)
}

func TestChatServer_ListSaved_ServiceError(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.ListSavedRequest{
		UserId: "user1",
		Limit:  10,
		Cursor: "",
	}

	// Настройка моков - сервис возвращает ошибку
	mockService.On("ListSaved", ctx, "user1", 10, "").Return([]domain.Message{}, "", assert.AnError)

	// Выполнение
	resp, err := server.ListSaved(ctx, req)

	// Проверки
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcErr.Code())
	assert.Contains(t, grpcErr.Message(), "failed to list saved")

	mockService.AssertExpectations(t)
}

func TestChatServer_ToggleSaved_ServiceError(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.ToggleSavedRequest{
		UserId:    "user1",
		MessageId: "msg-123",
		Saved:     true,
	}

	// Настройка моков - сервис возвращает ошибку
	mockService.On("ToggleSaved", ctx, "user1", "msg-123", true).Return(assert.AnError)

	// Выполнение
	resp, err := server.ToggleSaved(ctx, req)

	// Проверки
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcErr.Code())
	assert.Contains(t, grpcErr.Message(), "failed to toggle saved")

	mockService.AssertExpectations(t)
}

func TestChatServer_UpdateMessage_ServiceError(t *testing.T) {
	// Подготовка
	server, mockService := createTestServer()
	ctx := context.Background()

	req := &chatpb.UpdateMessageRequest{
		MessageId: "msg-123",
		AuthorId:  "user1",
		Text:      "Updated text",
	}

	// Настройка моков - сервис возвращает ошибку
	mockService.On("UpdateMessage", ctx, "msg-123", "user1", &req.Text, (*[]domain.Media)(nil)).
		Return(domain.Message{}, assert.AnError)

	// Выполнение
	resp, err := server.UpdateMessage(ctx, req)

	// Проверки
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcErr.Code())
	assert.Contains(t, grpcErr.Message(), "failed to update message")

	mockService.AssertExpectations(t)
}

func TestToProtoChat(t *testing.T) {
	// Подготовка
	chat := createTestChat("chat1", domain.ChatKindGroup)
	chat.Title = "Test Group"
	chat.CreatedAt = 1234567890

	// Выполнение
	protoChat := toProtoChat(chat)

	// Проверки
	assert.Equal(t, "chat1", protoChat.Id)
	assert.Equal(t, "group", protoChat.Kind)
	assert.Equal(t, "Test Group", protoChat.Title)
	assert.Equal(t, "user1", protoChat.CreatedBy)
	assert.Equal(t, "1234567890", protoChat.CreatedAt)
	assert.Len(t, protoChat.MemberIds, 2)
	assert.Contains(t, protoChat.MemberIds, "user1")
	assert.Contains(t, protoChat.MemberIds, "user2")
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

// createTestChat создает тестовый чат
func createTestChat(id string, kind domain.ChatKind) domain.Chat {
	return domain.Chat{
		ID:        id,
		Kind:      kind,
		MemberIDs: []string{"user1", "user2"},
		Title:     "Test Chat",
		CreatedBy: "user1",
		CreatedAt: time.Now().Unix(),
	}
}

// createTestMessage создает тестовое сообщение
func createTestMessage(id, chatID, authorID, text string) domain.Message {
	return domain.Message{
		ID:        id,
		ChatID:    chatID,
		AuthorID:  authorID,
		Text:      text,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		ReadBy:    []domain.ReadInfo{},
		SavedBy:   []domain.SavedInfo{},
		Deleted:   false,
	}
}
