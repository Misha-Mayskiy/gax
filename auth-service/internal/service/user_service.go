package service

import (
	"context"
	"fmt"

	"auth-service/pkg/utils"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, uuid, username, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
	UpdatePassword(uuid, newPassword string) error
}

type UserService struct {
	repository UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repository: repo,
	}
}

func (us *UserService) Register(ctx context.Context, username, email, password string) (string, error) {
	userUuid := uuid.New().String()
	passwordHashed, err := utils.GeneratePasswordHash(password)
	if err != nil {
		return "", fmt.Errorf("cant generate password hash")
	}
	err = us.repository.Create(ctx, userUuid, username, email, passwordHashed)
	if err != nil {
		return "", err
	}
	return userUuid, nil
}
func (us *UserService) Login(ctx context.Context, email, password string) (string, error) {
	userUuid, err := us.repository.Login(ctx, email, password)
	if err != nil {
		return "", err
	}
	return userUuid, nil
}
func (us *UserService) ChangePassword(ctx context.Context, uuid, newPassword string) error {
	newPasswordHashed, err := utils.GeneratePasswordHash(newPassword)
	if err != nil {
		return fmt.Errorf("cant generate password hash")
	}
	err = us.repository.UpdatePassword(uuid, newPasswordHashed)
	if err != nil {
		return err
	}
	return nil
}
