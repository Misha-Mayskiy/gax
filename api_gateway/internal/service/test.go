// api_gateway/internal/service/service.go
package service

import (
	pb "api_gateway/pkg/api"
	pbAuth "api_gateway/pkg/api/auth"
	pbChat "api_gateway/pkg/api/chat"
	pbRoom "api_gateway/pkg/api/room"
)

// ExportAuthServiceForTest экспортирует приватную структуру для тестирования
type ExportAuthServiceForTest struct {
	AuthClient pbAuth.AuthServiceClient
}

// ExportChatServiceForTest экспортирует приватную структуру для тестирования
type ExportChatServiceForTest struct {
	ChatClient pbChat.ChatServiceClient
}

// ExportRoomServiceForTest экспортирует приватную структуру для тестирования
type ExportRoomServiceForTest struct {
	RoomClient pbRoom.RoomServiceClient
}

// Экспортируем функции для создания сервисов с моками
func NewAuthServiceWithClient(client pbAuth.AuthServiceClient) AuthService {
	return &authService{authClient: client}
}

func NewChatServiceWithClient(client pbChat.ChatServiceClient) ChatService {
	return &chatService{chatClient: client}
}

func NewRoomServiceWithClient(client pbRoom.RoomServiceClient) RoomService {
	return &roomService{roomclient: client}
}

// Конструкторы для тестов

// NewTestAuthService создает AuthService для тестов
func NewTestAuthService(client pbAuth.AuthServiceClient) AuthService {
	return &authService{authClient: client}
}

// NewTestChatService создает ChatService для тестов
func NewTestChatService(client pbChat.ChatServiceClient) ChatService {
	return &chatService{chatClient: client}
}

// NewTestRoomService создает RoomService для тестов
func NewTestRoomService(client pbRoom.RoomServiceClient) RoomService {
	return &roomService{roomclient: client}
}

// NewTestUserService создает UserService для тестов
func NewTestUserService(client pb.UserServiceClient) *UserService {
	return &UserService{client: client}
}
