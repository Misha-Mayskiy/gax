package repository

import (
	"main/internal/domain"
	user "main/pkg/api_user_service"
)

type ChatRepository interface {
	CreateDirect(userID, peerID string, userClient user.UserServiceClient) (domain.Chat, error)
	CreateGroup(creatorID string, members []string, title string, userClient user.UserServiceClient) (domain.Chat, error)
	UpdateGroup(chatID string, title *string, addMembers, removeMembers []string, requesterID string, userClient user.UserServiceClient) (domain.Chat, error)
	Get(chatID string) (domain.Chat, error)
	List(userID string, limit int, cursor string) ([]domain.Chat, string, error)
}

type MessageRepository interface {
	Send(msg domain.Message) (domain.Message, error)
	Get(id string) (domain.Message, error)
	Update(messageID, authorID string, text *string, media *[]domain.Media) (domain.Message, error)
	Delete(messageIDs []string, hard bool, requesterID string) ([]domain.Message, error)
	List(chatID string, limit int, cursor string) ([]domain.Message, string, error)
	MarkRead(chatID, userID, messageID string) error
	ToggleSaved(userID, messageID string, saved bool) error
	ListSaved(userID string, limit int, cursor string) ([]domain.Message, string, error)
	ListReadMessages(userID, chatID string, limit int) ([]domain.Message, error)
	GetUnreadCount(chatID, userID string) (int64, error)
}
