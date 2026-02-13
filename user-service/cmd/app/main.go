package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"main/internal/config"
	"main/internal/logger"
	"main/internal/repository"
	redisrepo "main/internal/repository/redis"
	"main/internal/service"
	transportGrpc "main/internal/transport/grpc"
	transportKafka "main/internal/transport/kafka"
	user "main/pkg/api"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var ctx = context.Background()

func ensureTopic(broker, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
}

func initPostgres() *sql.DB {
	cfg := config.Load()
	dsn := cfg.DatabaseURL
	fmt.Println(dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к Postgres: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Postgres недоступен: %v", err)
	}
	return db
}

func initRedis() *redisrepo.Client {
	cfg := config.Load()
	addr := cfg.RedisAddr
	if addr == "" {
		addr = "localhost:6379" // дефолт для локального запуска
	}
	return redisrepo.New(addr)
}

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем логгер
	logger.Init(cfg.LogLevel, cfg.LogPretty)
	log := logger.GetLogger()

	// Инициализация базы данных
	db := initPostgres()
	defer db.Close()

	// Инициализация Redis
	redisClient := initRedis()

	// Инициализация репозитория
	repo := repository.NewUserRepository(db)

	// Инициализация Kafka writer
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBroker),
		Topic:    "users",
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	// Создание сервиса
	userService := service.NewUserService(repo, kafkaWriter, redisClient)

	// Создание Kafka topic
	if err := ensureTopic(cfg.KafkaBroker, cfg.KafkaTopicSearch); err != nil {
		log.Warn().Err(err).Msg("Не удалось создать Kafka topic (возможно уже существует)")
	}

	// Запуск Kafka consumer
	go transportKafka.KafkaCons(userService)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer()
	userHandler := transportGrpc.NewUserHandler(userService)
	user.RegisterUserServiceServer(grpcServer, userHandler)
	reflection.Register(grpcServer)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", cfg.UserServicePort)
	fmt.Println(cfg.UserServicePort)
	if err != nil {
		log.Fatal().Err(err).Msgf("Не удалось слушать порт %s", cfg.UserServicePort)
	}

	log.Info().Msgf("gRPC server listening on port %s", cfg.UserServicePort)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("Ошибка запуска gRPC сервера")
		}
	}()

	<-stop
	log.Info().Msg("Получен сигнал завершения")
	grpcServer.GracefulStop()
	log.Info().Msg("Сервер корректно завершил работу")
}
