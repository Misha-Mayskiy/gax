package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/config"
	"main/internal/domain"
	"main/internal/transport/kafka/models"

	"github.com/segmentio/kafka-go"
)

func KafkaProd(user *domain.User) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{config.Load().KafkaBroker},
		Topic:   models.WriteTopikName,
	})
	defer writer.Close()

	// сериализация структуры в JSON
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	// отправка сообщения
	err = writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: data,
		},
	)
	if err != nil {
		return fmt.Errorf("ошибка при отправке: %w", err)
	}

	return nil
}
