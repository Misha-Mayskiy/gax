package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/internal/config"
	"main/internal/service"
	"main/internal/transport/kafka/models"

	"github.com/segmentio/kafka-go"
)

func KafkaCons(userSvc service.UserService) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{config.Load().KafkaBroker},
		Topic:   models.ReadTopikName,
		GroupID: models.GroupIDname,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("Ошибка при получении:", err)
			continue
		}

		fmt.Printf("Получено сообщение: %s\n", string(msg.Value))

		// Десериализация JSON в CreateUserRequest
		var req service.CreateUserRequest
		if err := json.Unmarshal(msg.Value, &req); err != nil {
			log.Println("Ошибка парсинга:", err)
			continue
		}

		// Вызов ручки CreateUser
		user, err := userSvc.CreateUser(context.Background(), &req)
		if err != nil {
			log.Println("Ошибка при создании пользователя:", err)
		}
		err = KafkaProd(user)
		if err != nil {
			log.Println("Ошибка при создании пользователя:", err)
		}
	}
}
