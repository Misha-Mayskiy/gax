package service

import (
	"context"
	"main/internal/domain"
	"main/internal/repository"
	chatpb "main/pkg/api"
	user "main/pkg/api_user_service"
	"time"

	"github.com/google/uuid"
)

// Интерфейс продюсера Kafka
type KafkaProducer interface {
	PublishNewMessage(ctx context.Context, e domain.NewMessageEvent) error
	PublishEvent(ctx context.Context, evt domain.SearchEvent) error
}

// ChatService — бизнес-логика
type ChatService struct {
	chats      repository.ChatRepository
	msgs       repository.MessageRepository
	kafka      KafkaProducer
	userClient user.UserServiceClient
}

// Конструктор
func NewChatService(
	ch repository.ChatRepository,
	mr repository.MessageRepository,
	kp KafkaProducer,
	uc user.UserServiceClient,
) *ChatService {
	return &ChatService{
		chats:      ch,
		msgs:       mr,
		kafka:      kp,
		userClient: uc,
	}
}

// --- Методы ---

// Создание приватного чата
func (s *ChatService) CreateDirect(ctx context.Context, userID, peerID string) (domain.Chat, error) {
	// Передаем userClient из структуры сервиса
	return s.chats.CreateDirect(userID, peerID, s.userClient)
}

// Создание группового чата
func (s *ChatService) CreateGroup(ctx context.Context, creatorID string, members []string, title string) (domain.Chat, error) {
	chat, err := s.chats.CreateGroup(creatorID, members, title, s.userClient)
	if err == nil {
		_ = s.kafka.PublishEvent(ctx, domain.SearchEvent{Type: "chat", Data: chat})
	}
	return s.chats.CreateGroup(creatorID, members, title, s.userClient)
}

// Отправка сообщения
func (s *ChatService) SendMessage(ctx context.Context, m domain.Message) (domain.Message, error) {
	m.ID = uuid.New().String() // или primitive.NewObjectID().Hex()
	m.CreatedAt = time.Now().Unix()
	msg, err := s.msgs.Send(m)
	if err != nil {
		return msg, err
	}

	// Отправляем специализированное событие
	event := domain.NewMessageEvent{
		MessageID: msg.ID,
		ChatID:    msg.ChatID,
		AuthorID:  msg.AuthorID,
		Text:      msg.Text,
		Timestamp: msg.CreatedAt,
	}
	_ = s.kafka.PublishNewMessage(ctx, event)

	// Отправляем универсальное событие для SearchService
	go func(ctx context.Context) {
		_ = s.kafka.PublishEvent(ctx, domain.SearchEvent{Type: "message", Data: msg})
	}(ctx)

	return msg, nil
}

// Обновление сообщения
func (s *ChatService) UpdateMessage(ctx context.Context, messageID, authorID string, text *string, media *[]domain.Media) (domain.Message, error) {
	return s.msgs.Update(messageID, authorID, text, media)
}

// Удаление сообщения
func (s *ChatService) DeleteMessage(messageIDs []string, hard bool, requesterID string) ([]domain.Message, error) {
	return s.msgs.Delete(messageIDs, hard, requesterID)
}

// Получение списка сообщений
func (s *ChatService) ListMessages(ctx context.Context, chatID string, limit int, cursor string) ([]domain.Message, string, error) {
	return s.msgs.List(chatID, limit, cursor)
}

// Отметка прочтения
func (s *ChatService) MarkRead(ctx context.Context, chatID, userID, messageID string) error {
	return s.msgs.MarkRead(chatID, userID, messageID)
}

// Работа с избранным
func (s *ChatService) ToggleSaved(ctx context.Context, userID, messageID string, saved bool) error {
	return s.msgs.ToggleSaved(userID, messageID, saved)
}

func (s *ChatService) ListSaved(ctx context.Context, userID string, limit int, cursor string) ([]domain.Message, string, error) {
	return s.msgs.ListSaved(userID, limit, cursor)
}

// Получение списка чатов пользователя
func (s *ChatService) ListChats(ctx context.Context, userID string, limit int, cursor string) ([]domain.Chat, string, error) {
	return s.chats.List(userID, limit, cursor)
}

// Получение информации о чате
func (s *ChatService) GetChat(ctx context.Context, chatID string) (domain.Chat, error) {
	return s.chats.Get(chatID)
}

// Обновление группового чата
func (s *ChatService) UpdateGroupChat(ctx context.Context, req *chatpb.UpdateGroupChatRequest) (domain.Chat, error) {
	var titlePtr *string
	if req.Title != "" {
		// только если клиент передал непустой title — обновим
		t := req.Title
		titlePtr = &t
	}
	return s.chats.UpdateGroup(
		req.ChatId,
		titlePtr,
		req.AddMemberIds,
		req.RemoveMemberIds,
		req.RequesterId,
		s.userClient,
	)
}

func (s *ChatService) ListReadMessages(ctx context.Context, userID, chatID string, limit int) ([]domain.Message, error) {
	msgs, err := s.msgs.ListReadMessages(userID, chatID, limit)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (s *ChatService) publishToKafkaMessage(m domain.Message) error {
	evt := domain.NewMessageEvent{
		MessageID: m.ID,
		ChatID:    m.ChatID,
		AuthorID:  m.AuthorID,
		Text:      m.Text,
		Timestamp: m.CreatedAt,
	}
	return s.kafka.PublishNewMessage(context.Background(), evt)
}
