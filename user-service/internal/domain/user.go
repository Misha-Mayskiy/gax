// domain/user.go
package domain

import (
	"time"

	// Импортируйте правильный пакет для вашего proto
	// Это может быть другой путь
	userpb "main/pkg/api"
)

type User struct {
	UUID      string    `json:"uuid" db:"uuid"`
	Email     string    `json:"email" db:"email"`
	UserName  string    `json:"user_name" db:"user_name"`
	Avatar    *string   `json:"avatar,omitempty" db:"avatar"`
	AboutMe   *string   `json:"about_me,omitempty" db:"about_me"`
	Status    string    `json:"status" db:"status"`
	Friends   []string  `json:"friends,omitempty" db:"friends"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
	UUID     string   `json:"uuid"`
	Email    string   `json:"email"`
	UserName string   `json:"user_name"`
	Avatar   *string  `json:"avatar,omitempty"`
	AboutMe  *string  `json:"about_me,omitempty"`
	Status   string   `json:"status"`
	Friends  []string `json:"friends,omitempty" db:"friends"`
}

func (u *User) ToProto() *userpb.UserResponse {
	resp := &userpb.UserResponse{
		Uuid:      u.UUID,
		Email:     u.Email,
		UserName:  u.UserName,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}

	// Обрабатываем optional поля
	if u.Avatar != nil {
		resp.Avatar = *u.Avatar
	}
	if u.AboutMe != nil {
		resp.AboutMe = *u.AboutMe
	}

	// Если в proto есть поле Friends
	// resp.Friends = u.Friends

	return resp
}
