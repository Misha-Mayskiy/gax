package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"main/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB, log zerolog.Logger) *Repo { return &Repo{db: db} }

func (r *Repo) SaveMessage(ctx context.Context, m domain.Message, log zerolog.Logger) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO messages (id, chat_id, author_id, text, created_at)
         VALUES ($1,$2,$3,$4,$5)
         ON CONFLICT (id) DO NOTHING`,
		m.ID, m.ChatID, m.AuthorID, m.Text, m.CreatedAt)
	if err != nil {
		log.Err(err).Msg("error saveMessage postrger")
	}
	return err
}

func (r *Repo) SaveChat(ctx context.Context, c domain.Chat, log zerolog.Logger) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO chats (id, kind, title, created_by, created_at)
         VALUES ($1,$2,$3,$4,$5)
         ON CONFLICT (id) DO NOTHING`,
		c.ID, c.Kind, c.Title, c.CreatedBy, c.CreatedAt)
	if err != nil {
		log.Err(err).Msg("error SaveChat postrger")
	}
	return err
}
func (r *Repo) SaveUser(ctx context.Context, u domain.UserIndex, log zerolog.Logger) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, username, email, about_me, updated_at)
         VALUES ($1,$2,$3,$4,$5)
         ON CONFLICT (id) DO NOTHING`,
		u.UUID, u.UserName, u.Email, u.AboutMe, u.UpdatedAt)
	if err != nil {
		log.Err(err).Msg("error SaveChat postrger")
	}
	return err
}
func (r *Repo) SaveMedia(ctx context.Context, m domain.Media, log zerolog.Logger) error {
	// Генерируем ID если не указан
	if m.ID == "" {
		m.ID = uuid.New().String()
	}

	// Устанавливаем время создания если не указано
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO files (id, filename, bucket, object_name, content_type, size, created_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7)
         ON CONFLICT (id) DO NOTHING`,
		m.ID, m.Filename, m.Bucket, m.ObjectName, m.ContentType, m.Size, m.CreatedAt)

	if err != nil {
		log.Err(err).Msg("error SaveMedia postgres")
		return fmt.Errorf("failed to save media: %w", err)
	}

	return nil
}

// GetMedia получает информацию о медиафайле по ID
func (r *Repo) GetMedia(ctx context.Context, id string, log zerolog.Logger) (domain.Media, error) {
	var media domain.Media

	query := `SELECT id, filename, bucket, object_name, content_type, size, created_at 
			  FROM files WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&media.ID,
		&media.Filename,
		&media.Bucket,
		&media.ObjectName,
		&media.ContentType,
		&media.Size,
		&media.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Media{}, fmt.Errorf("media not found")
		}
		log.Err(err).Msg("error GetMedia postgres")
		return domain.Media{}, fmt.Errorf("failed to get media: %w", err)
	}

	return media, nil
}

// DeleteMedia удаляет информацию о медиафайле
func (r *Repo) DeleteMedia(ctx context.Context, id string, log zerolog.Logger) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM files WHERE id = $1", id)
	if err != nil {
		log.Err(err).Msg("error DeleteMedia postgres")
		return fmt.Errorf("failed to delete media: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("media not found")
	}

	return nil
}

// ListMedia получает список медиафайлов с пагинацией
func (r *Repo) ListMedia(ctx context.Context, limit, offset int, log zerolog.Logger) ([]domain.Media, error) {
	query := `SELECT id, filename, bucket, object_name, content_type, size, created_at 
			  FROM files ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.Err(err).Msg("error ListMedia postgres")
		return nil, fmt.Errorf("failed to list media: %w", err)
	}
	defer rows.Close()

	var mediaList []domain.Media
	for rows.Next() {
		var media domain.Media
		if err := rows.Scan(
			&media.ID,
			&media.Filename,
			&media.Bucket,
			&media.ObjectName,
			&media.ContentType,
			&media.Size,
			&media.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan media row: %w", err)
		}
		mediaList = append(mediaList, media)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating media rows: %w", err)
	}

	return mediaList, nil
}

// CountMedia возвращает общее количество медиафайлов
func (r *Repo) CountMedia(ctx context.Context, log zerolog.Logger) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM files").Scan(&count)
	if err != nil {
		log.Err(err).Msg("error CountMedia postgres")
		return 0, fmt.Errorf("failed to count media: %w", err)
	}
	return count, nil
}
