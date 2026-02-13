package service

import (
	"context"
	"main/internal/domain"
	chatpb "main/pkg/api"
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
