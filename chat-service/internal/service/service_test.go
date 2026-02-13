package service

import (
	"context"
	"errors"
	"main/internal/domain"
	"testing"
	"time"

	chatpb "main/pkg/api"
	userpb "main/pkg/api_user_service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// ==================== МОКИ ====================

// MockChatRepository - мок для ChatRepository
type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) CreateDirect(userID, peerID string, userClient userpb.UserServiceClient) (domain.Chat, error) {
	args := m.Called(userID, peerID, userClient)
	return args.Get(0).(domain.Chat), args.Error(1)
}

func (m *MockChatRepository) CreateGroup(creatorID string, members []string, title string, userClient userpb.UserServiceClient) (domain.Chat, error) {
	args := m.Called(creatorID, members, title, userClient)
	return args.Get(0).(domain.Chat), args.Error(1)
}

func (m *MockChatRepository) UpdateGroup(chatID string, title *string, addMembers, removeMembers []string, requesterID string, userClient userpb.UserServiceClient) (domain.Chat, error) {
	args := m.Called(chatID, title, addMembers, removeMembers, requesterID, userClient)
	return args.Get(0).(domain.Chat), args.Error(1)
}

func (m *MockChatRepository) Get(chatID string) (domain.Chat, error) {
	args := m.Called(chatID)
	return args.Get(0).(domain.Chat), args.Error(1)
}

func (m *MockChatRepository) List(userID string, limit int, cursor string) ([]domain.Chat, string, error) {
	args := m.Called(userID, limit, cursor)
	return args.Get(0).([]domain.Chat), args.String(1), args.Error(2)
}

// MockMessageRepository - мок для MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Send(msg domain.Message) (domain.Message, error) {
	args := m.Called(msg)
	return args.Get(0).(domain.Message), args.Error(1)
}

func (m *MockMessageRepository) Get(id string) (domain.Message, error) {
	args := m.Called(id)
	return args.Get(0).(domain.Message), args.Error(1)
}

func (m *MockMessageRepository) Update(messageID, authorID string, text *string, media *[]domain.Media) (domain.Message, error) {
	args := m.Called(messageID, authorID, text, media)
	return args.Get(0).(domain.Message), args.Error(1)
}

func (m *MockMessageRepository) Delete(messageIDs []string, hard bool, requesterID string) ([]domain.Message, error) {
	args := m.Called(messageIDs, hard, requesterID)
	return args.Get(0).([]domain.Message), args.Error(1)
}

func (m *MockMessageRepository) List(chatID string, limit int, cursor string) ([]domain.Message, string, error) {
	args := m.Called(chatID, limit, cursor)
	return args.Get(0).([]domain.Message), args.String(1), args.Error(2)
}

func (m *MockMessageRepository) MarkRead(chatID, userID, messageID string) error {
	args := m.Called(chatID, userID, messageID)
	return args.Error(0)
}

func (m *MockMessageRepository) ToggleSaved(userID, messageID string, saved bool) error {
	args := m.Called(userID, messageID, saved)
	return args.Error(0)
}

func (m *MockMessageRepository) ListSaved(userID string, limit int, cursor string) ([]domain.Message, string, error) {
	args := m.Called(userID, limit, cursor)
	return args.Get(0).([]domain.Message), args.String(1), args.Error(2)
}

func (m *MockMessageRepository) ListReadMessages(userID, chatID string, limit int) ([]domain.Message, error) {
	args := m.Called(userID, chatID, limit)
	return args.Get(0).([]domain.Message), args.Error(1)
}

func (m *MockMessageRepository) GetUnreadCount(chatID, userID string) (int64, error) {
	args := m.Called(chatID, userID)
	return args.Get(0).(int64), args.Error(1)
}

// MockKafkaProducer - мок для KafkaProducer
type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) PublishNewMessage(ctx context.Context, e domain.NewMessageEvent) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockKafkaProducer) PublishEvent(ctx context.Context, evt domain.SearchEvent) error {
	args := m.Called(ctx, evt)
	return args.Error(0)
}

// MockUserServiceClient - мок для UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) CreateUser(ctx context.Context, in *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

