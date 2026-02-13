package kafka

import (
	"context"
	"encoding/json"
	"main/internal/domain"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) PublishNewMessage(ctx context.Context, e domain.NewMessageEvent) error {
	data, _ := json.Marshal(e)
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(e.MessageID),
		Value: data,
	})
}

func (p *Producer) PublishEvent(ctx context.Context, evt domain.SearchEvent) error {
	data, _ := json.Marshal(evt)
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(evt.Type),
		Value: data,
	})
}
