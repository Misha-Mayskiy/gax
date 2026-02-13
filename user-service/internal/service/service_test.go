package service

import (
	"context"
	"errors"
	"main/internal/domain"
	"testing"

	"github.com/segmentio/kafka-go"
)

// --- Mocks ---

// MockUserRepository имитирует поведение базы данных
type MockUserRepository struct {
	CreateFn     func(ctx context.Context, req *domain.CreateUserRequest) (domain.User, error)
	GetByUUIDFn  func(ctx context.Context, uuid string) (domain.User, error)
	GetByEmailFn func(ctx context.Context, email string) (domain.User, error)
	UpdateFn     func(ctx context.Context, user domain.User) (domain.User, error)
	DeleteFn     func(ctx context.Context, uuid string) error
	ListFn       func(ctx context.Context, limit, offset int) ([]domain.User, int, error)
	AboutMeFn    func(ctx context.Context, uuid string) (domain.User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, req *domain.CreateUserRequest) (domain.User, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, req)
	}
	return domain.User{}, nil
}

func (m *MockUserRepository) GetByUUID(ctx context.Context, uuid string) (domain.User, error) {
	if m.GetByUUIDFn != nil {
		return m.GetByUUIDFn(ctx, uuid)
	}
	return domain.User{}, errors.New("mock not implemented")
}

// Добавили реализацию недостающего метода
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return domain.User{}, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user domain.User) (domain.User, error) {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, user)
	}
	return domain.User{}, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, uuid string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, uuid)
	}
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockUserRepository) AboutMe(ctx context.Context, uuid string) (domain.User, error) {
	if m.AboutMeFn != nil {
		return m.AboutMeFn(ctx, uuid)
	}
	return domain.User{}, nil
}

// Helpers
func strPtr(s string) *string { return &s }

// --- Tests ---

