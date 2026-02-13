package repository

import (
	"context"
	"database/sql"
	"fmt"
	"main/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRepository interface {
	Create(ctx context.Context, req *domain.CreateUserRequest) (domain.User, error)
	GetByUUID(ctx context.Context, uuid string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Update(ctx context.Context, user domain.User) (domain.User, error)
	Delete(ctx context.Context, uuid string) error
	List(ctx context.Context, limit, offset int) ([]domain.User, int, error)
	AboutMe(ctx context.Context, uuid string) (domain.User, error)
}

type pgUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &pgUserRepository{db: db}
}

func (r *pgUserRepository) Create(ctx context.Context, req *domain.CreateUserRequest) (domain.User, error) {
	// Генерируем UUID если не передан
	userUUID := req.UUID
	if userUUID == "" {
		userUUID = uuid.New().String()
	}

	query := `
		INSERT INTO users (
			uuid, email, user_name, avatar, about_me, status, friends
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING uuid, email, user_name, avatar, about_me, status, 
				  friends, created_at, updated_at
	`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query,
		userUUID,
		req.Email,
		req.UserName,
		req.Avatar,
		req.AboutMe,
		req.Status,
		pq.Array([]string{}),
	).Scan(
		&user.UUID,
		&user.Email,
		&user.UserName,
		&user.Avatar,
		&user.AboutMe,
		&user.Status,
		pq.Array(&user.Friends),
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return domain.User{}, fmt.Errorf("failed to create user: %w", err)
	}
	fmt.Println(req)
	return user, nil
}

func (r *pgUserRepository) GetByUUID(ctx context.Context, uuid string) (domain.User, error) {
	query := `
		SELECT uuid, email, user_name, avatar, about_me, status, 
			   friends, created_at, updated_at
		FROM users
		WHERE uuid = $1
	`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, uuid).Scan(
		&user.UUID,
		&user.Email,
		&user.UserName,
		&user.Avatar,
		&user.AboutMe,
		&user.Status,
		pq.Array(&user.Friends),
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, fmt.Errorf("user not found")
		}
		return domain.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *pgUserRepository) Update(ctx context.Context, user domain.User) (domain.User, error) {
	query := `
		UPDATE users 
		SET email = $1, user_name = $2, avatar = $3, 
			about_me = $4, status = $5, friends = $6,
			updated_at = $7
		WHERE uuid = $8
		RETURNING uuid, email, user_name, avatar, about_me, status, 
				  friends, created_at, updated_at
	`

	var updatedUser domain.User
	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.UserName,
		user.Avatar,
		user.AboutMe,
		user.Status,
		pq.Array(user.Friends),
		time.Now(),
		user.UUID,
	).Scan(
		&updatedUser.UUID,
		&updatedUser.Email,
		&updatedUser.UserName,
		&updatedUser.Avatar,
		&updatedUser.AboutMe,
		&updatedUser.Status,
		pq.Array(&updatedUser.Friends),
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, fmt.Errorf("user not found")
		}
		return domain.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

func (r *pgUserRepository) Delete(ctx context.Context, uuid string) error {
	query := `DELETE FROM users WHERE uuid = $1`
	result, err := r.db.ExecContext(ctx, query, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *pgUserRepository) AboutMe(ctx context.Context, uuid string) (domain.User, error) {
	// То же самое что GetByUUID
	return r.GetByUUID(ctx, uuid)
}

func (r *pgUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `
		SELECT uuid, email, user_name, avatar, about_me, status, 
			   friends, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.UUID,
		&user.Email,
		&user.UserName,
		&user.Avatar,
		&user.AboutMe,
		&user.Status,
		pq.Array(&user.Friends),
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, fmt.Errorf("user not found")
		}
		return domain.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *pgUserRepository) List(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	// TODO: реализовать если нужно
	return nil, 0, nil
}
