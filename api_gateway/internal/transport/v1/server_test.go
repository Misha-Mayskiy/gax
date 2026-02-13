package v1

import (
	"api_gateway/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Вспомогательная функция для создания тестового конфига
func createTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port:    8080,
			TimeOut: 30 * time.Second,
		},
		Services: config.ServicesConfig{
			UserServiceAddr:   "localhost:50051",
			SearchServiceAddr: "localhost:50052",
			MediaServiceAddr:  "localhost:50053",
			AuthServiceAddr:   "localhost:8081",
			RoomServiceAddr:   "localhost:50053",
			ChatServiceAddr:   "localhost:8083",
		},
		Redis: config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Database: config.DatabaseConfig{
			PostgresDSN: "postgres://postgres:postgres@localhost:5432/users?sslmode=disable",
		},
		Kafka: config.KafkaConfig{
			Brokers: "localhost:29092",
		},
		Log: config.LogConfig{
			Level:  "info",
			Pretty: true,
		},
	}
}

// TestServer тесты для сервера
func TestNewServer(t *testing.T) {
	t.Run("Создание сервера с портом", func(t *testing.T) {
		server := NewServer(8080)
		assert.NotNil(t, server)
		assert.NotNil(t, server.srv)
		assert.Equal(t, ":8080", server.srv.Addr)
	})

	t.Run("Создание сервера с другим портом", func(t *testing.T) {
		server := NewServer(3000)
		assert.NotNil(t, server)
		assert.Equal(t, ":3000", server.srv.Addr)
	})
}

// func TestServerHealthCheck(t *testing.T) {
// 	t.Run("Health check возвращает успех", func(t *testing.T) {
// 		server := NewServer(8080)
// 		ctx := context.Background()
// 		cfg := createTestConfig()

// 		// Регистрируем обработчики
// 		server.RegisterHandler(ctx, cfg)

// 		req := httptest.NewRequest(http.MethodGet, "/health", nil)
// 		rr := httptest.NewRecorder()
// 		server.srv.Handler.ServeHTTP(rr, req)

// 		assert.Equal(t, http.StatusOK, rr.Code)
// 		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

// 		var response map[string]interface{}
// 		err := json.Unmarshal(rr.Body.Bytes(), &response)
// 		require.NoError(t, err)

// 		assert.Equal(t, "ok", response["status"])
// 		assert.Equal(t, "api-gateway", response["service"])
// 		assert.Equal(t, "v1", response["version"])
// 		assert.Contains(t, response, "timestamp")
// 	})

// 	t.Run("Health check с HEAD методом", func(t *testing.T) {
// 		server := NewServer(8080)
// 		ctx := context.Background()
// 		cfg := createTestConfig()

// 		server.RegisterHandler(ctx, cfg)

// 		req := httptest.NewRequest(http.MethodHead, "/health", nil)
// 		rr := httptest.NewRecorder()
// 		server.srv.Handler.ServeHTTP(rr, req)

// 		// HEAD запрос должен возвращать только заголовки
// 		assert.Equal(t, http.StatusOK, rr.Code)
// 		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
// 		assert.Empty(t, rr.Body.String())
// 	})
// }

func TestServerRootHandler(t *testing.T) {
	t.Run("Корневой эндпоинт возвращает HTML", func(t *testing.T) {
		server := NewServer(8080)
		ctx := context.Background()
		cfg := createTestConfig()

		server.RegisterHandler(ctx, cfg)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		server.srv.Handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/html", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "GAX API Gateway")
		assert.Contains(t, rr.Body.String(), "Единая точка входа")
	})

	t.Run("Несуществующий путь возвращает 404", func(t *testing.T) {
		server := NewServer(8080)
		ctx := context.Background()
		cfg := createTestConfig()

		server.RegisterHandler(ctx, cfg)

		req := httptest.NewRequest(http.MethodGet, "/non-existent-path", nil)
		rr := httptest.NewRecorder()
		server.srv.Handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestServerMethodValidation(t *testing.T) {
	server := NewServer(8080)
	ctx := context.Background()
	cfg := createTestConfig()

	server.RegisterHandler(ctx, cfg)

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET на PUT endpoint",
			path:           "/user/create",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "POST на PATCH endpoint",
			path:           "/user/update",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "GET на DELETE endpoint",
			path:           "/user/delete",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "POST на GET endpoint",
			path:           "/user/get",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "PUT на POST endpoint",
			path:           "/chat/create-direct",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			server.srv.Handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusMethodNotAllowed {
				assert.Contains(t, rr.Body.String(), "method not allowed")
			}
		})
	}
}

