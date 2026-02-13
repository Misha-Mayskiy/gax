package v1

import (
	"api_gateway/internal/config"
	"api_gateway/internal/service"
	"api_gateway/internal/transport/handlers"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultHeaderTimeout = 5 * time.Second
)

type Server struct {
	srv *http.Server
}

func NewServer(port int) *Server {
	s := http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           nil,
		ReadHeaderTimeout: defaultHeaderTimeout,
	}
	return &Server{
		srv: &s,
	}
}

func (s *Server) RegisterHandler(ctx context.Context, cfg *config.Config) {
	userService, err := service.NewUserService(cfg.Services.UserServiceAddr)
	if err != nil {
		log.Fatalf("Failed to create user service: %v", err)
	}

	go func() {
		<-ctx.Done()
		userService.Close()
	}()

	authService := service.NewAuthService(cfg.Services.AuthServiceAddr)
	roomService := service.NewRoomService(cfg.Services.RoomServiceAddr)
	chatService := service.NewChatService(cfg.Services.ChatServiceAddr)
	searchService := service.NewSearchService(cfg.Services.SearchServiceAddr)
	mediaService := service.NewMediaService(cfg.Services.MediaServiceAddr)
	hub := handlers.NewRoomWebSocketHub()

	go hub.Run()
	handler := handlers.NewHandlerFacade(ctx, userService, chatService, searchService, mediaService, authService, roomService, hub)
	mux := http.NewServeMux()
	s.srv.Handler = withCORS(mux)

	// ========== AUTH HANDLERS ==========
	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.Register(w, r)
	})
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.Login(w, r)
	})
	mux.HandleFunc("/auth/password_change", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.PasswordChange(w, r)
	})

	// ========== USER HANDLERS ==========
	mux.HandleFunc("/user/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.UserCreate(w, r)
	})
	mux.HandleFunc("/user/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.UserUpdate(w, r)
	})
	mux.HandleFunc("/user/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.UserDelete(w, r)
	})
	mux.HandleFunc("/user/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.UserGetInfo(w, r)
	})
	mux.HandleFunc("/user/set_online", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.UserSetOnline(w, r)
	})
	mux.HandleFunc("/user/is_online", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.UserIsOnline(w, r)
	})
	mux.HandleFunc("/user/get_online_users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.UserGetOnlineUsers(w, r)
	})

	// ========== CHAT HANDLERS ==========

	// Создание директ-чата между двумя пользователями
	mux.HandleFunc("/chat/create-direct", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.CreateDirectChat(w, r)
	})

	// Создание группового чата
	mux.HandleFunc("/chat/create-group", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.CreateGroupChat(w, r)
	})

	// Обновление группового чата (название, участники)
	mux.HandleFunc("/chat/update-group", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.UpdateGroupChat(w, r)
	})

	// Получение информации о чате
	mux.HandleFunc("/chat/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.GetChat(w, r)
	})

	// Список чатов пользователя
	mux.HandleFunc("/chat/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.ListChats(w, r)
	})

	// ========== MESSAGE HANDLERS ==========

	// Отправка сообщения
	mux.HandleFunc("/chat/message/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.SendMessage(w, r)
	})

	// Обновление сообщения
	mux.HandleFunc("/chat/message/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.UpdateMessage(w, r)
	})

	// Удаление сообщения (мягкое или полное)
	mux.HandleFunc("/chat/message/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.DeleteMessage(w, r)
	})

	// Список сообщений в чате
	mux.HandleFunc("/chat/messages/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.ListMessages(w, r)
	})

	// Отметка сообщения как прочитанного
	mux.HandleFunc("/chat/message/mark-read", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.MarkRead(w, r)
	})

	// Сохранение/удаление сообщения из избранного
	mux.HandleFunc("/chat/message/toggle-saved", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.ToggleSaved(w, r)
	})

	// Список сохраненных сообщений пользователя
	mux.HandleFunc("/chat/messages/saved", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.ListSaved(w, r)
	})

	// Список прочитанных сообщений пользователя в чате
	mux.HandleFunc("/chat/messages/read", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.ListReadMessages(w, r)
	})
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.Search(w, r)
	})
	// ========== ROOM HANDLERS ==========
	mux.HandleFunc("/room/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.CreateRoom(w, r)
	})

	mux.HandleFunc("/room/join", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.JoinRoom(w, r)
	})

	mux.HandleFunc("/room/playback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.SetPlayback(w, r)
	})

	mux.HandleFunc("/room/state", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.GetState(w, r)
	})

	// ========== WEBSOCKET ROOM HANDLERS ==========
	mux.HandleFunc("/room/ws/", func(w http.ResponseWriter, r *http.Request) {
		handler.RoomWebSocket(w, r)
	})

	// ========== HEALTH & INFO ==========

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"service":   "api-gateway",
			"timestamp": time.Now().Unix(),
			"version":   "v1",
		})
	})

	// API documentation root
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>GAX API Gateway</title>
				<style>
					body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
					h1 { color: #333; }
					h2 { color: #555; margin-top: 30px; }
					.endpoint { background: #f5f5f5; padding: 10px 15px; margin: 10px 0; border-radius: 5px; }
					.method { display: inline-block; padding: 3px 8px; border-radius: 3px; font-weight: bold; margin-right: 10px; }
					.get { background: #61affe; color: white; }
					.post { background: #49cc90; color: white; }
					.put { background: #fca130; color: white; }
					.patch { background: #50e3c2; color: white; }
					.delete { background: #f93e3e; color: white; }
					.path { font-family: monospace; color: #333; }
				</style>
			</head>
			<body>
				<h1>🚀 GAX API Gateway</h1>
				<p>Единая точка входа для всех микросервисов GAX</p>
				
				<h2>📊 Health Check</h2>
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="path">/health</span>
					<p>Проверка состояния сервиса</p>
				</div>
				
				<h2>👤 User Service</h2>
				<div class="endpoint">
					<span class="method put">PUT</span>
					<span class="path">/user/create</span>
					<p>Создание нового пользователя</p>
				</div>
				<div class="endpoint">
					<span class="method patch">PATCH</span>
					<span class="path">/user/update</span>
					<p>Обновление данных пользователя</p>
				</div>
				<div class="endpoint">
					<span class="method delete">DELETE</span>
					<span class="path">/user/delete</span>
					<p>Удаление пользователя</p>
				</div>
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="path">/user/get</span>
					<p>Получение информации о пользователе</p>
				</div>
				<div class="endpoint">
					<span class="method post">POST</span>
					<span class="path">/user/set_online</span>
					<p>Установить пользователя онлайн</p>
				</div>
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="path">/user/is_online</span>
					<p>Проверить онлайн-статус пользователя</p>
				</div>
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="path">/user/get_online_users</span>
					<p>Получить список онлайн-пользователей</p>
				</div>
				
				<h2>💬 Chat Service</h2>
				<div class="endpoint">
					<span class="method post">POST</span>
					<span class="path">/chat/create-direct</span>
					<p>Создать директ-чат между двумя пользователями</p>
		
	s.srv.Handler = mux
 }		</div>
				<div class="endpoint">
					<span class="method post">POST</span>
					<span class="path">/chat/create-group</span>
					<p>Создать групповой чат</p>
				</div>
				<div class="endpoint">
					<span class="method patch">PATCH</span>
					<span class="path">/chat/update-group</span>
					<p>Обновить групповой чат (название, участники)</p>
				</div>
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="path">/chat/get</span>
					<p>Получить информацию о чате</p>
				</div>
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="path">/chat/list</span>
					<p>Список чатов пользователя</p>
				</div>
				
				<h2>💌 Message Service</h2>
				<div class="endpoint">
					<span class="method post">POST</span>
					<span class="path">/chat/message/send</span>
					<p>Отправить сообщение в чат</p>
				</div>
				<div class="endpoint">
					<span class="method patch">PATCH</span>
					<span class="path">/chat/message/update</span>
					<p>Обновить сообщение</p>
				</div>
				<div class="endpoint">
					<span class="method delete">DELETE</span>
					<span class="path">/chat/message/delete</span>
					<p>Удалить сообщение</p>
				</div>
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="path">/chat/messages/list</span>
					<p>Список сообщений в чате</p>
				</div>
				<div class="endpoint">
					<span class="method post">POST</span>
					<span class="path">/chat/message/mark-read</span>
					<p>Отметить сообщение как прочитанное</p>
				</div>
				<div class="endpoint">
					<span class="method post">POST</span>
					<span class="path">/chat/message/toggle-saved</span>
					<p>Сохранить/удалить из избранного</p>
				</div>
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="path">/chat/messages/saved</span>
					<p>Список сохраненных сообщений</p>
				</div>
				<div class="endpoint">
					<span class="method get">GET</span>
					<span class="path">/chat/messages/read</span>
					<p>Список прочитанных сообщений</p>
				</div>
				
				<p><em>Для тестирования API используйте curl, Postman или Swagger UI</em></p>
			</body>
			</html>
		`))
	})

	// Swagger/OpenAPI документация (опционально)
	mux.HandleFunc("/api-docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		docs := map[string]interface{}{
			"openapi": "3.0.0",
			"info": map[string]interface{}{
				"title":       "GAX API Gateway",
				"description": "Единая точка входа для микросервисов GAX",
				"version":     "1.0.0",
			},
			"servers": []map[string]interface{}{
				{
					"url":         "http://localhost:8080",
					"description": "Development server",
				},
			},
			"paths": map[string]interface{}{
				// Можно добавить полную OpenAPI спецификацию
			},
		}
		json.NewEncoder(w).Encode(docs)
	})
	// ========== MEDIA HANDLERS ==========
	mux.HandleFunc("/media/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.UploadFile(w, r)
	})

	mux.HandleFunc("/media/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.DownloadFile(w, r)
	})

	mux.HandleFunc("/media/meta", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.GetFileMeta(w, r)
	})

	mux.HandleFunc("/media/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.DeleteFile(w, r)
	})

	mux.HandleFunc("/media/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.ListUserFiles(w, r)
	})

	s.srv.Handler = withCORS(mux)
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем все источники для разработки
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Обрабатываем предварительный запрос OPTIONS
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
