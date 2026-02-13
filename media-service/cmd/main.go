package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"media-service/internal/kafka"
	"media-service/internal/repository"
	"media-service/internal/service"
	"media-service/internal/transport"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	// Конфиг
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	dbURL := os.Getenv("DB_URL")
	minioEp := os.Getenv("MINIO_ENDPOINT")
	minioKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSec := os.Getenv("MINIO_SECRET_KEY")
	kafkaAddr := os.Getenv("KAFKA_BROKER") // например "kafka:9092"

	// Kafka (опционально)
	var kafkaProducer *kafka.Producer
	if kafkaAddr != "" {
		kafkaProducer = kafka.NewProducer([]string{kafkaAddr}, "file.events")
		defer kafkaProducer.Close()
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	pgRepo := repository.NewPgRepo(db)
	minioRepo, err := repository.NewMinioRepo(minioEp, minioKey, minioSec)
	if err != nil {
		log.Fatal(err)
	}

	svc := service.NewMediaService(pgRepo, minioRepo, kafkaProducer)
	h := transport.NewHandler(svc)

	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"service": "media-service",
			"version": "v1.0",
		})
	}).Methods("GET")

	r.HandleFunc("/upload", h.Upload).Methods("POST")
	r.HandleFunc("/download/{id}", h.Download).Methods("GET")
	r.HandleFunc("/media/delete/{id}", h.DeleteFile).Methods("DELETE")
	r.HandleFunc("/media/list", h.ListFiles).Methods("GET")
	r.HandleFunc("/media/{id}", h.GetFileMeta).Methods("GET")

	// Корневой endpoint
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "media-service",
			"version": "v1.0",
			"status":  "running",
			"endpoints": map[string]string{
				"POST /upload":              "Upload a file",
				"GET /download/{id}":        "Download a file",
				"GET /media/{id}":           "Get file metadata",
				"DELETE /media/delete/{id}": "Delete a file",
				"GET /media/list":           "List user files",
				"GET /health":               "Health check",
			},
		})
	}).Methods("GET")

	log.Printf("Media Service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