func (m *MockUserServiceClient) UpdateUser(ctx context.Context, in *userpb.UpdateUserRequest, opts ...grpc.CallOption) (*userpb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

func (m *MockUserServiceClient) DeleteUser(ctx context.Context, in *userpb.DeleteUserRequest, opts ...grpc.CallOption) (*userpb.DeleteUserResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.DeleteUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) AboutMeUser(ctx context.Context, in *userpb.AboutMeRequest, opts ...grpc.CallOption) (*userpb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

// createTestService создает тестовый сервис с моками
func createTestService() (*ChatService, *MockChatRepository, *MockMessageRepository, *MockKafkaProducer, *MockUserServiceClient) {
	mockChatRepo := &MockChatRepository{}
	mockMsgRepo := &MockMessageRepository{}
	mockKafka := &MockKafkaProducer{}
	mockUserClient := &MockUserServiceClient{}

	service := NewChatService(
		mockChatRepo,
		mockMsgRepo,
		mockKafka,
		mockUserClient,
	)

	return service, mockChatRepo, mockMsgRepo, mockKafka, mockUserClient
}

// createTestChat создает тестовый чат
func createTestChat(id string, kind domain.ChatKind) domain.Chat {
	return domain.Chat{
		ID:        id,
		Kind:      kind,
		MemberIDs: []string{"user1", "user2"},
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
		ReadBy:    []domain.ReadInfo{},
		SavedBy:   []domain.SavedInfo{},
		Deleted:   false,
	}
}

// ==================== ТЕСТЫ ====================

func TestChatService_CreateDirect_Success(t *testing.T) {
	// Подготовка
	service, mockChatRepo, _, _, _ := createTestService()
	ctx := context.Background()

	expectedChat := createTestChat("user1_user2", domain.ChatKindDirect)

	// Настройка моков
	mockChatRepo.On("CreateDirect", "user1", "user2", mock.Anything).Return(expectedChat, nil)

	// Выполнение
	chat, err := service.CreateDirect(ctx, "user1", "user2")

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, expectedChat, chat)
	mockChatRepo.AssertExpectations(t)
}

func TestChatService_CreateGroup_Success(t *testing.T) {
	// Подготовка
	service, mockChatRepo, _, mockKafka, _ := createTestService()
	ctx := context.Background()

	expectedChat := createTestChat("group-123", domain.ChatKindGroup)
	expectedChat.Title = "Test Group"
	expectedChat.MemberIDs = []string{"creator1", "user1", "user2"}

	// Настройка моков
	mockChatRepo.On("CreateGroup", "creator1", []string{"user1", "user2"}, "Test Group", mock.Anything).Return(expectedChat, nil)
	mockKafka.On("PublishEvent", ctx, mock.AnythingOfType("domain.SearchEvent")).Return(nil)

	// Выполнение
	chat, err := service.CreateGroup(ctx, "creator1", []string{"user1", "user2"}, "Test Group")

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, expectedChat, chat)
	mockChatRepo.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
}

// func TestChatService_SendMessage_Success(t *testing.T) {
// 	// Подготовка
// 	service, _, mockMsgRepo, mockKafka, _ := createTestService()
// 	ctx := context.Background()

// 	msg := domain.Message{
// 		ChatID:   "chat1",
// 		AuthorID: "user1",
// 		Text:     "Hello World",
// 	}

// 	expectedMsg := createTestMessage("msg-123", "chat1", "user1", "Hello World")

// 	// Настройка моков
// 	mockMsgRepo.On("Send", mock.MatchedBy(func(m domain.Message) bool {
// 		return m.ChatID == "chat1" && m.AuthorID == "user1" && m.Text == "Hello World"
// 	})).Return(expectedMsg, nil)
// 	mockKafka.On("PublishNewMessage", ctx, mock.AnythingOfType("domain.NewMessageEvent")).Return(nil)
// 	mockKafka.On("PublishEvent", ctx, mock.AnythingOfType("domain.SearchEvent")).Return(nil)

// 	// Выполнение
// 	result, err := service.SendMessage(ctx, msg)

// 	// Проверки
// 	assert.NoError(t, err)
// 	assert.Equal(t, "msg-123", result.ID)
// 	assert.Equal(t, "chat1", result.ChatID)
// 	assert.Equal(t, "Hello World", result.Text)
// 	assert.NotZero(t, result.CreatedAt)

// 	mockMsgRepo.AssertExpectations(t)
// 	mockKafka.AssertExpectations(t)
// }

func TestChatService_SendMessage_RepositoryError(t *testing.T) {
	// Подготовка
	service, _, mockMsgRepo, _, _ := createTestService()
	ctx := context.Background()

	msg := domain.Message{
		ChatID:   "chat1",
		AuthorID: "user1",
		Text:     "Hello World",
	}

	// Настройка моков - репозиторий возвращает ошибку
	mockMsgRepo.On("Send", mock.Anything).Return(domain.Message{}, errors.New("database error"))

	// Выполнение
	result, err := service.SendMessage(ctx, msg)

	// Проверки
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Equal(t, domain.Message{}, result)

	mockMsgRepo.AssertExpectations(t)
}

