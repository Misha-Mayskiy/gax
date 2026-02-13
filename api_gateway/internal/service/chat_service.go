package service

import (
	"context"
	"fmt"

	"api_gateway/client"
	pb "api_gateway/pkg/api/chat"
)

type ChatService interface {
	CreateDirectChat(ctx context.Context, req *pb.CreateDirectChatRequest) (*pb.ChatResponse, error)
	CreateGroupChat(ctx context.Context, req *pb.CreateGroupChatRequest) (*pb.ChatResponse, error)
	UpdateGroupChat(ctx context.Context, req *pb.UpdateGroupChatRequest) (*pb.ChatResponse, error)
	GetChat(ctx context.Context, req *pb.GetChatRequest) (*pb.ChatResponse, error)
	ListChats(ctx context.Context, req *pb.ListChatsRequest) (*pb.ListChatsResponse, error)
	SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.MessageResponse, error)
	UpdateMessage(ctx context.Context, req *pb.UpdateMessageRequest) (*pb.MessageResponse, error)
	DeleteMessage(ctx context.Context, req *pb.DeleteMessageRequest) (*pb.DeleteMessageResponse, error)
	ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error)
	MarkRead(ctx context.Context, req *pb.MarkReadRequest) (*pb.MarkReadResponse, error)
	ToggleSaved(ctx context.Context, req *pb.ToggleSavedRequest) (*pb.ToggleSavedResponse, error)
	ListSaved(ctx context.Context, req *pb.ListSavedRequest) (*pb.ListSavedResponse, error)
	ListReadMessages(ctx context.Context, req *pb.ListReadMessagesRequest) (*pb.ListReadMessagesResponse, error)
}

type chatService struct {
	chatClient pb.ChatServiceClient
}

func NewChatService(addr string) ChatService {
	chatClient, err := client.NewChatClient(addr)
	if err != nil {
		fmt.Printf("Warning: failed to create chat client: %v\n", err)
	}

	return &chatService{
		chatClient: chatClient,
	}
}

func (s *chatService) CreateDirectChat(ctx context.Context, req *pb.CreateDirectChatRequest) (*pb.ChatResponse, error) {
	return s.chatClient.CreateDirectChat(ctx, req)
}

func (s *chatService) CreateGroupChat(ctx context.Context, req *pb.CreateGroupChatRequest) (*pb.ChatResponse, error) {
	return s.chatClient.CreateGroupChat(ctx, req)
}

func (s *chatService) UpdateGroupChat(ctx context.Context, req *pb.UpdateGroupChatRequest) (*pb.ChatResponse, error) {
	return s.chatClient.UpdateGroupChat(ctx, req)
}

func (s *chatService) GetChat(ctx context.Context, req *pb.GetChatRequest) (*pb.ChatResponse, error) {
	return s.chatClient.GetChat(ctx, req)
}

func (s *chatService) ListChats(ctx context.Context, req *pb.ListChatsRequest) (*pb.ListChatsResponse, error) {
	return s.chatClient.ListChats(ctx, req)
}

func (s *chatService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.MessageResponse, error) {
	return s.chatClient.SendMessage(ctx, req)
}

func (s *chatService) UpdateMessage(ctx context.Context, req *pb.UpdateMessageRequest) (*pb.MessageResponse, error) {
	return s.chatClient.UpdateMessage(ctx, req)
}

func (s *chatService) DeleteMessage(ctx context.Context, req *pb.DeleteMessageRequest) (*pb.DeleteMessageResponse, error) {
	return s.chatClient.DeleteMessage(ctx, req)
}

func (s *chatService) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	return s.chatClient.ListMessages(ctx, req)
}

func (s *chatService) MarkRead(ctx context.Context, req *pb.MarkReadRequest) (*pb.MarkReadResponse, error) {
	return s.chatClient.MarkRead(ctx, req)
}

func (s *chatService) ToggleSaved(ctx context.Context, req *pb.ToggleSavedRequest) (*pb.ToggleSavedResponse, error) {
	return s.chatClient.ToggleSaved(ctx, req)
}

func (s *chatService) ListSaved(ctx context.Context, req *pb.ListSavedRequest) (*pb.ListSavedResponse, error) {
	return s.chatClient.ListSaved(ctx, req)
}

func (s *chatService) ListReadMessages(ctx context.Context, req *pb.ListReadMessagesRequest) (*pb.ListReadMessagesResponse, error) {
	return s.chatClient.ListReadMessages(ctx, req)
}
