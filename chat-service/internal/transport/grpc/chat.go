package grpc

import (
	"context"
	"strconv"

	"main/internal/domain"
	chatpb "main/pkg/api"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ChatServiceInterface - интерфейс для сервиса чатов
type ChatServiceInterface interface {
	CreateDirect(ctx context.Context, userID, peerID string) (domain.Chat, error)
	CreateGroup(ctx context.Context, creatorID string, members []string, title string) (domain.Chat, error)
	SendMessage(ctx context.Context, msg domain.Message) (domain.Message, error)
	UpdateMessage(ctx context.Context, messageID, authorID string, text *string, media *[]domain.Media) (domain.Message, error)
	DeleteMessage(messageIDs []string, hard bool, requesterID string) ([]domain.Message, error)
	ListMessages(ctx context.Context, chatID string, limit int, cursor string) ([]domain.Message, string, error)
	MarkRead(ctx context.Context, chatID, userID, messageID string) error
	ToggleSaved(ctx context.Context, userID, messageID string, saved bool) error
	ListSaved(ctx context.Context, userID string, limit int, cursor string) ([]domain.Message, string, error)
	ListChats(ctx context.Context, userID string, limit int, cursor string) ([]domain.Chat, string, error)
	GetChat(ctx context.Context, chatID string) (domain.Chat, error)
	UpdateGroupChat(ctx context.Context, req *chatpb.UpdateGroupChatRequest) (domain.Chat, error)
	ListReadMessages(ctx context.Context, userID, chatID string, limit int) ([]domain.Message, error)
}

type ChatServer struct {
	chatpb.UnimplementedChatServiceServer
	svc ChatServiceInterface
}

func NewChatServer(svc ChatServiceInterface) *ChatServer {
	return &ChatServer{svc: svc}
}

// --- Messages ---

func (s *ChatServer) SendMessage(ctx context.Context, req *chatpb.SendMessageRequest) (*chatpb.MessageResponse, error) {
	msg, err := s.svc.SendMessage(ctx, domain.Message{
		ChatID:   req.ChatId,
		AuthorID: req.AuthorId,
		Text:     req.Text,
		// media из req.Media можно промапить позже
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
	}

	return &chatpb.MessageResponse{
		Message: &chatpb.Message{
			Id:        msg.ID,
			ChatId:    msg.ChatID,
			AuthorId:  msg.AuthorID,
			Text:      msg.Text,
			CreatedAt: strconv.FormatInt(msg.CreatedAt, 10),
			UpdatedAt: strconv.FormatInt(msg.UpdatedAt, 10),
			Deleted:   msg.Deleted,
		},
	}, nil
}

func (s *ChatServer) UpdateMessage(ctx context.Context, req *chatpb.UpdateMessageRequest) (*chatpb.MessageResponse, error) {
	text := req.Text
	msg, err := s.svc.UpdateMessage(ctx, req.MessageId, req.AuthorId, &text, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update message: %v", err)
	}
	return &chatpb.MessageResponse{
		Message: &chatpb.Message{
			Id:        msg.ID,
			ChatId:    msg.ChatID,
			AuthorId:  msg.AuthorID,
			Text:      msg.Text,
			CreatedAt: strconv.FormatInt(msg.CreatedAt, 10),
			UpdatedAt: strconv.FormatInt(msg.UpdatedAt, 10),
			Deleted:   msg.Deleted,
		},
	}, nil
}

func (s *ChatServer) DeleteMessage(ctx context.Context, req *chatpb.DeleteMessageRequest) (*chatpb.DeleteMessageResponse, error) {
	_, err := s.svc.DeleteMessage(req.MessageIds, req.HardDelete, req.RequesterId)
	if err != nil {
		return &chatpb.DeleteMessageResponse{Success: false, Message: "failed to delete messages: " + err.Error()}, nil
	}
	return &chatpb.DeleteMessageResponse{Success: true, Message: "deleted"}, nil
}
func (s *ChatServer) ListMessages(ctx context.Context, req *chatpb.ListMessagesRequest) (*chatpb.ListMessagesResponse, error) {
	msgs, cursor, err := s.svc.ListMessages(ctx, req.ChatId, int(req.Limit), req.Cursor)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list messages: %v", err)
	}
	resp := make([]*chatpb.Message, 0, len(msgs))
	for _, m := range msgs {
		resp = append(resp, &chatpb.Message{
			Id:        m.ID,
			ChatId:    m.ChatID,
			AuthorId:  m.AuthorID,
			Text:      m.Text,
			CreatedAt: strconv.FormatInt(m.CreatedAt, 10),
			UpdatedAt: strconv.FormatInt(m.UpdatedAt, 10),
			Deleted:   m.Deleted,
		})
	}
	return &chatpb.ListMessagesResponse{Messages: resp, NextCursor: cursor}, nil
}

func (s *ChatServer) MarkRead(ctx context.Context, req *chatpb.MarkReadRequest) (*chatpb.MarkReadResponse, error) {
	if err := s.svc.MarkRead(ctx, req.ChatId, req.UserId, req.MessageId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark read: %v", err)
	}
	return &chatpb.MarkReadResponse{Success: true}, nil
}

func (s *ChatServer) ToggleSaved(ctx context.Context, req *chatpb.ToggleSavedRequest) (*chatpb.ToggleSavedResponse, error) {
	if err := s.svc.ToggleSaved(ctx, req.UserId, req.MessageId, req.Saved); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to toggle saved: %v", err)
	}
	return &chatpb.ToggleSavedResponse{Success: true}, nil
}

