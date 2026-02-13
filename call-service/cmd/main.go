package main

import (
	"call-service/internal/repository"
	"call-service/internal/sfu"
	wsTransport "call-service/internal/transport/websocket"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6380"
	}

	redisRepo := repository.NewRedisRepo(redisAddr, "")
	log.Println("Connected to Redis at", redisAddr)

	roomManager := sfu.NewRoomManager(redisRepo)
	handler := wsTransport.NewHandler(roomManager)

	http.HandleFunc("/ws", handler.Handle)

	log.Printf("SFU Signaling Server started on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
