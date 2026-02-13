package mongo

import (
	"context"
	"main/internal/domain"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageRepo struct {
	col Collection
}

func NewMessageRepo(db *mongo.Database) *MessageRepo {
	return &MessageRepo{
		col: db.Collection("messages"),
	}
}

// NewTestMessageRepo - конструктор для тестов
func NewTestMessageRepo(col Collection) *MessageRepo {
	return &MessageRepo{col: col}
}

func (r *MessageRepo) Send(m domain.Message) (domain.Message, error) {
	m.ID = uuid.New().String()
	m.CreatedAt = time.Now().Unix()
	m.ReadBy = []domain.ReadInfo{}
	m.SavedBy = []domain.SavedInfo{}

	_, err := r.col.InsertOne(context.Background(), m)
	return m, err
}

func (r *MessageRepo) Get(id string) (domain.Message, error) {
	var m domain.Message
	err := r.col.FindOne(context.Background(), bson.M{"id": id}).Decode(&m)
	return m, err
}

func (r *MessageRepo) Update(messageID, authorID string, text *string, media *[]domain.Media) (domain.Message, error) {
	ctx := context.Background()

	filter := bson.M{"id": messageID, "author_id": authorID, "deleted": false}
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now().Unix(),
		},
	}

	if text != nil {
		update["$set"].(bson.M)["text"] = *text
	}
	if media != nil {
		update["$set"].(bson.M)["media"] = *media
	}

	result, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return domain.Message{}, err
	}
	if result.MatchedCount == 0 {
		return domain.Message{}, mongo.ErrNoDocuments
	}

	var msg domain.Message
	err = r.col.FindOne(ctx, bson.M{"id": messageID}).Decode(&msg)
	return msg, err
}

func (r *MessageRepo) Delete(messageIDs []string, hard bool, requesterID string) ([]domain.Message, error) {
	ctx := context.Background()

	filter := bson.M{
		"id":        bson.M{"$in": messageIDs},
		"author_id": requesterID,
	}

	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []domain.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return messages, nil
	}

	var idsToDelete []string
	for _, msg := range messages {
		idsToDelete = append(idsToDelete, msg.ID)
	}

	if hard {
		_, err = r.col.DeleteMany(ctx, bson.M{"id": bson.M{"$in": idsToDelete}})
	} else {
		update := bson.M{
			"$set": bson.M{
				"deleted":    true,
				"deleted_at": time.Now().Unix(),
			},
		}
		_, err = r.col.UpdateMany(ctx, bson.M{"id": bson.M{"$in": idsToDelete}}, update)
	}

	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepo) List(chatID string, limit int, cursor string) ([]domain.Message, string, error) {
	ctx := context.Background()

	filter := bson.M{"chat_id": chatID, "deleted": false}

	if cursor != "" {
		filter["id"] = bson.M{"$gt": cursor}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "id", Value: 1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	defer cur.Close(ctx)

	var messages []domain.Message
	if err := cur.All(ctx, &messages); err != nil {
		return nil, "", err
	}

	nextCursor := ""
	if len(messages) > 0 {
		nextCursor = messages[len(messages)-1].ID
	}

	return messages, nextCursor, nil
}

func (r *MessageRepo) MarkRead(chatID, userID, messageID string) error {
	ctx := context.Background()

	filter := bson.M{
		"id":              messageID,
		"chat_id":         chatID,
		"read_by.user_id": bson.M{"$ne": userID},
	}

	update := bson.M{
		"$push": bson.M{
			"read_by": domain.ReadInfo{
				UserID: userID,
				ReadAt: time.Now().Unix(),
			},
		},
	}

	_, err := r.col.UpdateOne(ctx, filter, update)
	return err
}

func (r *MessageRepo) ToggleSaved(userID, messageID string, saved bool) error {
	ctx := context.Background()

	if saved {
		filter := bson.M{"id": messageID}
		update := bson.M{
			"$addToSet": bson.M{
				"saved_by": domain.SavedInfo{
					UserID:  userID,
					SavedAt: time.Now().Unix(),
				},
			},
		}
		_, err := r.col.UpdateOne(ctx, filter, update)
		return err
	} else {
		filter := bson.M{"id": messageID}
		update := bson.M{
			"$pull": bson.M{
				"saved_by": bson.M{"user_id": userID},
			},
		}
		_, err := r.col.UpdateOne(ctx, filter, update)
		return err
	}
}

func (r *MessageRepo) ListSaved(userID string, limit int, cursor string) ([]domain.Message, string, error) {
	ctx := context.Background()

	filter := bson.M{
		"saved_by.user_id": userID,
		"deleted":          false,
	}

	if cursor != "" {
		filter["id"] = bson.M{"$gt": cursor}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "id", Value: 1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	defer cur.Close(ctx)

	var messages []domain.Message
	if err := cur.All(ctx, &messages); err != nil {
		return nil, "", err
	}

	nextCursor := ""
	if len(messages) > 0 {
		nextCursor = messages[len(messages)-1].ID
	}

	return messages, nextCursor, nil
}

func (r *MessageRepo) ListReadMessages(userID string, chatID string, limit int) ([]domain.Message, error) {
	ctx := context.Background()

	filter := bson.M{
		"read_by.user_id": userID,
		"deleted":         false,
	}

	if chatID != "" {
		filter["chat_id"] = chatID
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var messages []domain.Message
	if err := cur.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepo) GetUnreadCount(chatID, userID string) (int64, error) {
	ctx := context.Background()

	filter := bson.M{
		"chat_id":         chatID,
		"read_by.user_id": bson.M{"$ne": userID},
		"deleted":         false,
	}

	count, err := r.col.CountDocuments(ctx, filter)
	return count, err
}