func (s *ChatServer) ListSaved(ctx context.Context, req *chatpb.ListSavedRequest) (*chatpb.ListSavedResponse, error) {
	msgs, cursor, err := s.svc.ListSaved(ctx, req.UserId, int(req.Limit), req.Cursor)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list saved: %v", err)
	}
	resp := make([]*chatpb.Message, 0, len(msgs))
	for _, m := range msgs {
		resp = append(resp, &chatpb.Message{
			Id:        m.ID,
			ChatId:    m.ChatID,
			AuthorId:  m.AuthorID,
			Text:      m.Text,
			CreatedAt: strconv.FormatInt(m.CreatedAt, 10),
			UpdatedAt: strconv.FormatInt(m.UpdatedAt, 10),
			Deleted:   m.Deleted,
		})
	}
	return &chatpb.ListSavedResponse{Messages: resp, NextCursor: cursor}, nil
}

// --- Chats ---

func (s *ChatServer) CreateDirectChat(ctx context.Context, req *chatpb.CreateDirectChatRequest) (*chatpb.ChatResponse, error) {
	chat, err := s.svc.CreateDirect(ctx, req.UserId, req.PeerId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to create direct chat: %v", err)
	}
	return &chatpb.ChatResponse{Chat: toProtoChat(chat)}, nil
}

func (s *ChatServer) CreateGroupChat(ctx context.Context, req *chatpb.CreateGroupChatRequest) (*chatpb.ChatResponse, error) {
	chat, err := s.svc.CreateGroup(ctx, req.UserId, req.MemberIds, req.Title)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group chat: %v", err)
	}
	return &chatpb.ChatResponse{
		Chat: &chatpb.Chat{
			Id:        chat.ID,
			Kind:      string(chat.Kind),
			MemberIds: chat.MemberIDs,
			Title:     chat.Title,
			CreatedBy: chat.CreatedBy,
			CreatedAt: strconv.FormatInt(chat.CreatedAt, 10),
		},
	}, nil
}

func (s *ChatServer) ListChats(ctx context.Context, req *chatpb.ListChatsRequest) (*chatpb.ListChatsResponse, error) {
	chats, cursor, err := s.svc.ListChats(ctx, req.UserId, int(req.Limit), req.Cursor)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list chats: %v", err)
	}
	resp := make([]*chatpb.Chat, 0, len(chats))
	for _, c := range chats {
		resp = append(resp, &chatpb.Chat{
			Id:        c.ID,
			Kind:      string(c.Kind),
			MemberIds: c.MemberIDs,
			Title:     c.Title,
			CreatedBy: c.CreatedBy,
			CreatedAt: strconv.FormatInt(c.CreatedAt, 10),
		})
	}
	return &chatpb.ListChatsResponse{Chats: resp, NextCursor: cursor}, nil
}

func (s *ChatServer) GetChat(ctx context.Context, req *chatpb.GetChatRequest) (*chatpb.ChatResponse, error) {
	chat, err := s.svc.GetChat(ctx, req.ChatId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get chat: %v", err)
	}
	return &chatpb.ChatResponse{
		Chat: &chatpb.Chat{
			Id:        chat.ID,
			Kind:      string(chat.Kind),
			MemberIds: chat.MemberIDs,
			Title:     chat.Title,
			CreatedBy: chat.CreatedBy,
			CreatedAt: strconv.FormatInt(chat.CreatedAt, 10),
		},
	}, nil
}

func (s *ChatServer) UpdateGroupChat(ctx context.Context, req *chatpb.UpdateGroupChatRequest) (*chatpb.ChatResponse, error) {
	chat, err := s.svc.UpdateGroupChat(ctx, req) // chat == domain.Chat
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update group chat: %v", err)
	}
	return &chatpb.ChatResponse{Chat: toProtoChat(chat)}, nil
}

func (s *ChatServer) ListReadMessages(ctx context.Context, req *chatpb.ListReadMessagesRequest) (*chatpb.ListReadMessagesResponse, error) {
	// Используем chat_id из запроса (может быть пустым для получения всех прочитанных сообщений)
	msgs, err := s.svc.ListReadMessages(ctx, req.UserId, req.ChatId, int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list read messages: %v", err)
	}

	var pbMsgs []*chatpb.Message
	for _, m := range msgs {
		pbMsgs = append(pbMsgs, &chatpb.Message{
			Id:        m.ID,
			ChatId:    m.ChatID,
			AuthorId:  m.AuthorID,
			Text:      m.Text,
			CreatedAt: strconv.FormatInt(m.CreatedAt, 10),
			UpdatedAt: strconv.FormatInt(m.UpdatedAt, 10),
			Deleted:   m.Deleted,
		})
	}

	return &chatpb.ListReadMessagesResponse{
		Messages: pbMsgs,
		// NextCursor: cursor, // больше не возвращаем курсор, так как метод теперь возвращает только []Message
	}, nil
}

func toProtoChat(c domain.Chat) *chatpb.Chat {
	return &chatpb.Chat{
		Id:        c.ID,
		Kind:      string(c.Kind),
		MemberIds: c.MemberIDs,
		Title:     c.Title,
		CreatedBy: c.CreatedBy,
		CreatedAt: strconv.FormatInt(c.CreatedAt, 10),
	}
}
