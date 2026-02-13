package main

import (
	"context"
	"net"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"main/internal/config"
	"main/internal/logger"
	mongorepo "main/internal/repository/mongo"
	userserviceclient "main/internal/user-service-client"
	chatpb "main/pkg/api"

	trgrpc "main/internal/transport/grpc"

	"main/internal/service"
	"main/internal/transport/kafka"
)

func main() {
	config := config.New()

	// logger init
	logger.Init(config.LogLevel, config.LogPretty)

	log := logger.GetLogger()
	// Mongo
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.MongoURI))
	if err != nil {
		log.Fatal().Msgf("failed to connect to mongo: %s", err)
	}
	defer client.Disconnect(context.Background())
	mongoDB := client.Database("chatdb")

	// Kafka
	kp := kafka.NewProducer([]string{config.KafkaBroker}, config.KafkaTopic)

	// Репозитории
	chatRepo := mongorepo.NewChatRepo(mongoDB, log)
	messageRepo := mongorepo.NewMessageRepo(mongoDB)
	//подключение к клиенту
	userClient := userserviceclient.NewUserClient(config.UserServiceAddr, log)
	// Сервис
	svc := service.NewChatService(chatRepo, messageRepo, kp, userClient)

	// gRPC сервер
	lis, err := net.Listen("tcp", config.ChatServicePort)
	if err != nil {
		log.Fatal().Msgf("failed to listen: %s", err)

	}
	grpcServer := grpc.NewServer()

	// Регистрируем именно ChatServer
	chatpb.RegisterChatServiceServer(grpcServer, trgrpc.NewChatServer(svc))

	// Reflection для grpcurl
	reflection.Register(grpcServer)

	log.Info().Msgf("ChatService gRPC запущен на %s", config.ChatServicePort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}