func TestChatService_UpdateMessage_Success(t *testing.T) {
	// Подготовка
	service, _, mockMsgRepo, _, _ := createTestService()
	ctx := context.Background()

	newText := "Updated text"
	expectedMsg := createTestMessage("msg-123", "chat1", "user1", "Updated text")

	// Настройка моков
	mockMsgRepo.On("Update", "msg-123", "user1", &newText, (*[]domain.Media)(nil)).Return(expectedMsg, nil)

	// Выполнение
	result, err := service.UpdateMessage(ctx, "msg-123", "user1", &newText, nil)

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, "Updated text", result.Text)

	mockMsgRepo.AssertExpectations(t)
}

func TestChatService_DeleteMessage_Success(t *testing.T) {
	// Подготовка
	service, _, mockMsgRepo, _, _ := createTestService()

	expectedMessages := []domain.Message{
		createTestMessage("msg-1", "chat1", "user1", "Message 1"),
		createTestMessage("msg-2", "chat1", "user1", "Message 2"),
	}

	// Настройка моков
	mockMsgRepo.On("Delete", []string{"msg-1", "msg-2"}, false, "user1").Return(expectedMessages, nil)

	// Выполнение
	result, err := service.DeleteMessage([]string{"msg-1", "msg-2"}, false, "user1")

	// Проверки
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "msg-1", result[0].ID)
	assert.Equal(t, "msg-2", result[1].ID)

	mockMsgRepo.AssertExpectations(t)
}

func TestChatService_ListMessages_Success(t *testing.T) {
	// Подготовка
	service, _, mockMsgRepo, _, _ := createTestService()
	ctx := context.Background()

	expectedMessages := []domain.Message{
		createTestMessage("msg-1", "chat1", "user1", "Message 1"),
		createTestMessage("msg-2", "chat1", "user2", "Message 2"),
		createTestMessage("msg-3", "chat1", "user1", "Message 3"),
	}

	// Настройка моков
	mockMsgRepo.On("List", "chat1", 10, "").Return(expectedMessages, "msg-3", nil)

	// Выполнение
	messages, nextCursor, err := service.ListMessages(ctx, "chat1", 10, "")

	// Проверки
	assert.NoError(t, err)
	assert.Len(t, messages, 3)
	assert.Equal(t, "msg-3", nextCursor)

	mockMsgRepo.AssertExpectations(t)
}

func TestChatService_MarkRead_Success(t *testing.T) {
	// Подготовка
	service, _, mockMsgRepo, _, _ := createTestService()
	ctx := context.Background()

	// Настройка моков
	mockMsgRepo.On("MarkRead", "chat1", "user1", "msg-123").Return(nil)

	// Выполнение
	err := service.MarkRead(ctx, "chat1", "user1", "msg-123")

	// Проверки
	assert.NoError(t, err)

	mockMsgRepo.AssertExpectations(t)
}

func TestChatService_ToggleSaved_Success(t *testing.T) {
	// Подготовка
	service, _, mockMsgRepo, _, _ := createTestService()
	ctx := context.Background()

	// Настройка моков - сохранить сообщение
	mockMsgRepo.On("ToggleSaved", "user1", "msg-123", true).Return(nil)

	// Выполнение
	err := service.ToggleSaved(ctx, "user1", "msg-123", true)

	// Проверки
	assert.NoError(t, err)

	mockMsgRepo.AssertExpectations(t)
}

func TestChatService_ListSaved_Success(t *testing.T) {
	// Подготовка
	service, _, mockMsgRepo, _, _ := createTestService()
	ctx := context.Background()

	expectedMessages := []domain.Message{
		createTestMessage("msg-1", "chat1", "user1", "Saved message 1"),
		createTestMessage("msg-2", "chat2", "user2", "Saved message 2"),
	}

	// Настройка моков
	mockMsgRepo.On("ListSaved", "user1", 10, "").Return(expectedMessages, "msg-2", nil)

	// Выполнение
	messages, nextCursor, err := service.ListSaved(ctx, "user1", 10, "")

	// Проверки
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "msg-2", nextCursor)

	mockMsgRepo.AssertExpectations(t)
}

func TestChatService_ListChats_Success(t *testing.T) {
	// Подготовка
	service, mockChatRepo, _, _, _ := createTestService()
	ctx := context.Background()

	expectedChats := []domain.Chat{
		createTestChat("chat1", domain.ChatKindDirect),
		createTestChat("chat2", domain.ChatKindGroup),
	}

	// Настройка моков
	mockChatRepo.On("List", "user1", 10, "").Return(expectedChats, "chat2", nil)

	// Выполнение
	chats, nextCursor, err := service.ListChats(ctx, "user1", 10, "")

	// Проверки
	assert.NoError(t, err)
	assert.Len(t, chats, 2)
	assert.Equal(t, "chat2", nextCursor)

	mockChatRepo.AssertExpectations(t)
}

