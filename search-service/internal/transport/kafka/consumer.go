package kafka

import (
	"context"
	"encoding/json"
	"log"
	"main/internal/domain"
	"main/internal/repository/es"
	"main/internal/repository/postgres"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

type SearchEvent struct {
	Type string      `json:"type"` // "user", "chat", "message"
	Data interface{} `json:"data"` // сам объект
}
type Consumer struct {
	reader *kafka.Reader
	esRepo *es.Repo
	pgRepo *postgres.Repo
	topic  string
	log    zerolog.Logger
}

func NewConsumer(brokers []string, topic, groupID string, esRepo *es.Repo, pgRepo *postgres.Repo, log zerolog.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
		Logger:  &log,
	})
	return &Consumer{reader: r, esRepo: esRepo, pgRepo: pgRepo, topic: topic}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Kafka read error: %v", err)
			continue
		}

		var evt SearchEvent
		if err := json.Unmarshal(m.Value, &evt); err != nil {
			c.log.Err(err).Msg(" Unmarshal error")
			continue
		}
		c.log.Info().Msg("")

		switch evt.Type {
		case "user":
			var u domain.UserIndex
			b, _ := json.Marshal(evt.Data)
			if err := json.Unmarshal(b, &u); err != nil {
				c.log.Err(err).Msg(" User decode error")
				continue
			}
			if err := c.pgRepo.SaveUser(ctx, u, c.log); err != nil {
				c.log.Err(err).Msg(" PG save user error")

			}
			if err := c.esRepo.IndexUser(ctx, u); err != nil {
				c.log.Err(err).Msg(" ES index user error")

			}
			c.log.Info().Msg("sent new user")

		case "chat":
			var chat domain.Chat
			b, _ := json.Marshal(evt.Data)
			if err := json.Unmarshal(b, &chat); err != nil {
				c.log.Err(err).Msg(" Chat decode error")

				continue
			}
			if err := c.pgRepo.SaveChat(ctx, chat, c.log); err != nil {
				c.log.Err(err).Msg(" PG save chat error")

				log.Printf("PG save chat error: %v", err)
			}
			if err := c.esRepo.IndexChat(ctx, chat); err != nil {
				c.log.Err(err).Msg(" ES index chat error")

			}
			c.log.Info().Msg("sent new chat")

		case "message":
			var msg domain.Message
			b, _ := json.Marshal(evt.Data)
			if err := json.Unmarshal(b, &msg); err != nil {
				c.log.Err(err).Msg(" Message decode error")
				continue
			}
			if err := c.pgRepo.SaveMessage(ctx, msg, c.log); err != nil {
				c.log.Err(err).Msg(" PG save message error")

			}
			if err := c.esRepo.IndexMessage(ctx, msg); err != nil {
				c.log.Err(err).Msg(" ES index message error")

			}
			c.log.Info().Msg("sent new message")

		case "file.events":
			c.handleFileEvent(ctx, evt.Data)

		case "file.created", "file.uploaded":
			c.handleFileCreated(ctx, evt.Data)

		case "file.deleted":
			c.handleFileDeleted(ctx, evt.Data)

		case "file.updated":
			c.handleFileUpdated(ctx, evt.Data)
		}

	}
}

// handleFileEvent обрабатывает события связанные с файлами
func (c *Consumer) handleFileEvent(ctx context.Context, data interface{}) {
	var file domain.FileIndex
	b, _ := json.Marshal(data)
	if err := json.Unmarshal(b, &file); err != nil {
		c.log.Err(err).Msg(" File decode error")
		return
	}

	// Индексируем файл в Elasticsearch
	if err := c.esRepo.IndexFile(ctx, file); err != nil {
		c.log.Err(err).Msg(" ES index file error")
		return
	}

	c.log.Info().Str("file_id", file.ID).Msg("File indexed in Elasticsearch")
}

// handleFileCreated обрабатывает создание файла
func (c *Consumer) handleFileCreated(ctx context.Context, data interface{}) {
	var file domain.FileIndex
	b, _ := json.Marshal(data)
	if err := json.Unmarshal(b, &file); err != nil {
		c.log.Err(err).Msg(" File created decode error")
		return
	}

	// Устанавливаем timestamp если не установлен
	if file.UpdatedAt == 0 {
		file.UpdatedAt = time.Now().Unix()
	}

	// Индексируем в Elasticsearch
	if err := c.esRepo.IndexFile(ctx, file); err != nil {
		c.log.Err(err).Msg(" ES index file (created) error")
		return
	}

	c.log.Info().Str("file_id", file.ID).Str("filename", file.ID).Msg("File created event processed")
}

// handleFileDeleted обрабатывает удаление файла
func (c *Consumer) handleFileDeleted(ctx context.Context, data interface{}) {
	var payload map[string]interface{}
	b, _ := json.Marshal(data)
	if err := json.Unmarshal(b, &payload); err != nil {
		c.log.Err(err).Msg(" File deleted decode error")
		return
	}

	fileID, ok := payload["id"].(string)
	if !ok {
		c.log.Error().Msg("File deleted event missing ID")
		return
	}

	// TODO: Удалить файл из Elasticsearch
	// Нужно добавить метод DeleteFile в es.Repo
	// if err := c.esRepo.DeleteFile(ctx, fileID); err != nil {
	//     c.log.Err(err).Msg(" ES delete file error")
	// }

	c.log.Info().Str("file_id", fileID).Msg("File deleted event processed")
}

// handleFileUpdated обрабатывает обновление файла
func (c *Consumer) handleFileUpdated(ctx context.Context, data interface{}) {
	var file domain.FileIndex
	b, _ := json.Marshal(data)
	if err := json.Unmarshal(b, &file); err != nil {
		c.log.Err(err).Msg(" File updated decode error")
		return
	}

	// Обновляем timestamp
	file.UpdatedAt = time.Now().Unix()

	// Обновляем индекс в Elasticsearch
	if err := c.esRepo.IndexFile(ctx, file); err != nil {
		c.log.Err(err).Msg(" ES index file (updated) error")
		return
	}

	c.log.Info().Str("file_id", file.ID).Msg("File updated event processed")
}