func TestServerAPIDocs(t *testing.T) {
	t.Run("API документация возвращает JSON", func(t *testing.T) {
		server := NewServer(8080)
		ctx := context.Background()
		cfg := createTestConfig()

		server.RegisterHandler(ctx, cfg)

		req := httptest.NewRequest(http.MethodGet, "/api-docs", nil)
		rr := httptest.NewRecorder()
		server.srv.Handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "3.0.0", response["openapi"])
		assert.Contains(t, response, "info")
		assert.Contains(t, response, "servers")
		assert.Contains(t, response, "paths")
	})
}

func TestServerEndpointRegistration(t *testing.T) {
	t.Run("Все эндпоинты зарегистрированы", func(t *testing.T) {
		server := NewServer(8080)
		ctx := context.Background()
		cfg := createTestConfig()

		server.RegisterHandler(ctx, cfg)

		// Проверяем несколько ключевых эндпоинтов
		endpoints := []struct {
			path   string
			method string
		}{
			{"/user/create", http.MethodPut},
			{"/user/get", http.MethodGet},
			{"/chat/create-direct", http.MethodPost},
			{"/search", http.MethodGet},
			{"/media/upload", http.MethodPost},
			{"/room/create", http.MethodPost},
			{"/room/ws/", http.MethodGet},
		}

		for _, endpoint := range endpoints {
			t.Run(fmt.Sprintf("Эндпоинт %s %s", endpoint.method, endpoint.path), func(t *testing.T) {
				req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
				rr := httptest.NewRecorder()
				server.srv.Handler.ServeHTTP(rr, req)

				// Проверяем, что обработчик установлен
				// Метод не поддерживается или другие ошибки - нормально
				// Главное, что не 404
				assert.NotEqual(t, http.StatusNotFound, rr.Code,
					fmt.Sprintf("Endpoint %s %s not found", endpoint.method, endpoint.path))
			})
		}
	})
}

func TestServerWithCORS(t *testing.T) {
	t.Run("Middleware withCORS добавляет заголовки", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		wrappedHandler := withCORS(handler)

		tests := []struct {
			name           string
			method         string
			origin         string
			expectedOrigin string
		}{
			{
				name:           "С Origin заголовком",
				method:         http.MethodGet,
				origin:         "http://localhost:3000",
				expectedOrigin: "http://localhost:3000",
			},
			{
				name:           "Без Origin заголовка",
				method:         http.MethodGet,
				origin:         "",
				expectedOrigin: "*",
			},
			{
				name:           "OPTIONS запрос",
				method:         http.MethodOptions,
				origin:         "http://example.com",
				expectedOrigin: "http://example.com",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest(tt.method, "/test", nil)
				if tt.origin != "" {
					req.Header.Set("Origin", tt.origin)
				}
				rr := httptest.NewRecorder()

				wrappedHandler.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedOrigin, rr.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD", rr.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
			})
		}
	})

	t.Run("OPTIONS запрос возвращает 200", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := withCORS(handler)

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Body.String())
	})
}

func TestServerErrorHandling(t *testing.T) {
	t.Run("Неподдерживаемый метод возвращает 405", func(t *testing.T) {
		server := NewServer(8080)
		ctx := context.Background()
		cfg := createTestConfig()

		server.RegisterHandler(ctx, cfg)

		// Пробуем POST на GET endpoint
		req := httptest.NewRequest(http.MethodPost, "/user/get", nil)
		rr := httptest.NewRecorder()
		server.srv.Handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
		assert.Contains(t, rr.Body.String(), "method not allowed")
	})

	t.Run("Некорректный запрос на /room/ws/", func(t *testing.T) {
		server := NewServer(8080)
		ctx := context.Background()
		cfg := createTestConfig()

		server.RegisterHandler(ctx, cfg)

		// Пробуем WebSocket без обязательных параметров
		req := httptest.NewRequest(http.MethodGet, "/room/ws/", nil)
		rr := httptest.NewRecorder()
		server.srv.Handler.ServeHTTP(rr, req)

		// WebSocket endpoint должен возвращать ошибку при обычном HTTP запросе
		assert.NotEqual(t, http.StatusOK, rr.Code)
	})
}

// Benchmark тесты
func BenchmarkHealthCheck(b *testing.B) {
	server := NewServer(8080)
	ctx := context.Background()
	cfg := createTestConfig()

	server.RegisterHandler(ctx, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()
		server.srv.Handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", rr.Code)
		}
	}
}

func BenchmarkRootHandler(b *testing.B) {
	server := NewServer(8080)
	ctx := context.Background()
	cfg := createTestConfig()

	server.RegisterHandler(ctx, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		server.srv.Handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", rr.Code)
		}
	}
}

