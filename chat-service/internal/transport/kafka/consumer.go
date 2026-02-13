package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func StartConsumer(brokers []string, topic, groupID string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			fmt.Println("Ошибка чтения:", err)
			continue
		}
		fmt.Printf("Получено сообщение: key=%s value=%s\n", string(m.Key), string(m.Value))
	}
}
