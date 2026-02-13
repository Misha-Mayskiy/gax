package mongo

import (
	"context"
	"main/internal/domain"
	"main/internal/logger"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	userpb "main/pkg/api_user_service"
)

// ==================== МОКИ ====================

// MockCollection - полный мок для коллекции
type MockCollection struct {
	mock.Mock
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *MockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func (m *MockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockCollection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

// MockSingleResult - мок для mongo.SingleResult
type MockSingleResult struct {
	mock.Mock
}

func (m *MockSingleResult) Decode(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockSingleResult) Err() error {
	args := m.Called()
	return args.Error(0)
}

// MockCursor - мок для mongo.Cursor
type MockCursor struct {
	mock.Mock
}

func (m *MockCursor) All(ctx context.Context, results interface{}) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

func (m *MockCursor) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockUserServiceClient - правильная реализация интерфейса user.UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
	users map[string]*domain.User
}

// NewMockUserServiceClient создает мок с тестовыми пользователями
func NewMockUserServiceClient(users map[string]*domain.User) *MockUserServiceClient {
	return &MockUserServiceClient{
		users: users,
	}
}

// GetUserInfo - вспомогательный метод для тестов (не из интерфейса)
func (m *MockUserServiceClient) GetUserInfo(userID string) (*domain.User, error) {
	if user, exists := m.users[userID]; exists {
		return user, nil
	}
	return nil, nil
}

// AboutMeUser - правильная сигнатура из интерфейса
func (m *MockUserServiceClient) AboutMeUser(ctx context.Context, in *userpb.AboutMeRequest, opts ...grpc.CallOption) (*userpb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

// CreateUser - правильная сигнатура (заглушка)
func (m *MockUserServiceClient) CreateUser(ctx context.Context, in *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

// UpdateUser - правильная сигнатура (заглушка)
func (m *MockUserServiceClient) UpdateUser(ctx context.Context, in *userpb.UpdateUserRequest, opts ...grpc.CallOption) (*userpb.UserResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

// DeleteUser - правильная сигнатура (заглушка)
func (m *MockUserServiceClient) DeleteUser(ctx context.Context, in *userpb.DeleteUserRequest, opts ...grpc.CallOption) (*userpb.DeleteUserResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.DeleteUserResponse), args.Error(1)
}

// MockLogger - простая структура, совместимая с zerolog.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug() *MockLogger {
	return m
}

func (m *MockLogger) Info() *MockLogger {
	return m
}

func (m *MockLogger) Warn() *MockLogger {
	return m
}

func (m *MockLogger) Error() *MockLogger {
	return m
}

func (m *MockLogger) Msg(msg string) *MockLogger {
	m.Called(msg)
	return m
}

func (m *MockLogger) Msgf(format string, args ...interface{}) *MockLogger {
	m.Called(append([]interface{}{format}, args...)...)
	return m
}

func (m *MockLogger) Str(key, value string) *MockLogger {
	m.Called(key, value)
	return m
}

func (m *MockLogger) Err(err error) *MockLogger {
	m.Called(err)
	return m
}

// Остальные методы zerolog.Logger (заглушки)
func (m *MockLogger) With() *MockLogger {
	return m
}

func (m *MockLogger) Log() *MockLogger {
	return m
}

func (m *MockLogger) Print(v ...interface{}) {
	m.Called(v...)
}

func (m *MockLogger) Printf(format string, v ...interface{}) {
	m.Called(append([]interface{}{format}, v...)...)
}

func (m *MockLogger) Fatal() *MockLogger {
	return m
}

func (m *MockLogger) Panic() *MockLogger {
	return m
}

func (m *MockLogger) Trace() *MockLogger {
	return m
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

// createTestChatRepo создает тестовый ChatRepo с минимальными зависимостями
func createTestChatRepo() (*ChatRepo, *MockCollection) {
	mockCol := &MockCollection{}
	// mockLogger := &MockLogger{}
	// Создаем ChatRepo напрямую, без NewTestChatRepo
	repo := &ChatRepo{
		col: mockCol,
		log: logger.GetLogger(),
	}
	return repo, mockCol
}

// createTestMessageRepo создает тестовый MessageRepo
func createTestMessageRepo() (*MessageRepo, *MockCollection) {
	mockCol := &MockCollection{}
	repo := &MessageRepo{
		col: mockCol,
	}
	return repo, mockCol
}

// createTestUserResponse создает тестовый UserResponse
// createTestUserResponse создает тестовый UserResponse
func createTestUserResponse(userID string) *userpb.UserResponse {
	return &userpb.UserResponse{
		Uuid:      userID,
		Email:     userID + "@example.com",
		UserName:  "user_" + userID,
		Avatar:    "https://example.com/avatar/" + userID,
		AboutMe:   "About " + userID,
		Friends:   []string{},
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
}

// ==================== ТЕСТЫ ====================

func TestChatRepository_CreateDirect_Success(t *testing.T) {
	// Подготовка
	repo, mockCol := createTestChatRepo()

	// Создаем мок пользовательского клиента
	userClient := &MockUserServiceClient{}

	// Настройка моков для AboutMeUser (вызывается через GetUserInfo в репозитории)
	userClient.On("AboutMeUser", mock.Anything, &userpb.AboutMeRequest{Uuid: "user1"}, mock.Anything).
		Return(createTestUserResponse("user1"), nil)
	userClient.On("AboutMeUser", mock.Anything, &userpb.AboutMeRequest{Uuid: "user2"}, mock.Anything).
		Return(createTestUserResponse("user2"), nil)

	// Настройка моков коллекции
	mockCol.On("CountDocuments", mock.Anything,
		bson.M{"id": bson.M{"$in": []string{"user1_user2", "user2_user1"}}},
		mock.Anything).Return(int64(0), nil)

	mockCol.On("InsertOne", mock.Anything, mock.MatchedBy(func(doc interface{}) bool {
		chat, ok := doc.(domain.Chat)
		return ok && chat.ID == "user1_user2"
	})).Return(&mongo.InsertOneResult{InsertedID: "user1_user2"}, nil)

	// Выполнение
	chat, err := repo.CreateDirect("user1", "user2", userClient)

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, "user1_user2", chat.ID)
	assert.Equal(t, domain.ChatKindDirect, chat.Kind)
	assert.Contains(t, chat.MemberIDs, "user1")
	assert.Contains(t, chat.MemberIDs, "user2")
	assert.Equal(t, "user1", chat.CreatedBy)
	assert.NotZero(t, chat.CreatedAt)

	mockCol.AssertExpectations(t)
	userClient.AssertExpectations(t)
}

func TestChatRepository_CreateDirect_ChatExists(t *testing.T) {
	// Подготовка
	repo, mockCol := createTestChatRepo()

	userClient := &MockUserServiceClient{}
	userClient.On("AboutMeUser", mock.Anything, &userpb.AboutMeRequest{Uuid: "user1"}, mock.Anything).
		Return(createTestUserResponse("user1"), nil)
	userClient.On("AboutMeUser", mock.Anything, &userpb.AboutMeRequest{Uuid: "user2"}, mock.Anything).
		Return(createTestUserResponse("user2"), nil)

	// Настройка моков - чат уже существует
	mockCol.On("CountDocuments", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	// Выполнение
	chat, err := repo.CreateDirect("user1", "user2", userClient)

	// Проверки
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "чат между пользователями уже существует")
	assert.Equal(t, domain.Chat{}, chat)

	mockCol.AssertExpectations(t)
}

// func TestChatRepository_Get_Success(t *testing.T) {
// 	// Подготовка
// 	repo, mockCol := createTestChatRepo()

// 	// Ожидаемый чат
// 	expectedChat := domain.Chat{
// 		ID:        "chat1",
// 		Kind:      domain.ChatKindDirect,
// 		MemberIDs: []string{"user1", "user2"},
// 		CreatedBy: "user1",
// 		CreatedAt: time.Now().Unix(),
// 	}

// 	// Настройка моков
// 	mockResult := &MockSingleResult{}
// 	mockResult.On("Decode", mock.Anything).Run(func(args mock.Arguments) {
// 		arg := args.Get(0).(*domain.Chat)
// 		*arg = expectedChat
// 	}).Return(nil)

// 	mockCol.On("FindOne", mock.Anything, bson.M{"id": "chat1"}).Return(mockResult)

// 	// Выполнение
// 	chat, err := repo.Get("chat1")

// 	// Проверки
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedChat, chat)

// 	mockCol.AssertExpectations(t)
// 	mockResult.AssertExpectations(t)
// }

func TestMessageRepository_Send_Success(t *testing.T) {
	// Подготовка
	repo, mockCol := createTestMessageRepo()

	// Настройка моков
	mockCol.On("InsertOne", mock.Anything, mock.Anything).Return(
		&mongo.InsertOneResult{InsertedID: "msg-123"}, nil)

	// Тестовое сообщение
	msg := domain.Message{
		ChatID:   "chat1",
		AuthorID: "user1",
		Text:     "Hello World",
	}

	// Выполнение
	result, err := repo.Send(msg)

	// Проверки
	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "chat1", result.ChatID)
	assert.Equal(t, "user1", result.AuthorID)
	assert.Equal(t, "Hello World", result.Text)
	assert.NotZero(t, result.CreatedAt)
	assert.Empty(t, result.ReadBy)
	assert.Empty(t, result.SavedBy)
	assert.False(t, result.Deleted)

	mockCol.AssertExpectations(t)
}

// func TestMessageRepository_Get_Success(t *testing.T) {
// 	// Подготовка
// 	repo, mockCol := createTestMessageRepo()

// 	// Ожидаемое сообщение
// 	expectedMsg := domain.Message{
// 		ID:        "msg1",
// 		ChatID:    "chat1",
// 		AuthorID:  "user1",
// 		Text:      "Test message",
// 		CreatedAt: time.Now().Unix(),
// 	}

// 	// Настройка моков
// 	mockResult := &MockSingleResult{}
// 	mockResult.On("Decode", mock.Anything).Run(func(args mock.Arguments) {
// 		arg := args.Get(0).(*domain.Message)
// 		*arg = expectedMsg
// 	}).Return(nil)

// 	mockCol.On("FindOne", mock.Anything, bson.M{"id": "msg1"}).Return(mockResult)

// 	// Выполнение
// 	msg, err := repo.Get("msg1")

// 	// Проверки
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedMsg, msg)

// 	mockCol.AssertExpectations(t)
// 	mockResult.AssertExpectations(t)
// }

func TestMessageRepository_MarkRead_Success(t *testing.T) {
	// Подготовка
	repo, mockCol := createTestMessageRepo()

	// Настройка моков
	mockCol.On("UpdateOne", mock.Anything,
		bson.M{
			"id":              "msg1",
			"chat_id":         "chat1",
			"read_by.user_id": bson.M{"$ne": "user1"},
		},
		mock.Anything, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1}, nil)

	// Выполнение
	err := repo.MarkRead("chat1", "user1", "msg1")

	// Проверки
	assert.NoError(t, err)

	mockCol.AssertExpectations(t)
}

func TestMessageRepository_GetUnreadCount_Success(t *testing.T) {
	// Подготовка
	repo, mockCol := createTestMessageRepo()

	// Настройка моков
	mockCol.On("CountDocuments", mock.Anything,
		bson.M{
			"chat_id":         "chat1",
			"read_by.user_id": bson.M{"$ne": "user1"},
			"deleted":         false,
		}, mock.Anything).Return(int64(5), nil)

	// Выполнение
	count, err := repo.GetUnreadCount("chat1", "user1")

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)

	mockCol.AssertExpectations(t)
}
