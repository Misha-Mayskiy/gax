package repository

import (
	"context"
	"database/sql"
	"media-service/internal/domain"
)

type PgRepo struct {
	db *sql.DB
}

func NewPgRepo(db *sql.DB) *PgRepo {
	return &PgRepo{db: db}
}

// Save сохраняет метаданные файла в БД
func (r *PgRepo) Save(ctx context.Context, f *domain.FileMeta) error {
	query := `INSERT INTO files (id, filename, bucket, object_name, content_type, size, created_at, user_id, description, chat_id) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.ExecContext(ctx, query,
		f.ID, f.Filename, f.Bucket, f.ObjectName, f.ContentType, f.Size, f.CreatedAt, f.UserID, f.Description, f.ChatID,
	)
	return err
}

// GetByID получает метаданные по UUID
func (r *PgRepo) GetByID(ctx context.Context, id string) (*domain.FileMeta, error) {
	f := &domain.FileMeta{}
	query := `SELECT id, filename, bucket, object_name, content_type, size, created_at, user_id, description, chat_id 
	          FROM files WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&f.ID, &f.Filename, &f.Bucket, &f.ObjectName, &f.ContentType, &f.Size, &f.CreatedAt, &f.UserID, &f.Description, &f.ChatID,
	)
	return f, err
}

// GetFilesByUser получает список файлов пользователя
func (r *PgRepo) GetFilesByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.FileMeta, error) {
	query := `SELECT id, filename, bucket, object_name, content_type, size, created_at, user_id, description, chat_id 
	          FROM files WHERE user_id = $1 
	          ORDER BY created_at DESC 
	          LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*domain.FileMeta
	for rows.Next() {
		f := &domain.FileMeta{}
		err := rows.Scan(
			&f.ID, &f.Filename, &f.Bucket, &f.ObjectName, &f.ContentType, &f.Size, &f.CreatedAt, &f.UserID, &f.Description, &f.ChatID,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

// GetTotalFilesByUser получает общее количество файлов пользователя
func (r *PgRepo) GetTotalFilesByUser(ctx context.Context, userID string) (int, error) {
	var total int
	query := `SELECT COUNT(*) FROM files WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	return total, err
}

// Delete удаляет запись о файле по id
func (r *PgRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM files WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
