package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"

	"main/internal/config"
	"main/internal/domain"
	"main/internal/logger"
	"main/internal/repository/es"
	"main/internal/repository/postgres"
	"main/internal/service"
	httpserver "main/internal/transport/http"
	kafkaconsumer "main/internal/transport/kafka"
)

type SearchEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Функция для создания всех необходимых топиков
func ensureTopics(broker string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	// Список всех необходимых топиков
	topics := []kafka.TopicConfig{
		{
			Topic:             "messages",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             "users",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             "chats",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             "files",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             "search-events",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             "file.events",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             "media-events",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	return conn.CreateTopics(topics...)
}

func sendTestMessages(w *kafka.Writer) error {
	ctx := context.Background()
	log := logger.GetLogger()
	// 1. Тестовое сообщение типа "message"
	msgEvent := SearchEvent{
		Type: "message",
		Data: domain.Message{
			ID:        "msg-" + time.Now().Format("20060102150405"),
			ChatID:    "chat-001",
			AuthorID:  "user-001",
			Text:      "Привет, как дела?",
			CreatedAt: time.Now().Unix(),
		},
	}
	b1, _ := json.Marshal(msgEvent)
	if err := w.WriteMessages(ctx, kafka.Message{
		Key:   []byte("message"),
		Value: b1,
	}); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	log.Info().Msg(" Test message sent to Kafka")

	// 2. Тестовое сообщение типа "user"
	userEvent := SearchEvent{
		Type: "user",
		Data: domain.UserIndex{
			ID:        "user-001",
			UserName:  "Алексей Петров",
			Email:     "alexey@example.com",
			AboutMe:   "Разработчик ПО, увлекаюсь фотографией",
			UpdatedAt: time.Now().Unix(),
		},
	}
	b2, _ := json.Marshal(userEvent)
	if err := w.WriteMessages(ctx, kafka.Message{
		Key:   []byte("user"),
		Value: b2,
	}); err != nil {
		return fmt.Errorf("failed to write user: %w", err)
	}
	log.Info().Msg(" Test user sent to Kafka")

	// 3. Тестовое сообщение типа "chat"
	chatEvent := SearchEvent{
		Type: "chat",
		Data: domain.ChatIndex{
			ID:        "chat-001",
			Title:     "Общий чат проекта",
			MemberIDs: []string{"user-001", "user-002", "user-003"},
			Kind:      "group",
			UpdatedAt: time.Now().Unix(),
		},
	}
	b3, _ := json.Marshal(chatEvent)
	if err := w.WriteMessages(ctx, kafka.Message{
		Key:   []byte("chat"),
		Value: b3,
	}); err != nil {
		return fmt.Errorf("failed to write chat: %w", err)
	}
	log.Info().Msg(" Test chat sent to Kafka")

	// 4. Тестовое сообщение типа "file" (медиа)
	fileEvent := SearchEvent{
		Type: "file.created",
		Data: domain.FileIndex{
			ID:        "file-" + time.Now().Format("20060102150405"),
			URL:       "http://localhost:9000/media-bucket/presentations/project.pdf",
			Type:      "document",
			SizeBytes: 5242880, // 5MB
			AuthorID:  "user-001",
			ChatID:    "chat-001",
			UpdatedAt: time.Now().Unix(),
		},
	}
	b4, _ := json.Marshal(fileEvent)
	if err := w.WriteMessages(ctx, kafka.Message{
		Key:   []byte("file"),
		Value: b4,
	}); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	log.Info().Msg(" Test file sent to Kafka")

	// 5. Еще один файл (изображение)
	imageEvent := SearchEvent{
		Type: "file",
		Data: domain.FileIndex{
			ID:        "img-" + time.Now().Format("20060102150405"),
			URL:       "http://localhost:9000/media-bucket/images/architecture.png",
			Type:      "image",
			SizeBytes: 2097152, // 2MB
			AuthorID:  "user-002",
			ChatID:    "chat-001",
			UpdatedAt: time.Now().Unix(),
		},
	}
	b5, _ := json.Marshal(imageEvent)
	if err := w.WriteMessages(ctx, kafka.Message{
		Key:   []byte("image"),
		Value: b5,
	}); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}
	log.Info().Msg(" Test image sent to Kafka")

	return nil
}

func main() {
	config := config.New()

	// Инициализация логгера
	logger.Init(config.LogLevel, config.LogPretty)
	log := logger.GetLogger()
	ctx := context.Background()

	log.Info().Msg("Starting Search Service...")

	// Создаём все топики Kafka
	log.Info().Msg("Creating Kafka topics...")
	if err := ensureTopics(config.KafkaBroker); err != nil {
		log.Warn().Err(err).Msg("Failed to create topics (they may already exist)")
	} else {
		log.Info().Msg(" Kafka topics created successfully")
	}

	// Продюсер Kafka для тестовых сообщений
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{config.KafkaBroker},
		Topic:   "search-events", // Используем общий топик для всех событий
	})
	defer w.Close()

	// Отправляем тестовые сообщения
	log.Info().Msg(" Sending test messages to Kafka...")
	if err := sendTestMessages(w); err != nil {
		log.Fatal().Err(err).Msg("Failed to send test messages")
	}

	// Подключение к Postgres
	log.Info().Msg(" Connecting to PostgreSQL...")
	db, err := sql.Open("postgres", config.PostgresDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Проверка соединения с БД
	if err := db.Ping(); err != nil {
		log.Warn().Err(err).Msg(" Database connection failed, continuing without PostgreSQL")
	}

	// Подключение к Elasticsearch
	log.Info().Msg(" Connecting to Elasticsearch...")
	esRepo, err := es.NewRepo(config.ElasticURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Elasticsearch")
	}

	// Создание индексов в Elasticsearch
	log.Info().Msg(" Ensuring Elasticsearch indices...")
	if err := esRepo.EnsureIndices(ctx); err != nil {
		log.Warn().Err(err).Msg(" Failed to create indices (they may already exist)")
	} else {
		log.Info().Msg(" Elasticsearch indices created")
	}

	pgRepo := postgres.NewRepo(db, log)

	// Запускаем Kafka consumers для разных топиков
	log.Info().Msg(" Starting Kafka consumers...")

	// Consumer для общего топика search-events
	searchConsumer := kafkaconsumer.NewConsumer(
		[]string{config.KafkaBroker},
		"search-events",         // топик
		"search-consumer-group", // группа
		esRepo,
		pgRepo,
		log,
	)

	// Consumer для файловых событий
	fileConsumer := kafkaconsumer.NewConsumer(
		[]string{config.KafkaBroker},
		"file.events",         // топик для медиа
		"file-consumer-group", // группа
		esRepo,
		pgRepo,
		log,
	)

	// Запуск consumers в отдельных горутинах
	go func() {
		log.Info().Msg(" Search consumer started")
		searchConsumer.Run(ctx)
	}()

	go func() {
		log.Info().Msg(" File consumer started")
		fileConsumer.Run(ctx)
	}()

	// Создаем и запускаем HTTP сервис
	log.Info().Msg(" Starting HTTP server...")
	svc := service.NewSearchService(esRepo, log)
	server := httpserver.NewServer(svc, log)

	// Запускаем тестовый поиск через несколько секунд
	go func() {
		time.Sleep(3 * time.Second) // Даем время на индексацию
		log.Info().Msg("🔍 Running test search queries...")

		// Тестовые поисковые запросы
		testQueries := []struct {
			query string
			typ   string
		}{
			{"pdf", "file"},
			{"presentation", "file"},
			{"диаграмма", "file"},
			{"Алексей", "user"},
			{"проект", "chat"},
			{"привет", "message"},
		}

		for _, test := range testQueries {
			log.Info().Str("query", test.query).Str("type", test.typ).Msg("Testing search...")
			// Здесь можно добавить вызов поиска через HTTP
		}
	}()

	log.Info().Msgf(" Search Service HTTP listening on %s", config.HTTPPort)
	log.Info().Msgf(" Health check: http://localhost%s/health", config.HTTPPort)
	log.Info().Msgf(" Search endpoint: http://localhost%s/search?q=pdf&type=file", config.HTTPPort)

	if err := server.Listen(config.HTTPPort); err != nil {
		log.Fatal().Err(err).Msg("HTTP server failed")
	}
}
