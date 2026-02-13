package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"main/internal/repository"
	"main/internal/service"
	trgrpc "main/internal/transport/grpc"
	"main/internal/transport/websocket"
	api "main/pkg/api"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// ctx := context.Background()
	var wg sync.WaitGroup

	// Инициализация Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisRepo := repository.NewRoomRedisRepo(redisAddr)

	// Сервис
	svc := service.NewRoomService(redisRepo)

	// WebSocket сервер
	wsServer := websocket.NewWebSocketServer()

	// Запускаем gRPC сервер
	wg.Add(1)
	go func() {
		defer wg.Done()
		runGRPCServer(svc)
	}()

	// Запускаем HTTP сервер с WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		runHTTPServer(wsServer)
	}()

	log.Println("All servers started")
	wg.Wait()
}

func runGRPCServer(svc service.RoomService) {
	lis, err := net.Listen("tcp", ":8087")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	handler := trgrpc.NewRoomHandler(svc, websocket.NewWebSocketServer())
	api.RegisterRoomServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	log.Println("RoomService gRPC listening on :8087")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func runHTTPServer(wsServer *websocket.WebSocketServer) {
	r := mux.NewRouter()

	// WebSocket endpoint
	r.HandleFunc("/ws", wsServer.HandleWebSocket)

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Статические файлы (для демо)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend")))

	server := &http.Server{
		Addr:    ":8099", // чтоб был фронт
		Handler: r,
	}

	log.Println("HTTP server with WebSocket listening on :8087")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
