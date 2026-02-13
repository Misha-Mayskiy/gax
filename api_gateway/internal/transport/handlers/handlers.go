package handlers

import (
	"api_gateway/internal/domain"
	"api_gateway/internal/service"
	user "api_gateway/pkg/api"
	"api_gateway/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type UserService interface {
	Create(ctx context.Context, uuid, email, userName, avatar, aboutMe string, friends []string) (*user.UserResponse, error)
	Update(ctx context.Context, uuid, email, userName, avatar, aboutMe string, friends []string) (*user.UserResponse, error)
	Delete(ctx context.Context, uuid string) (*user.DeleteUserResponse, error)
	Get(ctx context.Context, uuid string) (*user.UserResponse, error)
	SetOnline(ctx context.Context, uuid string, ttlSeconds int32) (*user.StatusResponse, error)
	IsOnline(ctx context.Context, uuid string) (*user.IsOnlineResponse, error)
	GetOnlineUsers(ctx context.Context) (*user.GetOnlineUsersResponse, error)
}

type HandlerFacade struct {
	ctx              context.Context
	authService      service.AuthService
	userService      UserService
	chatService      service.ChatService
	roomService      service.RoomService
	searchService    *service.SearchService
	mediaService     *service.MediaService
	roomWebSocketHub *RoomWebSocketHub
}

func NewHandlerFacade(ctx context.Context, userService UserService, chatService service.ChatService, searchService *service.SearchService, mediaService *service.MediaService, authService service.AuthService, roomService service.RoomService, roomWebSocketHub *RoomWebSocketHub) *HandlerFacade {
	return &HandlerFacade{
		ctx:              ctx,
		authService:      authService,
		userService:      userService,
		chatService:      chatService,
		roomService:      roomService,
		searchService:    searchService,
		mediaService:     mediaService,
		roomWebSocketHub: roomWebSocketHub,
	}
}

func (h *HandlerFacade) UserCreate(w http.ResponseWriter, r *http.Request) {
	var req domain.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		fmt.Printf("ERROR: UserCreate: invalid request body: %v\n", err)
		return
	}

	// Валидация
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if req.UserName == "" {
		http.Error(w, "user name is required", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.Create(h.ctx, req.UUID, req.Email, req.UserName, req.Avatar, req.AboutMe, req.Friends)
	if err != nil {
		fmt.Printf("ERROR: UserCreate failed: %v\n", err)

		// Определяем HTTP статус на основе типа ошибки
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "invalid argument") {
			statusCode = http.StatusBadRequest
		}

		http.Error(w, err.Error(), statusCode)
		return
	}

	// Создаем JWT токен
	jwtToken, err := utils.CreateToken(resp.Uuid)
	if err != nil {
		fmt.Printf("ERROR: UserCreate: failed to create token: %v\n", err)
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	// Устанавливаем куки
	http.SetCookie(w, &http.Cookie{
		Name:     "jwtToken",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   24 * 60 * 60, // 24 часа
	})

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Printf("ERROR: UserCreate: failed to encode response: %v\n", err)
	}
}

func (h *HandlerFacade) UserUpdate(w http.ResponseWriter, r *http.Request) {
	var req domain.UserUpdateRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cant read request body", http.StatusBadRequest)
		fmt.Printf("ERROR: UserUpdate cant read request body: %v\n", err)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "cant unmarshal request body", http.StatusBadRequest)
		fmt.Printf("ERROR: UserUpdate cant unmarshal request body: %v\n", err)
		return
	}

	// Теперь ожидаем 2 значения
	resp, err := h.userService.Update(h.ctx, req.Uuid, req.Email, req.UserName, req.Avatar, req.AboutMe, req.Friends)
	if err != nil {
		fmt.Printf("ERROR: UserUpdate failed: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		fmt.Printf("ERROR: UserUpdate cant marshal resp: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (h *HandlerFacade) UserDelete(w http.ResponseWriter, r *http.Request) {
	var req domain.UserDeleteRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cant read request body", http.StatusBadRequest)
		fmt.Printf("ERROR: UserDelete cant read request body: %v\n", err)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "cant unmarshal request body", http.StatusBadRequest)
		fmt.Printf("ERROR: UserDelete cant unmarshal request body: %v\n", err)
		return
	}

	// Ожидаем 2 значения
	resp, err := h.userService.Delete(h.ctx, req.Uuid)
	if err != nil {
		fmt.Printf("ERROR: UserDelete error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		fmt.Printf("ERROR: UserDelete cant marshal response: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (h *HandlerFacade) UserGetInfo(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "missing uuid", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.Get(r.Context(), uuid)
	if err != nil {
		fmt.Printf("ERROR: UserGetInfo failed: %v\n", err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}

		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Printf("ERROR: UserGetInfo: failed to encode response: %v\n", err)
	}
}

func (h *HandlerFacade) UserSetOnline(w http.ResponseWriter, r *http.Request) {
	var req domain.UserSetOnlineRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cant read request body", http.StatusBadRequest)
		fmt.Printf("ERROR: UserSetOnline cant read request body: %v\n", err)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "cant unmarshal request body", http.StatusBadRequest)
		fmt.Printf("ERROR: UserSetOnline cant unmarshal request body: %v\n", err)
		return
	}

	resp, err := h.userService.SetOnline(h.ctx, req.Uuid, req.TtlSeconds)
	if err != nil {
		fmt.Printf("ERROR: UserSetOnline failed: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		fmt.Printf("ERROR: UserSetOnline cant marshal response: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (h *HandlerFacade) UserIsOnline(w http.ResponseWriter, r *http.Request) {
	var req domain.UserIsOnlineRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cant read request body", http.StatusBadRequest)
		fmt.Printf("ERROR: UserIsOnline cant read request body: %v\n", err)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "cant unmarshal request body", http.StatusBadRequest)
		fmt.Printf("ERROR: UserIsOnline cant unmarshal request body: %v\n", err)
		return
	}

	resp, err := h.userService.IsOnline(h.ctx, req.Uuid)
	if err != nil {
		fmt.Printf("ERROR: UserIsOnline failed: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		fmt.Printf("ERROR: UserIsOnline cant marshal response: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (h *HandlerFacade) UserGetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	resp, err := h.userService.GetOnlineUsers(h.ctx)
	if err != nil {
		fmt.Printf("ERROR: UserGetOnlineUsers failed: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(&resp)
	if err != nil {
		fmt.Printf("ERROR: UserGetOnlineUsers cant marshal response: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (h *HandlerFacade) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	t := r.URL.Query().Get("type")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	highlight := r.URL.Query().Get("highlight") == "true"

	limit, _ := strconv.Atoi(defaultIfEmpty(limitStr, "20"))
	offset, _ := strconv.Atoi(defaultIfEmpty(offsetStr, "0"))

	resp := h.searchService.Search(r.Context(), q, t, limit, offset, highlight)
	if resp == nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func defaultIfEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