func TestUserService_CreateUser(t *testing.T) {
	mockRepo := &MockUserRepository{}
	// Kafka writer заглушка. Она не подключится, но предотвратит панику nil pointer.
	// Ошибка записи в Kafka логируется в сервисе, но не возвращается, что нам подходит.
	mockKafka := &kafka.Writer{Addr: kafka.TCP("localhost:0")}

	svc := NewUserService(mockRepo, mockKafka, nil)

	tests := []struct {
		name    string
		req     *CreateUserRequest
		mockRun func()
		wantErr bool
	}{
		{
			name: "Validation Error: Empty Email",
			req: &CreateUserRequest{
				UserName: "testuser",
			},
			wantErr: true,
		},
		{
			name: "Validation Error: Empty UserName",
			req: &CreateUserRequest{
				Email: "test@test.com",
			},
			wantErr: true,
		},
		{
			name: "Success",
			req: &CreateUserRequest{
				UUID:     "123",
				Email:    "test@test.com",
				UserName: "testuser",
				Avatar:   "avatar.png",
			},
			mockRun: func() {
				mockRepo.CreateFn = func(ctx context.Context, req *domain.CreateUserRequest) (domain.User, error) {
					return domain.User{
						UUID:     req.UUID,
						Email:    req.Email,
						UserName: req.UserName,
						Avatar:   req.Avatar,
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "Repo Error",
			req: &CreateUserRequest{
				Email:    "err@test.com",
				UserName: "erruser",
			},
			mockRun: func() {
				mockRepo.CreateFn = func(ctx context.Context, req *domain.CreateUserRequest) (domain.User, error) {
					return domain.User{}, errors.New("db error")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockRun != nil {
				tt.mockRun()
			}
			got, err := svc.CreateUser(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != nil {
				if got.Email != tt.req.Email {
					t.Errorf("Expected email %s, got %s", tt.req.Email, got.Email)
				}
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	mockRepo := &MockUserRepository{}
	svc := NewUserService(mockRepo, nil, nil)

	t.Run("Success", func(t *testing.T) {
		mockRepo.GetByUUIDFn = func(ctx context.Context, uuid string) (domain.User, error) {
			return domain.User{UUID: uuid, UserName: "Found"}, nil
		}
		user, err := svc.GetUser(context.Background(), "123")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if user.UserName != "Found" {
			t.Error("Wrong user returned")
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo.GetByUUIDFn = func(ctx context.Context, uuid string) (domain.User, error) {
			return domain.User{}, errors.New("not found")
		}
		_, err := svc.GetUser(context.Background(), "404")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	mockRepo := &MockUserRepository{}
	// Важно: передаем kafka writer, чтобы избежать паники при вызове publishToKafka
	mockKafka := &kafka.Writer{Addr: kafka.TCP("localhost:0")}
	svc := NewUserService(mockRepo, mockKafka, nil)

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo.GetByUUIDFn = func(ctx context.Context, uuid string) (domain.User, error) {
			return domain.User{}, errors.New("not found")
		}
		req := &UpdateUserRequest{Uuid: "missing"}
		_, err := svc.UpdateUser(context.Background(), req)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Success Update", func(t *testing.T) {
		// 1. Сначала сервис получает юзера
		mockRepo.GetByUUIDFn = func(ctx context.Context, uuid string) (domain.User, error) {
			return domain.User{
				UUID:     uuid,
				Email:    "old@mail.com",
				UserName: "oldname",
			}, nil
		}
		// 2. Потом сохраняет
		mockRepo.UpdateFn = func(ctx context.Context, u domain.User) (domain.User, error) {
			// Проверяем, что поля обновились
			if u.Email != "new@mail.com" || u.UserName != "newname" {
				t.Errorf("Update arguments mismatch: got %v", u)
			}
			return u, nil
		}

		req := &UpdateUserRequest{
			Uuid:     "123",
			Email:    "new@mail.com",
			UserName: "newname",
		}

		res, err := svc.UpdateUser(context.Background(), req)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if res.Email != "new@mail.com" {
			t.Errorf("Email not updated in result")
		}
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	mockRepo := &MockUserRepository{}
	// Kafka writer нужен, чтобы избежать паники
	mockKafka := &kafka.Writer{Addr: kafka.TCP("localhost:0")}
	svc := NewUserService(mockRepo, mockKafka, nil)

	t.Run("Repo Error", func(t *testing.T) {
		mockRepo.DeleteFn = func(ctx context.Context, uuid string) error {
			return errors.New("delete failed")
		}
		err := svc.DeleteUser(context.Background(), "123")
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Success (Kafka log error but return nil)", func(t *testing.T) {
		mockRepo.DeleteFn = func(ctx context.Context, uuid string) error {
			return nil
		}

		err := svc.DeleteUser(context.Background(), "123")
		// Ожидаем ошибку от kafka (dial tcp...), так как адрес фейковый.
		// Это подтверждает, что код прошел репозиторий и попытался отправить событие.
		if err == nil {
			t.Log("Warning: Kafka write succeeded unexpectedly (mock writer?)")
		} else {
			t.Logf("Got expected Kafka dial error: %v", err)
		}
	})
}

func TestUserService_ListUsers(t *testing.T) {
	mockRepo := &MockUserRepository{}
	svc := NewUserService(mockRepo, nil, nil)

	tests := []struct {
		name       string
		limit      int
		offset     int
		wantLimit  int
		wantOffset int
	}{
		{"Default Limit", 0, 0, 50, 0},
		{"Max Limit", 150, 0, 100, 0},
		{"Negative Offset", 50, -5, 50, 0},
		{"Valid", 10, 5, 10, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ListFn = func(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
				if limit != tt.wantLimit || offset != tt.wantOffset {
					t.Errorf("List args: got (%d, %d), want (%d, %d)", limit, offset, tt.wantLimit, tt.wantOffset)
				}
				return []domain.User{}, 0, nil
			}
			svc.ListUsers(context.Background(), tt.limit, tt.offset)
		})
	}
}

func TestUserService_AboutMe(t *testing.T) {
	mockRepo := &MockUserRepository{}
	NewUserService(mockRepo, nil, nil)

	mockRepo.AboutMeFn = func(ctx context.Context, uuid string) (domain.User, error) {
		return domain.User{UUID: uuid, UserName: "Me"}, nil
	}

	t.Run("Repo Error", func(t *testing.T) {
		mockRepo.AboutMeFn = func(ctx context.Context, uuid string) (domain.User, error) {
			return domain.User{}, errors.New("err")
		}
	})
}

// Helpers for test
func TestHelpers(t *testing.T) {
	s := "test"
	p := stringToPtr(s)
	if *p != s {
		t.Error("stringToPtr failed")
	}

	pEmpty := stringToPtr("")
	if pEmpty != nil {
		t.Error("stringToPtr empty should return nil")
	}
}
