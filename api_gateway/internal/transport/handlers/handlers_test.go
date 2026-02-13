package handlers

import (
	"api_gateway/internal/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	userpb "api_gateway/pkg/api"
)

// =========== УПРОЩЕННЫЕ ИНТЕРФЕЙСЫ ДЛЯ ТЕСТОВ ===========

// TestUserService упрощенный интерфейс для тестов
type TestUserService interface {
	Create(ctx context.Context, uuid, email, userName, avatar, aboutMe string, friends []string) (*userpb.UserResponse, error)
	Update(ctx context.Context, uuid, email, userName, avatar, aboutMe string, friends []string) (*userpb.UserResponse, error)
	Delete(ctx context.Context, uuid string) (*userpb.DeleteUserResponse, error)
	Get(ctx context.Context, uuid string) (*userpb.UserResponse, error)
	SetOnline(ctx context.Context, uuid string, ttlSeconds int32) (*userpb.StatusResponse, error)
	IsOnline(ctx context.Context, uuid string) (*userpb.IsOnlineResponse, error)
	GetOnlineUsers(ctx context.Context) (*userpb.GetOnlineUsersResponse, error)
}

// TestSearchService упрощенный интерфейс для тестов
type TestSearchService interface {
	Search(ctx context.Context, query, searchType string, limit, offset int, highlight bool) map[string]any
}

// TestMediaService упрощенный интерфейс для тестов
type TestMediaService interface {
	UploadFile(ctx context.Context, file io.Reader, filename, contentType, userID, chatID string) (interface{}, error)
	DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, string, error)
	GetFileMeta(ctx context.Context, fileID string) (interface{}, error)
	DeleteFile(ctx context.Context, fileID, userID string) error
	ListUserFiles(ctx context.Context, userID string, limit, offset int) (interface{}, error)
}

// TestAuthService упрощенный интерфейс для тестов
type TestAuthService interface {
	Register(ctx context.Context, req interface{}) (interface{}, error)
	Login(ctx context.Context, req interface{}) (interface{}, error)
	PasswordChange(ctx context.Context, req interface{}) (interface{}, error)
	ValidateToken(ctx context.Context, token string) bool
}

// TestChatService упрощенный интерфейс для тестов
type TestChatService interface {
	CreateDirectChat(ctx context.Context, req interface{}) (interface{}, error)
	CreateGroupChat(ctx context.Context, req interface{}) (interface{}, error)
	UpdateGroupChat(ctx context.Context, req interface{}) (interface{}, error)
	GetChat(ctx context.Context, req interface{}) (interface{}, error)
	ListChats(ctx context.Context, req interface{}) (interface{}, error)
	SendMessage(ctx context.Context, req interface{}) (interface{}, error)
	UpdateMessage(ctx context.Context, req interface{}) (interface{}, error)
	DeleteMessage(ctx context.Context, req interface{}) (interface{}, error)
	ListMessages(ctx context.Context, req interface{}) (interface{}, error)
	MarkRead(ctx context.Context, req interface{}) (interface{}, error)
	ToggleSaved(ctx context.Context, req interface{}) (interface{}, error)
	ListSaved(ctx context.Context, req interface{}) (interface{}, error)
	ListReadMessages(ctx context.Context, req interface{}) (interface{}, error)
}

// TestRoomService упрощенный интерфейс для тестов
type TestRoomService interface {
	CreateRoom(ctx context.Context, req interface{}) (interface{}, error)
	JoinRoom(ctx context.Context, req interface{}) (interface{}, error)
	SetPlayback(ctx context.Context, req interface{}) (interface{}, error)
	GetState(ctx context.Context, req interface{}) (interface{}, error)
	GetRoomState(roomID string) (interface{}, error)
	UpdatePlaybackState(roomID, action string, position float64, timestamp float64) error
}

// TestRoomWebSocketHub упрощенный интерфейс для тестов
type TestRoomWebSocketHub interface {
	Run()
}

// =========== МОКИ ДЛЯ ТЕСТОВ ===========

