package mongo

import (
	"context"
	"fmt"
	"main/internal/domain"
	userserviceclient "main/internal/user-service-client"
	user "main/pkg/api_user_service"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Интерфейсы для тестирования
type Collection interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)
	CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
}

type ChatRepo struct {
	col Collection
	log zerolog.Logger
}

func NewChatRepo(db *mongo.Database, log zerolog.Logger) *ChatRepo {
	return &ChatRepo{col: db.Collection("chats"), log: log}
}

// NewTestChatRepo - конструктор для тестов
func NewTestChatRepo(col Collection, log zerolog.Logger) *ChatRepo {
	return &ChatRepo{col: col, log: log}
}
func (r *ChatRepo) CreateDirect(userID, peerID string, userClient user.UserServiceClient) (domain.Chat, error) {
	// Проверяем существование обоих пользователей
	_, err := userserviceclient.GetUserInfo(userClient, userID)
	if err != nil {
		return domain.Chat{}, fmt.Errorf("пользователь %s не найден: %w", userID, err)
	}

	_, err = userserviceclient.GetUserInfo(userClient, peerID)
	if err != nil {
		return domain.Chat{}, fmt.Errorf("пользователь %s не найден: %w", peerID, err)
	}

	// Проверяем, не существует ли уже такой чат
	existingChatID := userID + "_" + peerID
	existingChatIDReverse := peerID + "_" + userID

	count, err := r.col.CountDocuments(context.Background(), bson.M{
		"id": bson.M{"$in": []string{existingChatID, existingChatIDReverse}},
	})
	if err != nil {
		return domain.Chat{}, fmt.Errorf("ошибка проверки существующего чата: %w", err)
	}

	if count > 0 {
		return domain.Chat{}, fmt.Errorf("чат между пользователями уже существует")
	}

	// Создаем чат
	chat := domain.Chat{
		ID:        existingChatID,
		Kind:      domain.ChatKindDirect,
		MemberIDs: []string{userID, peerID},
		CreatedBy: userID,
		CreatedAt: time.Now().Unix(),
	}

	_, err = r.col.InsertOne(context.Background(), chat)
	return chat, err
}
func (r *ChatRepo) CreateGroup(creatorID string, members []string, title string, userClient user.UserServiceClient) (domain.Chat, error) {
	// Проверяем создателя
	_, err := userserviceclient.GetUserInfo(userClient, creatorID)
	if err != nil {
		return domain.Chat{}, fmt.Errorf("создатель %s не найден: %w", creatorID, err)
	}

	// Проверяем всех участников
	for _, memberID := range members {
		_, err := userserviceclient.GetUserInfo(userClient, memberID)
		if err != nil {
			return domain.Chat{}, fmt.Errorf("участник %s не найден: %w", memberID, err)
		}
	}

	// Все пользователи существуют → создаём группу
	groupID := uuid.New().String()
	chat := domain.Chat{
		ID:        groupID,
		Kind:      domain.ChatKindGroup,
		MemberIDs: append(members, creatorID),
		Title:     title,
		CreatedBy: creatorID,
		CreatedAt: time.Now().Unix(),
	}
	ctx := context.Background()
	_, err = r.col.InsertOne(ctx, chat)
	return chat, err
}

func (r *ChatRepo) UpdateGroup(chatID string, title *string, addMembers, removeMembers []string, requesterID string, userClient user.UserServiceClient) (domain.Chat, error) {

	ctx := context.Background()

	// Загружаем чат
	var chat domain.Chat
	if err := r.col.FindOne(ctx, bson.M{"id": chatID}).Decode(&chat); err != nil {
		return domain.Chat{}, fmt.Errorf("чат %s не найден: %w", chatID, err)
	}

	// Проверяем инициатора — только создатель может менять
	if chat.CreatedBy != requesterID {
		return domain.Chat{}, fmt.Errorf("пользователь %s не имеет права изменять чат %s", requesterID, chatID)
	}
	for _, memberID := range addMembers {
		_, err := userserviceclient.GetUserInfo(userClient, memberID)
		if err != nil {
			return domain.Chat{}, fmt.Errorf("участник %s не найден: %w", memberID, err)
		}
	}
	// Формируем одно атомарное обновление
	set := bson.M{}
	if title != nil {
		set["title"] = *title
	}
	ops := bson.M{}
	if len(set) > 0 {
		ops["$set"] = set
	}
	if len(addMembers) > 0 {
		_, err := r.col.UpdateOne(ctx,
			bson.M{"id": chatID},
			bson.M{"$addToSet": bson.M{"memberids": bson.M{"$each": addMembers}}},
		)
		if err != nil {
			return domain.Chat{}, err
		}
	}

	if len(removeMembers) > 0 {
		_, err := r.col.UpdateOne(ctx,
			bson.M{"id": chatID},
			bson.M{"$pullAll": bson.M{"memberids": removeMembers}},
		)
		if err != nil {
			return domain.Chat{}, err
		}
	}
	if len(ops) > 0 {
		if _, err := r.col.UpdateOne(ctx, bson.M{"id": chatID}, ops); err != nil {
			return domain.Chat{}, err
		}
	}

	// Возвращаем обновлённый чат
	if err := r.col.FindOne(ctx, bson.M{"id": chatID}).Decode(&chat); err != nil {
		return domain.Chat{}, err
	}
	return chat, nil
}

func (r *ChatRepo) Get(chatID string) (domain.Chat, error) {
	var chat domain.Chat
	err := r.col.FindOne(context.Background(), bson.M{"id": chatID}).Decode(&chat)
	return chat, err
}

func (r *ChatRepo) List(userID string, limit int, cursor string) ([]domain.Chat, string, error) {
	ctx := context.Background()
	filter := bson.M{"member_ids": bson.M{"$in": []string{userID}}}
	opts := options.Find().SetLimit(int64(limit))
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	defer cur.Close(ctx)
	var chats []domain.Chat
	if err := cur.All(ctx, &chats); err != nil {
		return nil, "", err
	}
	nextCursor := ""
	if len(chats) > 0 {
		nextCursor = chats[len(chats)-1].ID
	}
	return chats, nextCursor, nil
}
