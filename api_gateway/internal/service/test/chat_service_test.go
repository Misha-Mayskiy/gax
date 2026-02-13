// internal/service/test/chat_service_test.go
package test

import (
	"context"
	"testing"

	"api_gateway/internal/service"
	pb "api_gateway/pkg/api/chat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockChatClient - мок для ChatServiceClient
type MockChatClient struct {
	mock.Mock
}

func (m *MockChatClient) CreateDirectChat(ctx context.Context, in *pb.CreateDirectChatRequest, opts ...grpc.CallOption) (*pb.ChatResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ChatResponse), args.Error(1)
}
func (m *MockChatClient) CreateGroupChat(ctx context.Context, in *pb.CreateGroupChatRequest, opts ...grpc.CallOption) (*pb.ChatResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ChatResponse), args.Error(1)
}

func (m *MockChatClient) UpdateGroupChat(ctx context.Context, in *pb.UpdateGroupChatRequest, opts ...grpc.CallOption) (*pb.ChatResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ChatResponse), args.Error(1)
}

func (m *MockChatClient) GetChat(ctx context.Context, in *pb.GetChatRequest, opts ...grpc.CallOption) (*pb.ChatResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ChatResponse), args.Error(1)
}

func (m *MockChatClient) ListChats(ctx context.Context, in *pb.ListChatsRequest, opts ...grpc.CallOption) (*pb.ListChatsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ListChatsResponse), args.Error(1)
}

func (m *MockChatClient) SendMessage(ctx context.Context, in *pb.SendMessageRequest, opts ...grpc.CallOption) (*pb.MessageResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.MessageResponse), args.Error(1)
}

func (m *MockChatClient) UpdateMessage(ctx context.Context, in *pb.UpdateMessageRequest, opts ...grpc.CallOption) (*pb.MessageResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.MessageResponse), args.Error(1)
}

func (m *MockChatClient) DeleteMessage(ctx context.Context, in *pb.DeleteMessageRequest, opts ...grpc.CallOption) (*pb.DeleteMessageResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.DeleteMessageResponse), args.Error(1)
}

func (m *MockChatClient) ListMessages(ctx context.Context, in *pb.ListMessagesRequest, opts ...grpc.CallOption) (*pb.ListMessagesResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ListMessagesResponse), args.Error(1)
}

func (m *MockChatClient) MarkRead(ctx context.Context, in *pb.MarkReadRequest, opts ...grpc.CallOption) (*pb.MarkReadResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.MarkReadResponse), args.Error(1)
}

func (m *MockChatClient) ToggleSaved(ctx context.Context, in *pb.ToggleSavedRequest, opts ...grpc.CallOption) (*pb.ToggleSavedResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ToggleSavedResponse), args.Error(1)
}

func (m *MockChatClient) ListSaved(ctx context.Context, in *pb.ListSavedRequest, opts ...grpc.CallOption) (*pb.ListSavedResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ListSavedResponse), args.Error(1)
}

func (m *MockChatClient) ListReadMessages(ctx context.Context, in *pb.ListReadMessagesRequest, opts ...grpc.CallOption) (*pb.ListReadMessagesResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ListReadMessagesResponse), args.Error(1)
}

// Тесты для ChatService

// func TestChatService(t *testing.T) {
// 	mockClient := new(MockChatClient)
// 	chatService := service.NewTestChatService(mockClient)

// 	t.Run("CreateDirectChat успешно", func(t *testing.T) {
// 		req := &pb.CreateDirectChatRequest{
// 			UserId: "user-123",
// 			PeerId: "user-456",
// 		}

// 		expectedResponse := &pb.ChatResponse{
// 			Chat: &pb.Chat{
// 				Id:        "chat-123",
// 				Kind:      "direct",
// 				MemberIds: []string{"user-123", "user-456"},
// 				CreatedBy: "user-123",
// 				CreatedAt: "2024-01-01T12:00:00Z",
// 			},
// 		}

// 		mockClient.On("CreateDirectChat", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