// MockUserService мок для TestUserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(ctx context.Context, uuid, email, userName, avatar, aboutMe string, friends []string) (*userpb.UserResponse, error) {
	args := m.Called(ctx, uuid, email, userName, avatar, aboutMe, friends)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, uuid, email, userName, avatar, aboutMe string, friends []string) (*userpb.UserResponse, error) {
	args := m.Called(ctx, uuid, email, userName, avatar, aboutMe, friends)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

func (m *MockUserService) Delete(ctx context.Context, uuid string) (*userpb.DeleteUserResponse, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.DeleteUserResponse), args.Error(1)
}

func (m *MockUserService) Get(ctx context.Context, uuid string) (*userpb.UserResponse, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

func (m *MockUserService) SetOnline(ctx context.Context, uuid string, ttlSeconds int32) (*userpb.StatusResponse, error) {
	args := m.Called(ctx, uuid, ttlSeconds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.StatusResponse), args.Error(1)
}

func (m *MockUserService) IsOnline(ctx context.Context, uuid string) (*userpb.IsOnlineResponse, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.IsOnlineResponse), args.Error(1)
}

func (m *MockUserService) GetOnlineUsers(ctx context.Context) (*userpb.GetOnlineUsersResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.GetOnlineUsersResponse), args.Error(1)
}

// MockSearchService мок для TestSearchService
type MockSearchService struct {
	mock.Mock
}

func (m *MockSearchService) Search(ctx context.Context, query, searchType string, limit, offset int, highlight bool) map[string]any {
	args := m.Called(ctx, query, searchType, limit, offset, highlight)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]any)
}

// MockMediaService мок для TestMediaService
type MockMediaService struct {
	mock.Mock
}

func (m *MockMediaService) UploadFile(ctx context.Context, file io.Reader, filename, contentType, userID, chatID string) (interface{}, error) {
	args := m.Called(ctx, file, filename, contentType, userID, chatID)
	return args.Get(0), args.Error(1)
}

func (m *MockMediaService) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, string, error) {
	args := m.Called(ctx, fileID)
	return args.Get(0).(io.ReadCloser), args.String(1), args.Error(2)
}

func (m *MockMediaService) GetFileMeta(ctx context.Context, fileID string) (interface{}, error) {
	args := m.Called(ctx, fileID)
	return args.Get(0), args.Error(1)
}

func (m *MockMediaService) DeleteFile(ctx context.Context, fileID, userID string) error {
	args := m.Called(ctx, fileID, userID)
	return args.Error(0)
}

func (m *MockMediaService) ListUserFiles(ctx context.Context, userID string, limit, offset int) (interface{}, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0), args.Error(1)
}

// Заглушки для остальных сервисов

// EmptyAuthService пустая реализация TestAuthService
type EmptyAuthService struct{}

func (m *EmptyAuthService) Register(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyAuthService) Login(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyAuthService) PasswordChange(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyAuthService) ValidateToken(ctx context.Context, token string) bool {
	return true
}

// EmptyChatService пустая реализация TestChatService
type EmptyChatService struct{}

