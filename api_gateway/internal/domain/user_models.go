package domain

import "time"

type User struct {
	UUID      string    `json:"uuid"`
	Email     string    `json:"email"`
	UserName  string    `json:"user_name"`
	Avatar    *string   `json:"avatar,omitempty"`
	AboutMe   *string   `json:"about_me,omitempty"`
	Status    string    `json:"status"`
	Friends   []string  `json:"friends,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserCreateRequest struct {
	UUID     string   `json:"uuid,omitempty"`
	Email    string   `json:"email"`
	UserName string   `json:"user_name"`
	Avatar   string   `json:"avatar,omitempty"`   // Изменил *string на string
	AboutMe  string   `json:"about_me,omitempty"` // Изменил *string на string
	Friends  []string `json:"friends,omitempty"`
}

type UserUpdateRequest struct {
	Uuid     string   `json:"uuid"`
	Email    string   `json:"email,omitempty"`
	UserName string   `json:"user_name,omitempty"`
	Avatar   string   `json:"avatar,omitempty"`
	AboutMe  string   `json:"about_me,omitempty"`
	Friends  []string `json:"friends,omitempty"`
}

type UserDeleteRequest struct {
	Uuid string `json:"uuid"`
}

type UserGetRequest struct {
	Uuid string `json:"uuid"`
}

type UserSetOnlineRequest struct {
	Uuid       string `json:"uuid"`
	TtlSeconds int32  `json:"ttl_seconds"`
}

type UserIsOnlineRequest struct {
	Uuid string `json:"uuid"`
}
