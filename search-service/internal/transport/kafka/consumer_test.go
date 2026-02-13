package kafka

import (
	"context"
	"encoding/json"
	"main/internal/domain"
	"main/internal/repository/es"
	"main/internal/repository/postgres"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockESServer поднимает локальный HTTP сервер, который притворяется Эластиком.
// Мы передаем его адрес в NewRepo.
func mockESServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Важный заголовок для клиента ES v8
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
}

func TestConsumer_HandleFileEvents(t *testing.T) {
	// Logger
	logger := zerolog.New(os.Stdout)
	pgRepo := &postgres.Repo{}

	t.Run("handleFileCreated", func(t *testing.T) {
		// Подготовка данных
		fileData := domain.FileIndex{
			ID: "file-1",
		}

		// Поднимаем мок-сервер ES
		ts := mockESServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Contains(t, r.URL.Path, "/files/_doc/file-1")

			// Проверяем, что дата обновления проставилась
			var body domain.FileIndex
			_ = json.NewDecoder(r.Body).Decode(&body)
			assert.NotZero(t, body.UpdatedAt)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"result": "created"}`))
		})
		defer ts.Close()

		// Инициализируем ES repo с адресом нашего мок-сервера
		esRepo, err := es.NewRepo(ts.URL, logger)
		require.NoError(t, err)

		// Создаем консьюмер (Reader нам не нужен для тестирования методов handle*)
		c := &Consumer{
			esRepo: esRepo,
			pgRepo: pgRepo,
			log:    logger,
		}

		// Вызываем приватный метод напрямую (мы в пакете kafka, так что имеем доступ к c.handleFileCreated)
		c.handleFileCreated(context.Background(), fileData)
	})

	t.Run("handleFileUpdated", func(t *testing.T) {
		fileData := domain.FileIndex{
			ID:        "file-2",
			UpdatedAt: 100, // Старое время
		}

		ts := mockESServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			var body domain.FileIndex
			_ = json.NewDecoder(r.Body).Decode(&body)

			// Проверяем, что метод обновил время
			assert.Greater(t, body.UpdatedAt, int64(100))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result": "updated"}`))
		})
		defer ts.Close()

		esRepo, err := es.NewRepo(ts.URL, logger)
		require.NoError(t, err)

		c := &Consumer{esRepo: esRepo, pgRepo: pgRepo, log: logger}

		c.handleFileUpdated(context.Background(), fileData)
	})

	t.Run("handleFileDeleted", func(t *testing.T) {
		// Тут ожидается payload map[string]interface{} с полем id
		payload := map[string]interface{}{
			"id": "file-3",
		}

		// В коде consumer.go логика удаления пока закомментирована (TODO),
		// но тест подготовим. Если ты раскомментируешь вызов c.esRepo.DeleteFile,
		// то нужно будет проверить DELETE запрос.

		// Сейчас метод просто логирует. Проверим, что не падает.
		ts := mockESServer(t, func(w http.ResponseWriter, r *http.Request) {
			// Если логика будет включена:
			// assert.Equal(t, "DELETE", r.Method)
			// assert.Contains(t, r.URL.Path, "file-3")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result": "deleted"}`))
		})
		defer ts.Close()

		esRepo, err := es.NewRepo(ts.URL, logger)
		require.NoError(t, err)

		c := &Consumer{esRepo: esRepo, pgRepo: pgRepo, log: logger}

		c.handleFileDeleted(context.Background(), payload)
	})
}