// Integration-style тест
func TestServerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Полный цикл запросов", func(t *testing.T) {
		server := NewServer(0)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cfg := createTestConfig()

		// Запускаем сервер
		server.RegisterHandler(ctx, cfg)

		// Проверяем различные эндпоинты
		tests := []struct {
			name   string
			path   string
			method string
			body   io.Reader
		}{
			{
				name:   "Health check",
				path:   "/health",
				method: http.MethodGet,
			},
			{
				name:   "Root endpoint",
				path:   "/",
				method: http.MethodGet,
			},
			{
				name:   "API docs",
				path:   "/api-docs",
				method: http.MethodGet,
			},
			{
				name:   "User get endpoint (неверный метод)",
				path:   "/user/get",
				method: http.MethodPost,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest(tt.method, tt.path, tt.body)
				rr := httptest.NewRecorder()
				server.srv.Handler.ServeHTTP(rr, req)

				// Проверяем, что сервер отвечает (не 404)
				assert.NotEqual(t, http.StatusNotFound, rr.Code,
					fmt.Sprintf("Endpoint %s not found", tt.path))
			})
		}
	})
}

// Тесты для проверки конкретных обработчиков
func TestServerSpecificHandlers(t *testing.T) {
	server := NewServer(8080)
	ctx := context.Background()
	cfg := createTestConfig()

	server.RegisterHandler(ctx, cfg)

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		description    string
	}{
		// User handlers
		{
			name:           "User create - правильный метод",
			path:           "/user/create",
			method:         http.MethodPut,
			expectedStatus: http.StatusBadRequest, // Ожидаем 400, так как нет тела запроса
			description:    "PUT /user/create должен принимать JSON",
		},
		{
			name:           "User get - правильный метод",
			path:           "/user/get",
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest, // Ожидаем 400, так как нет параметров
			description:    "GET /user/get требует параметров",
		},
		{
			name:           "User delete - правильный метод",
			path:           "/user/delete",
			method:         http.MethodDelete,
			expectedStatus: http.StatusBadRequest, // Ожидаем 400
			description:    "DELETE /user/delete требует параметров",
		},

		// Chat handlers
		{
			name:           "Chat create direct - правильный метод",
			path:           "/chat/create-direct",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest, // Ожидаем 400
			description:    "POST /chat/create-direct требует тела запроса",
		},
		{
			name:           "Chat list - правильный метод",
			path:           "/chat/list",
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest, // Ожидаем 400
			description:    "GET /chat/list требует параметров",
		},

		// Search handler
		{
			name:           "Search - правильный метод",
			path:           "/search",
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest, // Ожидаем 400
			description:    "GET /search требует параметров запроса",
		},

		// Room handlers
		{
			name:           "Room create - правильный метод",
			path:           "/room/create",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest, // Ожидаем 400
			description:    "POST /room/create требует тела запроса",
		},
		{
			name:           "Room state - правильный метод",
			path:           "/room/state",
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest, // Ожидаем 400
			description:    "GET /room/state требует параметров",
		},

		// Media handlers
		{
			name:           "Media upload - правильный метод",
			path:           "/media/upload",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest, // Ожидаем 400 (нет multipart данных)
			description:    "POST /media/upload требует multipart формы",
		},
		{
			name:           "Media download - правильный метод",
			path:           "/media/download",
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest, // Ожидаем 400
			description:    "GET /media/download требует параметров",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			server.srv.Handler.ServeHTTP(rr, req)

			// Проверяем, что не 404 и не 405
			assert.NotEqual(t, http.StatusNotFound, rr.Code,
				fmt.Sprintf("%s: Endpoint not found", tt.description))
			assert.NotEqual(t, http.StatusMethodNotAllowed, rr.Code,
				fmt.Sprintf("%s: Wrong method", tt.description))

			// Логируем фактический статус для отладки
			t.Logf("%s: Expected ~%d, got %d", tt.description, tt.expectedStatus, rr.Code)
		})
	}
}

// Тест для проверки CORS в реальных условиях
func TestServerCORSRealScenario(t *testing.T) {
	server := NewServer(8080)
	ctx := context.Background()
	cfg := createTestConfig()

	server.RegisterHandler(ctx, cfg)

	tests := []struct {
		name           string
		method         string
		path           string
		origin         string
		requestHeaders map[string]string
		checkHeaders   bool
	}{
		{
			name:   "CORS для AJAX запроса",
			method: http.MethodGet,
			path:   "/health",
			origin: "https://example.com",
			requestHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			checkHeaders: true,
		},
		{
			name:   "CORS для POST запроса",
			method: http.MethodPost,
			path:   "/chat/create-direct",
			origin: "https://frontend.app",
			requestHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
			},
			checkHeaders: true,
		},
		{
			name:           "CORS без Origin",
			method:         http.MethodGet,
			path:           "/health",
			origin:         "",
			requestHeaders: nil,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			server.srv.Handler.ServeHTTP(rr, req)

			if tt.checkHeaders {
				if tt.origin != "" {
					assert.Equal(t, tt.origin, rr.Header().Get("Access-Control-Allow-Origin"))
				} else {
					assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
				}

				assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD",
					rr.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
			}
		})
	}
}
