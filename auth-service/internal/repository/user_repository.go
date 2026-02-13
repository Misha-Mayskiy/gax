package repository

import (
	"context"
	"fmt"

	"auth-service/pkg/utils"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db      *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db:      pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
func (ur *UserRepository) UserNameIsExist(ctx context.Context, userName string) bool {
	sql, args, err := ur.builder.Select("username").From("users").Where(squirrel.Eq{"username": userName}).ToSql()
	if err != nil {
		return false
	}
	row := ur.db.QueryRow(ctx, sql, args...)
	var dbUserName string
	err = row.Scan(&dbUserName)
	if err != nil {
		return false
	}
	if dbUserName != "" {
		return true
	}
	return false
}
func (ur *UserRepository) EmailIsExist(ctx context.Context, email string) bool {
	sql, args, err := ur.builder.Select("email").From("users").Where(squirrel.Eq{"email": email}).ToSql()
	if err != nil {
		return false
	}
	row := ur.db.QueryRow(ctx, sql, args...)
	var dbEmail string
	err = row.Scan(&dbEmail)
	if err != nil {
		return false
	}
	if dbEmail != "" {
		return true
	}
	return false
}
func (ur *UserRepository) Create(ctx context.Context, uuid, username, email, password string) error {
	userNameIsExist := ur.UserNameIsExist(ctx, username)
	if userNameIsExist {
		return fmt.Errorf("name taken")
	}
	emailIsExist := ur.EmailIsExist(ctx, email)
	if emailIsExist {
		return fmt.Errorf("email taken")
	}
	sql, args, err := ur.builder.Insert("users").
		Columns("uuid", "username", "email", "password").
		Values(uuid, username, email, password).ToSql()
	if err != nil {
		return fmt.Errorf("cant create user")
	}
	_, err = ur.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("cant create user")
	}
	return nil
}

func (ur *UserRepository) Login(ctx context.Context, email, password string) (string, error) {
	if emailIsExist := ur.EmailIsExist(ctx, email); !emailIsExist {
		return "", fmt.Errorf("user with this email not found")
	}
	sql, args, err := ur.builder.Select("uuid", "email", "password").
		From("users").
		Where(squirrel.Eq{"email": email}).ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build query: %w", err)
	}
	row := ur.db.QueryRow(ctx, sql, args...)
	var dbEmail, dbPassword, dbUuid string
	err = row.Scan(&dbUuid, &dbEmail, &dbPassword)
	if err != nil {
		return "", fmt.Errorf("failed to scan row: %w", err)
	}
	if err := utils.ComparePassword(dbPassword, password); err != nil {
		return "", fmt.Errorf("invalid password")
	}
	return dbUuid, nil
}

func (ur *UserRepository) UpdatePassword(uuid, newPassword string) error {
	sql, args, err := ur.builder.Update("users").
		Set("password", newPassword).
		Where(squirrel.Eq{"uuid": uuid}).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	_, err = ur.db.Exec(context.Background(), sql, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	return nil
}