// 		resp, err := chatService.CreateDirectChat(context.Background(), req)

// 		assert.NoError(t, err)
// 		assert.Equal(t, expectedResponse, resp)
// 		mockClient.AssertExpectations(t)
// 	})

// 	t.Run("CreateDirectChat с ошибкой", func(t *testing.T) {
// 		req := &pb.CreateDirectChatRequest{
// 			UserId: "user-123",
// 			PeerId: "user-456",
// 		}

// 		mockClient.On("CreateDirectChat", mock.Anything, req, mock.Anything).
// 			Return(nil, assert.AnError)

// 		resp, err := chatService.CreateDirectChat(context.Background(), req)

//			assert.Error(t, err)
//			assert.Nil(t, resp)
//			mockClient.AssertExpectations(t)
//		})
//	}
func TestChatService_CreateGroupChat(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("CreateGroupChat успешно", func(t *testing.T) {
		req := &pb.CreateGroupChatRequest{
			UserId:    "user-123",
			MemberIds: []string{"user-123", "user-456", "user-789"},
			Title:     "Музыкальная группа",
		}

		expectedChat := &pb.Chat{
			Id:        "chat-group-123",
			Kind:      "group",
			MemberIds: []string{"user-123", "user-456", "user-789"},
			Title:     "Музыкальная группа",
			CreatedBy: "user-123",
			CreatedAt: "2024-01-01T12:00:00Z",
		}

		expectedResponse := &pb.ChatResponse{
			Chat: expectedChat,
		}

		mockClient.On("CreateGroupChat", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.CreateGroupChat(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_GetChat(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("GetChat успешно", func(t *testing.T) {
		req := &pb.GetChatRequest{
			ChatId: "chat-123",
		}

		expectedChat := &pb.Chat{
			Id:        "chat-123",
			Kind:      "direct",
			MemberIds: []string{"user-123", "user-456"},
			CreatedAt: "2024-01-01T12:00:00Z",
			CreatedBy: "user-123",
		}

		expectedResponse := &pb.ChatResponse{
			Chat: expectedChat,
		}

		mockClient.On("GetChat", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.GetChat(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_SendMessage(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("SendMessage успешно", func(t *testing.T) {
		req := &pb.SendMessageRequest{
			ChatId:   "chat-123",
			AuthorId: "user-123",
			Text:     "Привет! Как дела?",
		}

		expectedMessage := &pb.Message{
			Id:        "msg-123",
			ChatId:    "chat-123",
			AuthorId:  "user-123",
			Text:      "Привет! Как дела?",
			CreatedAt: "2024-01-01T12:00:00Z",
		}

		expectedResponse := &pb.MessageResponse{
			Message: expectedMessage,
		}

		mockClient.On("SendMessage", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.SendMessage(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})

	t.Run("SendMessage с медиа", func(t *testing.T) {
		req := &pb.SendMessageRequest{
			ChatId:   "chat-123",
			AuthorId: "user-123",
			Text:     "Вот файл",
			Media: []*pb.Media{
				{
					Id:        "media-123",
					Type:      "image",
					Url:       "https://example.com/image.jpg",
					Mime:      "image/jpeg",
					SizeBytes: 102400,
				},
			},
		}

		expectedMessage := &pb.Message{
			Id:       "msg-456",
			ChatId:   "chat-123",
			AuthorId: "user-123",
			Text:     "Вот файл",
			Media: []*pb.Media{
				{
					Id:        "media-123",
					Type:      "image",
					Url:       "https://example.com/image.jpg",
					Mime:      "image/jpeg",
					SizeBytes: 102400,
				},
			},
			CreatedAt: "2024-01-01T12:00:00Z",
		}

		expectedResponse := &pb.MessageResponse{
			Message: expectedMessage,
		}

		mockClient.On("SendMessage", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.SendMessage(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_UpdateMessage(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("UpdateMessage успешно", func(t *testing.T) {
		req := &pb.UpdateMessageRequest{
			MessageId: "msg-123",
			AuthorId:  "user-123",
			Text:      "Обновленный текст",
		}

		expectedMessage := &pb.Message{
			Id:        "msg-123",
			ChatId:    "chat-123",
			AuthorId:  "user-123",
			Text:      "Обновленный текст",
			CreatedAt: "2024-01-01T12:00:00Z",
			UpdatedAt: "2024-01-01T12:05:00Z",
		}

		expectedResponse := &pb.MessageResponse{
			Message: expectedMessage,
		}

		mockClient.On("UpdateMessage", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.UpdateMessage(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_DeleteMessage(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("DeleteMessage успешно", func(t *testing.T) {
		req := &pb.DeleteMessageRequest{
			MessageIds:  []string{"msg-123", "msg-456"},
			RequesterId: "user-123",
			HardDelete:  false,
		}

		expectedResponse := &pb.DeleteMessageResponse{
			Success: true,
			Message: "Messages deleted successfully",
		}

		mockClient.On("DeleteMessage", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.DeleteMessage(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_ListMessages(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("ListMessages успешно", func(t *testing.T) {
		req := &pb.ListMessagesRequest{
			ChatId: "chat-123",
			Limit:  50,
			Cursor: "",
		}

		expectedMessages := []*pb.Message{
			{
				Id:        "msg-1",
				ChatId:    "chat-123",
				AuthorId:  "user-123",
				Text:      "Первое сообщение",
				CreatedAt: "2024-01-01T12:00:00Z",
			},
			{
				Id:        "msg-2",
				ChatId:    "chat-123",
				AuthorId:  "user-456",
				Text:      "Второе сообщение",
				CreatedAt: "2024-01-01T12:01:00Z",
			},
		}

		expectedResponse := &pb.ListMessagesResponse{
			Messages:   expectedMessages,
			NextCursor: "cursor-next",
		}

		mockClient.On("ListMessages", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.ListMessages(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		assert.Len(t, resp.Messages, 2)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_MarkRead(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("MarkRead успешно", func(t *testing.T) {
		req := &pb.MarkReadRequest{
			ChatId:    "chat-123",
			UserId:    "user-456",
			MessageId: "msg-123",
		}

		expectedResponse := &pb.MarkReadResponse{
			Success: true,
		}

		mockClient.On("MarkRead", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.MarkRead(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		assert.True(t, resp.Success)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_ToggleSaved(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("ToggleSaved успешно - сохранить", func(t *testing.T) {
		req := &pb.ToggleSavedRequest{
			UserId:    "user-456",
			MessageId: "msg-123",
			Saved:     true,
		}

		expectedResponse := &pb.ToggleSavedResponse{
			Success: true,
		}

		mockClient.On("ToggleSaved", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.ToggleSaved(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		assert.True(t, resp.Success)
		mockClient.AssertExpectations(t)
	})

	t.Run("ToggleSaved успешно - убрать из сохраненных", func(t *testing.T) {
		req := &pb.ToggleSavedRequest{
			UserId:    "user-456",
			MessageId: "msg-123",
			Saved:     false,
		}

		expectedResponse := &pb.ToggleSavedResponse{
			Success: true,
		}

		mockClient.On("ToggleSaved", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.ToggleSaved(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		assert.True(t, resp.Success)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_ListChats(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("ListChats успешно", func(t *testing.T) {
		req := &pb.ListChatsRequest{
			UserId: "user-123",
			Limit:  20,
			Cursor: "",
		}

		expectedChats := []*pb.Chat{
			{
				Id:        "chat-123",
				Kind:      "direct",
				MemberIds: []string{"user-123", "user-456"},
				CreatedAt: "2024-01-01T12:00:00Z",
				CreatedBy: "user-123",
			},
			{
				Id:        "chat-group-123",
				Kind:      "group",
				MemberIds: []string{"user-123", "user-456", "user-789"},
				Title:     "Музыкальная группа",
				CreatedAt: "2024-01-01T11:00:00Z",
				CreatedBy: "user-123",
			},
		}

		expectedResponse := &pb.ListChatsResponse{
			Chats:      expectedChats,
			NextCursor: "cursor-next",
		}

		mockClient.On("ListChats", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.ListChats(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		assert.Len(t, resp.Chats, 2)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_ListSaved(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("ListSaved успешно", func(t *testing.T) {
		req := &pb.ListSavedRequest{
			UserId: "user-123",
			Limit:  10,
			Cursor: "",
		}

		expectedMessages := []*pb.Message{
			{
				Id:        "msg-1",
				ChatId:    "chat-123",
				AuthorId:  "user-456",
				Text:      "Важное сообщение 1",
				CreatedAt: "2024-01-01T10:00:00Z",
			},
			{
				Id:        "msg-2",
				ChatId:    "chat-group-123",
				AuthorId:  "user-789",
				Text:      "Важное сообщение 2",
				CreatedAt: "2024-01-01T09:00:00Z",
			},
		}

		expectedResponse := &pb.ListSavedResponse{
			Messages:   expectedMessages,
			NextCursor: "cursor-next",
		}

		mockClient.On("ListSaved", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.ListSaved(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		assert.Len(t, resp.Messages, 2)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_ListReadMessages(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("ListReadMessages успешно", func(t *testing.T) {
		req := &pb.ListReadMessagesRequest{
			UserId: "user-123",
			ChatId: "chat-456",
			Limit:  10,
		}

		expectedMessages := []*pb.Message{
			{
				Id:        "msg-1",
				ChatId:    "chat-456",
				AuthorId:  "user-789",
				Text:      "Прочитанное сообщение 1",
				CreatedAt: "2024-01-01T08:00:00Z",
			},
			{
				Id:        "msg-2",
				ChatId:    "chat-456",
				AuthorId:  "user-789",
				Text:      "Прочитанное сообщение 2",
				CreatedAt: "2024-01-01T09:00:00Z",
			},
		}

		expectedResponse := &pb.ListReadMessagesResponse{
			Messages: expectedMessages,
		}

		mockClient.On("ListReadMessages", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.ListReadMessages(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		assert.Len(t, resp.Messages, 2)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_UpdateGroupChat(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("UpdateGroupChat успешно", func(t *testing.T) {
		req := &pb.UpdateGroupChatRequest{
			ChatId:          "chat-group-123",
			Title:           "Новое название группы",
			AddMemberIds:    []string{"user-999"},
			RemoveMemberIds: []string{"user-456"},
			RequesterId:     "user-123",
		}

		expectedChat := &pb.Chat{
			Id:        "chat-group-123",
			Kind:      "group",
			MemberIds: []string{"user-123", "user-789", "user-999"},
			Title:     "Новое название группы",
			CreatedBy: "user-123",
			CreatedAt: "2024-01-01T10:00:00Z",
		}

		expectedResponse := &pb.ChatResponse{
			Chat: expectedChat,
		}

		mockClient.On("UpdateGroupChat", mock.Anything, req, mock.Anything).Return(expectedResponse, nil)

		resp, err := chatService.UpdateGroupChat(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		mockClient.AssertExpectations(t)
	})
}

func TestChatService_ErrorHandling(t *testing.T) {
	mockClient := new(MockChatClient)
	chatService := service.NewTestChatService(mockClient)

	t.Run("SendMessage возвращает ошибку", func(t *testing.T) {
		req := &pb.SendMessageRequest{
			ChatId:   "chat-123",
			AuthorId: "user-123",
			Text:     "Тестовое сообщение",
		}

		mockClient.On("SendMessage", mock.Anything, req, mock.Anything).
			Return((*pb.MessageResponse)(nil), assert.AnError)

		resp, err := chatService.SendMessage(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockClient.AssertExpectations(t)
	})

	t.Run("GetChat возвращает ошибку", func(t *testing.T) {
		req := &pb.GetChatRequest{
			ChatId: "nonexistent-chat",
		}

		mockClient.On("GetChat", mock.Anything, req, mock.Anything).
			Return((*pb.ChatResponse)(nil), assert.AnError)

		resp, err := chatService.GetChat(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockClient.AssertExpectations(t)
	})
}