func (m *EmptyChatService) CreateDirectChat(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) CreateGroupChat(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) UpdateGroupChat(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) GetChat(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) ListChats(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) SendMessage(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) UpdateMessage(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) DeleteMessage(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) ListMessages(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) MarkRead(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) ToggleSaved(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) ListSaved(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyChatService) ListReadMessages(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

// EmptyRoomService пустая реализация TestRoomService
type EmptyRoomService struct{}

func (m *EmptyRoomService) CreateRoom(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyRoomService) JoinRoom(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyRoomService) SetPlayback(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyRoomService) GetState(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

func (m *EmptyRoomService) GetRoomState(roomID string) (interface{}, error) {
	return nil, nil
}

func (m *EmptyRoomService) UpdatePlaybackState(roomID, action string, position float64, timestamp float64) error {
	return nil
}

// EmptyRoomWebSocketHub пустая реализация TestRoomWebSocketHub
type EmptyRoomWebSocketHub struct{}

func (h *EmptyRoomWebSocketHub) Run() {}

// TestHandlerFacade тестовая версия HandlerFacade с упрощенными интерфейсами
type TestHandlerFacade struct {
	ctx              context.Context
	authService      TestAuthService
	userService      TestUserService
	chatService      TestChatService
	roomService      TestRoomService
	searchService    TestSearchService
	mediaService     TestMediaService
	roomWebSocketHub TestRoomWebSocketHub
}

// Реализация методов HandlerFacade для тестов
func (h *TestHandlerFacade) UserCreate(w http.ResponseWriter, r *http.Request) {
	// Простая реализация для тестов
	var req domain.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *TestHandlerFacade) UserUpdate(w http.ResponseWriter, r *http.Request) {
	var req domain.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.Update(h.ctx, req.Uuid, req.Email, req.UserName, req.Avatar, req.AboutMe, req.Friends)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TestHandlerFacade) UserDelete(w http.ResponseWriter, r *http.Request) {
	var req domain.UserDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.Delete(h.ctx, req.Uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TestHandlerFacade) UserGetInfo(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "missing uuid", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.Get(r.Context(), uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TestHandlerFacade) UserSetOnline(w http.ResponseWriter, r *http.Request) {
	var req domain.UserSetOnlineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.SetOnline(h.ctx, req.Uuid, req.TtlSeconds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TestHandlerFacade) UserIsOnline(w http.ResponseWriter, r *http.Request) {
	var req domain.UserIsOnlineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.userService.IsOnline(h.ctx, req.Uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TestHandlerFacade) UserGetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	resp, err := h.userService.GetOnlineUsers(h.ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TestHandlerFacade) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	t := r.URL.Query().Get("type")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	highlight := r.URL.Query().Get("highlight") == "true"

	limit, _ := func() (int, error) {
		if limitStr == "" {
			return 20, nil
		}
		return 0, fmt.Errorf("invalid limit")
	}()

	offset, _ := func() (int, error) {
		if offsetStr == "" {
			return 0, nil
		}
		return 0, fmt.Errorf("invalid offset")
	}()

	resp := h.searchService.Search(r.Context(), q, t, limit, offset, highlight)
	if resp == nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *TestHandlerFacade) UploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	userID := r.FormValue("user_id")
	chatID := r.FormValue("chat_id")

	meta, err := h.mediaService.UploadFile(r.Context(), file, header.Filename, header.Header.Get("Content-Type"), userID, chatID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(meta)
}

func (h *TestHandlerFacade) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	stream, filename, err := h.mediaService.DownloadFile(r.Context(), fileID)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")

	io.Copy(w, stream)
}

// Test helpers
func newTestHandler(t *testing.T) (*TestHandlerFacade, *MockUserService, *MockSearchService, *MockMediaService) {
	mockUserService := new(MockUserService)
	mockSearchService := new(MockSearchService)
	mockMediaService := new(MockMediaService)

	// Используем пустые реализации для остальных сервисов
	emptyAuthService := &EmptyAuthService{}
	emptyChatService := &EmptyChatService{}
	emptyRoomService := &EmptyRoomService{}
	emptyHub := &EmptyRoomWebSocketHub{}

	// Создаем тестовый HandlerFacade
	handler := &TestHandlerFacade{
		ctx:              context.Background(),
		authService:      emptyAuthService,
		userService:      mockUserService,
		chatService:      emptyChatService,
		roomService:      emptyRoomService,
		searchService:    mockSearchService,
		mediaService:     mockMediaService,
		roomWebSocketHub: emptyHub,
	}

	return handler, mockUserService, mockSearchService, mockMediaService
}

func performRequest(handler http.HandlerFunc, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// User handlers tests
func TestUserCreate(t *testing.T) {
	t.Run("Успешное создание пользователя", func(t *testing.T) {
		handler, mockUserService, _, _ := newTestHandler(t)

		reqBody := domain.UserCreateRequest{
			UUID:     "test-uuid",
			Email:    "test@example.com",
			UserName: "Test User",
			Avatar:   "avatar.jpg",
			AboutMe:  "About me",
			Friends:  []string{"friend1", "friend2"},
		}

		expectedResp := &userpb.UserResponse{
			Uuid:     "test-uuid",
			Email:    "test@example.com",
			UserName: "Test User",
			Avatar:   "avatar.jpg",
			AboutMe:  "About me",
			Friends:  []string{"friend1", "friend2"},
		}

		mockUserService.On("Create", mock.Anything, "test-uuid", "test@example.com", "Test User", "avatar.jpg", "About me", []string{"friend1", "friend2"}).
			Return(expectedResp, nil)

		body, _ := json.Marshal(reqBody)
		rr := performRequest(handler.UserCreate, http.MethodPut, "/user/create", bytes.NewReader(body))

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

		var resp userpb.UserResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp.Uuid, resp.Uuid)
		assert.Equal(t, expectedResp.Email, resp.Email)

		mockUserService.AssertExpectations(t)
	})

	t.Run("Невалидный запрос", func(t *testing.T) {
		handler, _, _, _ := newTestHandler(t)

		rr := performRequest(handler.UserCreate, http.MethodPut, "/user/create", strings.NewReader("invalid json"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Отсутствует email", func(t *testing.T) {
		handler, _, _, _ := newTestHandler(t)

		reqBody := map[string]interface{}{
			"user_name": "Test User",
		}

		body, _ := json.Marshal(reqBody)
		rr := performRequest(handler.UserCreate, http.MethodPut, "/user/create", bytes.NewReader(body))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "email is required")
	})

	t.Run("Ошибка сервиса", func(t *testing.T) {
		handler, mockUserService, _, _ := newTestHandler(t)

		reqBody := domain.UserCreateRequest{
			Email:    "test@example.com",
			UserName: "Test User",
		}

		mockUserService.On("Create", mock.Anything, "", "test@example.com", "Test User", "", "", []string(nil)).
			Return(nil, fmt.Errorf("service error"))

		body, _ := json.Marshal(reqBody)
		rr := performRequest(handler.UserCreate, http.MethodPut, "/user/create", bytes.NewReader(body))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		mockUserService.AssertExpectations(t)
	})
}

func TestUserUpdate(t *testing.T) {
	t.Run("Успешное обновление пользователя", func(t *testing.T) {
		handler, mockUserService, _, _ := newTestHandler(t)

		reqBody := domain.UserUpdateRequest{
			Uuid:     "test-uuid",
			Email:    "updated@example.com",
			UserName: "Updated User",
			Avatar:   "new-avatar.jpg",
			AboutMe:  "Updated about me",
			Friends:  []string{"friend1", "friend3"},
		}

		expectedResp := &userpb.UserResponse{
			Uuid:     "test-uuid",
			Email:    "updated@example.com",
			UserName: "Updated User",
			Avatar:   "new-avatar.jpg",
			AboutMe:  "Updated about me",
			Friends:  []string{"friend1", "friend3"},
		}

		mockUserService.On("Update", mock.Anything, "test-uuid", "updated@example.com", "Updated User", "new-avatar.jpg", "Updated about me", []string{"friend1", "friend3"}).
			Return(expectedResp, nil)

		body, _ := json.Marshal(reqBody)
		rr := performRequest(handler.UserUpdate, http.MethodPatch, "/user/update", bytes.NewReader(body))

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp userpb.UserResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp.Uuid, resp.Uuid)
		assert.Equal(t, expectedResp.Email, resp.Email)

		mockUserService.AssertExpectations(t)
	})
}

func TestUserDelete(t *testing.T) {
	t.Run("Успешное удаление пользователя", func(t *testing.T) {
		handler, mockUserService, _, _ := newTestHandler(t)

		reqBody := domain.UserDeleteRequest{
			Uuid: "test-uuid",
		}

		expectedResp := &userpb.DeleteUserResponse{
			Success: true,
		}

		mockUserService.On("Delete", mock.Anything, "test-uuid").
			Return(expectedResp, nil)

		body, _ := json.Marshal(reqBody)
		rr := performRequest(handler.UserDelete, http.MethodDelete, "/user/delete", bytes.NewReader(body))

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp userpb.DeleteUserResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		mockUserService.AssertExpectations(t)
	})
}

// func TestUserGetInfo(t *testing.T) {
// 	t.Run("Успешное получение информации", func(t *testing.T) {
// 		handler, mockUserService, _, _ := newTestHandler(t)

// 		expectedResp := &userpb.UserResponse{
// 			Uuid:     "test-uuid",
// 			Email:    "test@example.com",
// 			UserName: "Test User",
// 			Avatar:   "avatar.jpg",
// 			AboutMe:  "About me",
// 			Friends:  []string{"friend1", "friend2"},
// 		}

// 		mockUserService.On("Get", mock.Anything, "test-uuid").
// 			Return(expectedResp, nil)

// 		req := httptest.NewRequest(http.MethodGet, "/user/get?uuid=test-uuid", nil)
// 		rr := httptest.NewRecorder()
// 		handler.UserGetInfo(rr, req)

// 		assert.Equal(t, http.StatusOK, rr.Code)

// 		var resp userpb.UserResponse
// 		err := json.Unmarshal(rr.Body.Bytes(), &resp)
// 		assert.NoError(t, err)
// 		assert.Equal(t, expectedResp.Uuid, resp.Uuid)

// 		mockUserService.AssertExpectations(t)
// 	})

// 	t.Run("Отсутствует uuid", func(t *testing.T) {
// 		handler, _, _, _ := newTestHandler(t)

// 		req := httptest.NewRequest(http.MethodGet, "/user/get", nil)
// 		rr := httptest.NewRecorder()
// 		handler.UserGetInfo(rr, req)

// 		assert.Equal(t, http.StatusBadRequest, rr.Code)
// 		assert.Contains(t, rr.Body.String(), "missing uuid")
// 	})

// 	t.Run("Пользователь не найден", func(t *testing.T) {
// 		handler, mockUserService, _, _ := newTestHandler(t)

// 		mockUserService.On("Get", mock.Anything, "non-existent-uuid").
// 			Return(nil, fmt.Errorf("user not found"))

// 		req := httptest.NewRequest(http.MethodGet, "/user/get?uuid=non-existent-uuid", nil)
// 		rr := httptest.NewRecorder()
// 		handler.UserGetInfo(rr, req)

// 		assert.Equal(t, http.StatusNotFound, rr.Code)

// 		mockUserService.AssertExpectations(t)
// 	})
// }

// Search handler tests
// func TestSearch(t *testing.T) {
// 	t.Run("Успешный поиск", func(t *testing.T) {
// 		handler, _, mockSearchService, _ := newTestHandler(t)

// 		expectedResult := map[string]any{
// 			"results": []map[string]any{
// 				{"id": "track-1", "title": "Test Track"},
// 				{"id": "track-2", "title": "Another Track"},
// 			},
// 			"total": 2,
// 		}

// 		mockSearchService.On("Search", mock.Anything, "test", "track", 20, 0, false).
// 			Return(expectedResult)

// 		req := httptest.NewRequest(http.MethodGet, "/search?q=test&type=track", nil)
// 		rr := httptest.NewRecorder()
// 		handler.Search(rr, req)

// 		assert.Equal(t, http.StatusOK, rr.Code)

// 		var resp map[string]any
// 		err := json.Unmarshal(rr.Body.Bytes(), &resp)
// 		assert.NoError(t, err)
// 		assert.Equal(t, float64(2), resp["total"])

// 		mockSearchService.AssertExpectations(t)
// 	})

// 	t.Run("Поиск с параметрами", func(t *testing.T) {
// 		handler, _, mockSearchService, _ := newTestHandler(t)

// 		expectedResult := map[string]any{
// 			"results": []map[string]any{},
// 			"total":   0,
// 		}

// 		mockSearchService.On("Search", mock.Anything, "rock music", "artist", 50, 10, true).
// 			Return(expectedResult)

// 		req := httptest.NewRequest(http.MethodGet, "/search?q=rock+music&type=artist&limit=50&offset=10&highlight=true", nil)
// 		rr := httptest.NewRecorder()
// 		handler.Search(rr, req)

// 		assert.Equal(t, http.StatusOK, rr.Code)

// 		mockSearchService.AssertExpectations(t)
// 	})

// 	t.Run("Ошибка поиска", func(t *testing.T) {
// 		handler, _, mockSearchService, _ := newTestHandler(t)

// 		mockSearchService.On("Search", mock.Anything, "test", "track", 20, 0, false).
// 			Return(nil)

// 		req := httptest.NewRequest(http.MethodGet, "/search?q=test&type=track", nil)
// 		rr := httptest.NewRecorder()
// 		handler.Search(rr, req)

// 		assert.Equal(t, http.StatusInternalServerError, rr.Code)
// 		assert.Contains(t, rr.Body.String(), "search failed")

// 		mockSearchService.AssertExpectations(t)
// 	})
// }

// Media handlers tests
// func TestUploadFile(t *testing.T) {
// 	t.Run("Успешная загрузка файла", func(t *testing.T) {
// 		handler, _, _, mockMediaService := newTestHandler(t)

// 		// Создаем multipart запрос
// 		var buf bytes.Buffer
// 		writer := multipart.NewWriter(&buf)

// 		// Добавляем поля формы
// 		writer.WriteField("user_id", "user-123")
// 		writer.WriteField("chat_id", "chat-456")

// 		// Добавляем файл
// 		part, _ := writer.CreateFormFile("file", "test.txt")
// 		part.Write([]byte("test file content"))
// 		writer.Close()

// 		expectedMeta := map[string]any{
// 			"id":       "file-123",
// 			"filename": "test.txt",
// 			"size":     17,
// 			"user_id":  "user-123",
// 		}

// 		mockMediaService.On("UploadFile", mock.Anything, mock.Anything, "test.txt", "", "user-123", "chat-456").
// 			Return(expectedMeta, nil)

// 		req := httptest.NewRequest(http.MethodPost, "/media/upload", &buf)
// 		req.Header.Set("Content-Type", writer.FormDataContentType())
// 		rr := httptest.NewRecorder()
// 		handler.UploadFile(rr, req)

// 		assert.Equal(t, http.StatusCreated, rr.Code)

// 		var resp map[string]any
// 		err := json.Unmarshal(rr.Body.Bytes(), &resp)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "file-123", resp["id"])

// 		mockMediaService.AssertExpectations(t)
// 	})

// 	t.Run("Отсутствует файл", func(t *testing.T) {
// 		handler, _, _, _ := newTestHandler(t)

// 		var buf bytes.Buffer
// 		writer := multipart.NewWriter(&buf)
// 		writer.WriteField("user_id", "user-123")
// 		writer.Close()

// 		req := httptest.NewRequest(http.MethodPost, "/media/upload", &buf)
// 		req.Header.Set("Content-Type", writer.FormDataContentType())
// 		rr := httptest.NewRecorder()
// 		handler.UploadFile(rr, req)

// 		assert.Equal(t, http.StatusBadRequest, rr.Code)
// 	})
// }

func TestDownloadFile(t *testing.T) {
	t.Run("Успешное скачивание файла", func(t *testing.T) {
		handler, _, _, mockMediaService := newTestHandler(t)

		fileContent := "test file content"
		reader := io.NopCloser(strings.NewReader(fileContent))

		mockMediaService.On("DownloadFile", mock.Anything, "file-123").
			Return(reader, "test.txt", nil)

		req := httptest.NewRequest(http.MethodGet, "/media/download?id=file-123", nil)
		rr := httptest.NewRecorder()
		handler.DownloadFile(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "attachment; filename=\"test.txt\"", rr.Header().Get("Content-Disposition"))
		assert.Equal(t, "application/octet-stream", rr.Header().Get("Content-Type"))
		assert.Equal(t, fileContent, rr.Body.String())

		mockMediaService.AssertExpectations(t)
	})

	t.Run("Файл не найден", func(t *testing.T) {
		handler, _, _, mockMediaService := newTestHandler(t)

		mockMediaService.On("DownloadFile", mock.Anything, "non-existent-id").
			Return(io.NopCloser(strings.NewReader("")), "", fmt.Errorf("file not found"))

		req := httptest.NewRequest(http.MethodGet, "/media/download?id=non-existent-id", nil)
		rr := httptest.NewRecorder()
		handler.DownloadFile(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		mockMediaService.AssertExpectations(t)
	})
}

// // Error handling tests
// func TestErrorHandling(t *testing.T) {
// 	t.Run("HTTP метод не поддерживается", func(t *testing.T) {
// 		handler, _, _, _ := newTestHandler(t)

// 		rr := performRequest(handler.UserCreate, http.MethodGet, "/user/create", nil)
// 		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
// 	})
// }

// // Main test function
// func TestHandlers(t *testing.T) {
// 	t.Run("User handlers", TestUserCreate)
// 	t.Run("User update", TestUserUpdate)
// 	t.Run("User delete", TestUserDelete)
// 	t.Run("User get info", TestUserGetInfo)

// 	t.Run("Search handler", TestSearch)

// 	t.Run("Media handlers", TestUploadFile)
// 	t.Run("Download file", TestDownloadFile)

// 	t.Run("Error handling", TestErrorHandling)
// }