func TestChatService_GetChat_Success(t *testing.T) {
	// Подготовка
	service, mockChatRepo, _, _, _ := createTestService()
	ctx := context.Background()

	expectedChat := createTestChat("chat1", domain.ChatKindDirect)

	// Настройка моков
	mockChatRepo.On("Get", "chat1").Return(expectedChat, nil)

	// Выполнение
	chat, err := service.GetChat(ctx, "chat1")

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, expectedChat, chat)

	mockChatRepo.AssertExpectations(t)
}

func TestChatService_UpdateGroupChat_Success(t *testing.T) {
	// Подготовка
	service, mockChatRepo, _, _, _ := createTestService()
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
	mockChatRepo.On("UpdateGroup",
		"group-123",
		mock.AnythingOfType("*string"),
		[]string{"user3"},
		[]string{"user4"},
		"creator1",
		mock.Anything,
	).Return(expectedChat, nil)

	// Выполнение
	chat, err := service.UpdateGroupChat(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, "New Title", chat.Title)

	mockChatRepo.AssertExpectations(t)
}

func TestChatService_UpdateGroupChat_NoTitleChange(t *testing.T) {
	// Подготовка
	service, mockChatRepo, _, _, _ := createTestService()
	ctx := context.Background()

	// Запрос без title
	req := &chatpb.UpdateGroupChatRequest{
		ChatId:          "group-123",
		Title:           "", // Пустая строка означает "не менять"
		AddMemberIds:    []string{"user3"},
		RemoveMemberIds: []string{},
		RequesterId:     "creator1",
	}

	expectedChat := createTestChat("group-123", domain.ChatKindGroup)

	// Настройка моков - title передается как nil
	mockChatRepo.On("UpdateGroup",
		"group-123",
		(*string)(nil), // title должен быть nil
		[]string{"user3"},
		[]string{},
		"creator1",
		mock.Anything,
	).Return(expectedChat, nil)

	// Выполнение
	chat, err := service.UpdateGroupChat(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, expectedChat, chat)

	mockChatRepo.AssertExpectations(t)
}

func TestChatService_ListReadMessages_Success(t *testing.T) {
	// Подготовка
	service, _, mockMsgRepo, _, _ := createTestService()
	ctx := context.Background()

	expectedMessages := []domain.Message{
		createTestMessage("msg-1", "chat1", "user1", "Read message 1"),
		createTestMessage("msg-2", "chat1", "user2", "Read message 2"),
	}

	// Настройка моков
	mockMsgRepo.On("ListReadMessages", "user1", "chat1", 10).Return(expectedMessages, nil)

	// Выполнение
	messages, err := service.ListReadMessages(ctx, "user1", "chat1", 10)

	// Проверки
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "Read message 1", messages[0].Text)

	mockMsgRepo.AssertExpectations(t)
}

func TestChatService_ListReadMessages_WithEmptyChatID(t *testing.T) {
	// Подготовка
	service, _, mockMsgRepo, _, _ := createTestService()
	ctx := context.Background()

	expectedMessages := []domain.Message{
		createTestMessage("msg-1", "chat1", "user1", "Message 1"),
		createTestMessage("msg-2", "chat2", "user1", "Message 2"),
	}

	// Настройка моков - chatID пустая, получаем все прочитанные сообщения пользователя
	mockMsgRepo.On("ListReadMessages", "user1", "", 10).Return(expectedMessages, nil)

	// Выполнение
	messages, err := service.ListReadMessages(ctx, "user1", "", 10)

	// Проверки
	assert.NoError(t, err)
	assert.Len(t, messages, 2)

	mockMsgRepo.AssertExpectations(t)
}

func TestChatService_DeleteMessage_HardDelete(t *testing.T) {
	// Подготовка
	service, _, mockMsgRepo, _, _ := createTestService()

	expectedMessages := []domain.Message{
		createTestMessage("msg-1", "chat1", "user1", "Message 1"),
	}

	// Настройка моков - жесткое удаление
	mockMsgRepo.On("Delete", []string{"msg-1"}, true, "user1").Return(expectedMessages, nil)

	// Выполнение
	result, err := service.DeleteMessage([]string{"msg-1"}, true, "user1")

	// Проверки
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	mockMsgRepo.AssertExpectations(t)
}
