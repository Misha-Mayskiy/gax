package domain

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrUserNameRequired = errors.New("user name is required")
)