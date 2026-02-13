package service

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/domain"
	"main/internal/logger"
	"main/internal/repository"
	"main/internal/repository/redis"
	user "main/pkg/api"
	"time"

	"github.com/segmentio/kafka-go"
)

// Интерфейс сервиса
type UserService interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*domain.User, error)
	GetUser(ctx context.Context, uuid string) (*domain.User, error)
	UpdateUser(ctx context.Context, req *UpdateUserRequest) (*domain.User, error)
	DeleteUser(ctx context.Context, uuid string) error
	ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error)
	AboutMeUser(ctx context.Context, req *user.AboutMeRequest) (*user.UserResponse, error)
	SetOnlineUser(ctx context.Context, uuid string, ttl time.Duration) error
	SetOffline(ctx context.Context, uuid string) error
	IsOnline(ctx context.Context, uuid string) (bool, error)
	GetOnlineUsers(ctx context.Context) ([]string, error)
}

type userService struct {
	repo      repository.UserRepository
	kafkaProd *kafka.Writer
	redis     *redis.Client // Если используете redis
}

func NewUserService(repo repository.UserRepository, kafkaProd *kafka.Writer, redisClient *redis.Client) UserService {
	return &userService{
		repo:      repo,
		kafkaProd: kafkaProd,
		redis:     redisClient,
	}
}

var log = logger.GetLogger()

func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*domain.User, error) {
	// Валидация
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.UserName == "" {
		return nil, fmt.Errorf("user name is required")
	}

	// Преобразуем CreateUserRequest в domain.CreateUserRequest
	domainReq := domain.CreateUserRequest{
		UUID:     req.UUID, // Передаем UUID из запроса
		Email:    req.Email,
		UserName: req.UserName,
		Avatar:   stringToPtr(req.Avatar),
		AboutMe:  stringToPtr(req.AboutMe),
		Status:   "offline",
		Friends:  req.Friends, // Передаем friends
	}

	// Вызываем репозиторий
	user, err := s.repo.Create(ctx, &domainReq)
	if err != nil {
		return nil, err
	}

	// Отправляем событие в Kafka
	if s.kafkaProd != nil {
		if err := s.publishToKafka(&user); err != nil {
			log.Printf("failed to publish user creation event: %v", err)
		}
	}

	return &user, nil
}

func (s *userService) GetUser(ctx context.Context, uuid string) (*domain.User, error) {
	user, err := s.repo.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userService) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*domain.User, error) {
	// Получаем существующего пользователя
	user, err := s.repo.GetByUUID(ctx, req.Uuid)
	if err != nil {
		return nil, err
	}

	// Обновляем только переданные поля
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.UserName != "" {
		user.UserName = req.UserName
	}
	if req.Avatar != "" {
		avatar := req.Avatar
		user.Avatar = &avatar
	}
	if req.AboutMe != "" {
		aboutMe := req.AboutMe
		user.AboutMe = &aboutMe
	}
	user.UpdatedAt = time.Now()

	// Сохраняем в БД
	updatedUser, err := s.repo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	// Отправляем событие в Kafka
	if err := s.publishToKafka(&updatedUser); err != nil {
		log.Printf("failed to publish user update event: %v", err)
	}
	log.Debug().Msg("service update user")

	return &updatedUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, uuid string) error {
	if err := s.repo.Delete(ctx, uuid); err != nil {
		return err
	}
	// Отправляем событие о том, что пользователь удалён
	evt := SearchEvent{Type: "user", Data: map[string]string{"uuid": uuid, "deleted": "true"}}
	payload, _ := json.Marshal(evt)
	log.Debug().Msg("service delete user")

	return s.kafkaProd.WriteMessages(ctx, kafka.Message{Topic: "search-events", Value: payload})
}

func (s *userService) AboutMeUser(ctx context.Context, req *user.AboutMeRequest) (*user.UserResponse, error) {
	log.Debug().Msg("service about user")

	// Получаем пользователя по UUID из запроса
	user, err := s.repo.AboutMe(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}

	// Преобразуем в proto ответ
	return user.ToProto(), nil
}
func (s *userService) GetUserProto(ctx context.Context, uuid string) (*user.UserResponse, error) {
	user, err := s.repo.AboutMe(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return user.ToProto(), nil
}
func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *userService) SetOnlineUser(ctx context.Context, uuid string, ttl time.Duration) error {
	return s.redis.SetOnline(ctx, uuid, ttl)
}

func (s *userService) SetOffline(ctx context.Context, uuid string) error {
	return s.redis.SetOffline(ctx, uuid)
}

func (s *userService) IsOnline(ctx context.Context, uuid string) (bool, error) {
	return s.redis.IsOnline(ctx, uuid)
}

func (s *userService) GetOnlineUsers(ctx context.Context) ([]string, error) {
	return s.redis.GetOnlineUsers(ctx)
}

// Вспомогательные структуры для сервиса
type CreateUserRequest struct {
	UUID     string   `json:"uuid"`
	Email    string   `json:"email"`
	UserName string   `json:"user_name"`
	Avatar   string   `json:"avatar"`
	AboutMe  string   `json:"about_me"`
	Friends  []string `json:"friends"`
}

func (s *userService) publishToKafka(u *domain.User) error {
	evt := SearchEvent{Type: "user", Data: u}
	payload, _ := json.Marshal(evt)

	return s.kafkaProd.WriteMessages(context.Background(),
		kafka.Message{Value: payload})
}

type SearchEvent struct {
	Type string      `json:"type"` // "user", "chat", "message"
	Data interface{} `json:"data"` // сам объект
}

type UpdateUserRequest struct {
	Uuid     string   `json:"uuid"`
	Email    string   `json:"email"`
	UserName string   `json:"user_name"`
	Avatar   string   `json:"avatar"`
	AboutMe  string   `json:"about_me"`
	Friends  []string `json:"friends"`
}

// Вспомогательная функция
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
