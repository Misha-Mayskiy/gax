package es

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"main/internal/domain"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTransport позволяет перехватывать HTTP-запросы клиента Elastic
type mockTransport struct {
	RoundTripFn func(req *http.Request) (*http.Response, error)
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripFn != nil {
		return t.RoundTripFn(req)
	}
	// Если функция не задана, возвращаем 500
	return &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader(`{"error": "mock handler not defined"}`)),
		Header:     make(http.Header),
	}, nil
}

// helper для создания репозитория с мокнутым клиентом
func newTestRepo(t *testing.T, handler func(req *http.Request) (*http.Response, error)) *Repo {
	mockTr := &mockTransport{
		RoundTripFn: handler,
	}

	cfg := elasticsearch.Config{
		Transport: mockTr,
		// Отключаем попытки узнать информацию о кластере при старте, чтобы не усложнять моки
		DiscoverNodesOnStart: false,
	}

	client, err := elasticsearch.NewClient(cfg)
	require.NoError(t, err)

	return &Repo{
		client:    client,
		userIndex: "users",
		chatIndex: "chats",
		fileIndex: "files",
		msgIndex:  "messages",
	}
}

// helper для создания успешного ответа от ES
func mockResponse(statusCode int, body string) (*http.Response, error) {
	header := make(http.Header)
	// ВАЖНО: Этот заголовок нужен клиенту go-elasticsearch v8, чтобы признать ответ валидным
	header.Set("X-Elastic-Product", "Elasticsearch")
	header.Set("Content-Type", "application/json")

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     header,
	}, nil
}

func TestRepo_EnsureIndices(t *testing.T) {
	requestCount := 0

	repo := newTestRepo(t, func(req *http.Request) (*http.Response, error) {
		requestCount++
		// Проверяем, что метод PUT
		assert.Equal(t, "PUT", req.Method)
		return mockResponse(200, `{"acknowledged": true, "index": "test_index"}`)
	})

	err := repo.EnsureIndices(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 4, requestCount, "Should create 4 indices")
}

func TestRepo_IndexMethods(t *testing.T) {
	msg := domain.Message{
		ID:       "msg-1",
		ChatID:   "chat-1",
		Text:     "Hello world",
		AuthorID: "user-1",
	}

	repo := newTestRepo(t, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "PUT", req.Method)
		// Проверяем, что URL содержит правильный индекс и ID
		assert.Contains(t, req.URL.Path, "/messages/_doc/msg-1")

		// Проверяем тело запроса
		var body map[string]interface{}
		_ = json.NewDecoder(req.Body).Decode(&body)
		assert.Equal(t, "msg-1", body["id"])
		assert.Equal(t, "Hello world", body["text"])

		return mockResponse(201, `{"_id": "msg-1", "result": "created"}`)
	})

	err := repo.IndexMessage(context.Background(), msg)
	assert.NoError(t, err)
}

func TestRepo_DeleteFile(t *testing.T) {
	fileID := "file-123"

	repo := newTestRepo(t, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "DELETE", req.Method)
		assert.Contains(t, req.URL.Path, "/files/_doc/"+fileID)
		return mockResponse(200, `{"result": "deleted"}`)
	})

	err := repo.DeleteFile(context.Background(), fileID)
	assert.NoError(t, err)
}

func TestRepo_Search(t *testing.T) {
	// Подготавливаем JSON ответ от ElasticSearch
	esResponse := `{
		"hits": {
			"total": { "value": 2, "relation": "eq" },
			"hits": [
				{
					"_index": "users",
					"_source": {
						"id": "u1",
						"username": "alice",
						"email": "alice@example.com"
					},
					"highlight": {
						"username": ["<b>alice</b>"]
					}
				},
				{
					"_index": "messages",
					"_source": {
						"id": "m1",
						"text": "Hello alice",
						"chat_id": "c1",
						"author_id": "u2"
					}
				}
			]
		}
	}`

	t.Run("Mixed Search", func(t *testing.T) {
		repo := newTestRepo(t, func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "POST", req.Method)
			assert.Contains(t, req.URL.Path, "_search")

			// Проверяем параметры поиска в теле
			buf := new(bytes.Buffer)
			buf.ReadFrom(req.Body)
			bodyStr := buf.String()

			assert.Contains(t, bodyStr, `"query":"alice"`)
			assert.Contains(t, bodyStr, `"from":0`)
			assert.Contains(t, bodyStr, `"size":10`)

			return mockResponse(200, esResponse)
		})

		query := domain.SearchQuery{
			Q:         "alice",
			Offset:    0,
			Limit:     10,
			Highlight: true,
		}

		results, err := repo.Search(context.Background(), query)
		require.NoError(t, err)
		require.Len(t, results, 2)

		// Проверка User
		userRes := results[0]
		assert.Equal(t, "u1", userRes.ID)
		assert.Equal(t, "user", userRes.Type)
		assert.Equal(t, "alice", userRes.Title)
		assert.Equal(t, "<b>alice</b>", userRes.Snippet)
		assert.Contains(t, userRes.ExtraIDs, "alice@example.com")

		// Проверка Message
		msgRes := results[1]
		assert.Equal(t, "m1", msgRes.ID)
		assert.Equal(t, "message", msgRes.Type)
		assert.Equal(t, "Hello alice", msgRes.Title)
		assert.Contains(t, msgRes.ExtraIDs, "c1")
	})

	t.Run("Specific Type Search", func(t *testing.T) {
		repo := newTestRepo(t, func(req *http.Request) (*http.Response, error) {
			// Проверяем, что запрос идет только к индексу файлов
			assert.Contains(t, req.URL.Path, "/files/")
			return mockResponse(200, `{"hits": {"hits": []}}`)
		})

		query := domain.SearchQuery{
			Q:    "doc",
			Type: "file",
		}
		_, err := repo.Search(context.Background(), query)
		assert.NoError(t, err)
	})
}

func TestRepo_Search_Error(t *testing.T) {
	// Имитация сетевой ошибки
	repoNetworkError := newTestRepo(t, func(req *http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	})

	_, err := repoNetworkError.Search(context.Background(), domain.SearchQuery{Q: "test"})
	assert.Error(t, err)
}
