package kafka

import (
	"context"
	"encoding/json"
	"log"
	"media-service/internal/domain"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer реализует интерфейс service.EventProducer
type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		// Асинхронная запись для скорости
		Async: true,
	}
	return &Producer{writer: w}
}

// SendFileUploaded — это тот метод, который требует интерфейс
func (p *Producer) SendFileUploaded(meta *domain.FileMeta) error {
	// 1. Формируем JSON
	payload, err := json.Marshal(map[string]interface{}{
		"event_type": "FileUploaded",
		"payload":    meta,
		"timestamp":  time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	// 2. Отправляем в Kafka (с таймаутом)
	// Используем ID файла как Key, чтобы события одного файла шли в одну партицию
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(meta.ID),
		Value: payload,
	})

	if err != nil {
		log.Printf("Failed to send kafka event: %v", err)
	}
	return err
}

// Close закрывает соединение (нужно для graceful shutdown)
func (p *Producer) Close() error {
	return p.writer.Close()
}
