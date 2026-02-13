package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"main/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// Helper для создания мока
func newMockRepo(t *testing.T) (*sql.DB, sqlmock.Sqlmock, UserRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	repo := NewUserRepository(db)
	return db, mock, repo
}

func TestNewUserRepository(t *testing.T) {
	db, _, repo := newMockRepo(t)
	defer db.Close()
	assert.NotNil(t, repo)
}

func TestPgUserRepository_Create(t *testing.T) {
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	ctx := context.Background()
	req := &domain.CreateUserRequest{
		UUID:     "test-uuid",
		Email:    "test@example.com",
		UserName: "tester",
		Status:   "active",
	}

	// Ожидаемый результат
	now := time.Now()
	expectedUser := domain.User{
		UUID:      req.UUID,
		Email:     req.Email,
		UserName:  req.UserName,
		Avatar:    req.Avatar,
		AboutMe:   req.AboutMe,
		Status:    req.Status,
		Friends:   []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Настраиваем mock
	// Мы используем regexp.QuoteMeta, чтобы экранировать спецсимволы,
	// но так как запрос многострочный, проще использовать частичное совпадение
	query := "INSERT INTO users"

	rows := sqlmock.NewRows([]string{"uuid", "email", "user_name", "avatar", "about_me", "status", "friends", "created_at", "updated_at"}).
		AddRow(expectedUser.UUID, expectedUser.Email, expectedUser.UserName, expectedUser.Avatar, expectedUser.AboutMe, expectedUser.Status, "{}" /* pg array format */, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery(query).
		WithArgs(req.UUID, req.Email, req.UserName, req.Avatar, req.AboutMe, req.Status, sqlmock.AnyArg()). // AnyArg для pq.Array, т.к. сложно матчить драйверный тип
		WillReturnRows(rows)

	// Вызов
	user, err := repo.Create(ctx, req)

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.UUID, user.UUID)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.Empty(t, user.Friends)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPgUserRepository_Create_Error(t *testing.T) {
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	req := &domain.CreateUserRequest{Email: "fail@test.com"}

	mock.ExpectQuery("INSERT INTO users").
		WillReturnError(errors.New("db error"))

	_, err := repo.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestPgUserRepository_GetByUUID(t *testing.T) {
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	ctx := context.Background()
	uuid := "some-uuid"
	now := time.Now()

	rows := sqlmock.NewRows([]string{"uuid", "email", "user_name", "avatar", "about_me", "status", "friends", "created_at", "updated_at"}).
		AddRow(uuid, "e@mail.ru", "user", "a", "b", "s", "{f1,f2}", now, now)

	// Используем regexp для гибкости (пробелы, переносы строк)
	query := regexp.QuoteMeta("SELECT uuid, email, user_name, avatar, about_me, status, \n\t\t\t   friends, created_at, updated_at\n\t\tFROM users\n\t\tWHERE uuid = $1")

	mock.ExpectQuery(query).WithArgs(uuid).WillReturnRows(rows)

	user, err := repo.GetByUUID(ctx, uuid)
	assert.NoError(t, err)
	assert.Equal(t, uuid, user.UUID)
	assert.Equal(t, []string{"f1", "f2"}, user.Friends)
}

func TestPgUserRepository_GetByUUID_NotFound(t *testing.T) {
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	mock.ExpectQuery("SELECT .* FROM users").
		WithArgs("missing-uuid").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByUUID(context.Background(), "missing-uuid")
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestPgUserRepository_Update(t *testing.T) {
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	ctx := context.Background()
	updateUser := domain.User{
		UUID:     "u1",
		Email:    "new@mail",
		UserName: "newname",
		Status:   "offline",
		Friends:  []string{"f1"},
	}

	rows := sqlmock.NewRows([]string{"uuid", "email", "user_name", "avatar", "about_me", "status", "friends", "created_at", "updated_at"}).
		AddRow(updateUser.UUID, updateUser.Email, updateUser.UserName, updateUser.Avatar, updateUser.AboutMe, updateUser.Status, "{f1}", time.Now(), time.Now())

	mock.ExpectQuery("UPDATE users").
		WithArgs(
			updateUser.Email,
			updateUser.UserName,
			updateUser.Avatar,
			updateUser.AboutMe,
			updateUser.Status,
			sqlmock.AnyArg(), // pq.Array
			sqlmock.AnyArg(), // time.Now()
			updateUser.UUID,
		).
		WillReturnRows(rows)

	res, err := repo.Update(ctx, updateUser)
	assert.NoError(t, err)
	assert.Equal(t, "newname", res.UserName)
}

func TestPgUserRepository_Update_NotFound(t *testing.T) {
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	mock.ExpectQuery("UPDATE users").WillReturnError(sql.ErrNoRows)

	_, err := repo.Update(context.Background(), domain.User{UUID: "missing"})
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestPgUserRepository_Delete(t *testing.T) {
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	uuid := "del-uuid"

	// Успешное удаление (affected 1 row)
	mock.ExpectExec("DELETE FROM users").
		WithArgs(uuid).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), uuid)
	assert.NoError(t, err)
}

func TestPgUserRepository_Delete_NotFound(t *testing.T) {
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	// Удаление несуществующего (affected 0 rows)
	mock.ExpectExec("DELETE FROM users").
		WithArgs("missing").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), "missing")
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestPgUserRepository_Delete_Error(t *testing.T) {
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	mock.ExpectExec("DELETE FROM users").
		WillReturnError(errors.New("db down"))

	err := repo.Delete(context.Background(), "u")
	assert.Error(t, err)
}

func TestPgUserRepository_GetByEmail(t *testing.T) {
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	email := "find@me.com"
	rows := sqlmock.NewRows([]string{"uuid", "email", "user_name", "avatar", "about_me", "status", "friends", "created_at", "updated_at"}).
		AddRow("u1", email, "n", "a", "b", "s", "{}", time.Now(), time.Now())

	mock.ExpectQuery("SELECT .* FROM users WHERE email =").
		WithArgs(email).
		WillReturnRows(rows)

	user, err := repo.GetByEmail(context.Background(), email)
	assert.NoError(t, err)
	assert.Equal(t, email, user.Email)
}

func TestPgUserRepository_AboutMe(t *testing.T) {
	// AboutMe просто вызывает GetByUUID, но для покрытия нужно вызвать его
	db, mock, repo := newMockRepo(t)
	defer db.Close()

	uuid := "me"
	rows := sqlmock.NewRows([]string{"uuid", "email", "user_name", "avatar", "about_me", "status", "friends", "created_at", "updated_at"}).
		AddRow(uuid, "e", "n", "a", "b", "s", "{}", time.Now(), time.Now())

	mock.ExpectQuery("SELECT .* FROM users WHERE uuid =").
		WithArgs(uuid).
		WillReturnRows(rows)

	user, err := repo.AboutMe(context.Background(), uuid)
	assert.NoError(t, err)
	assert.Equal(t, uuid, user.UUID)
}

func TestPgUserRepository_List(t *testing.T) {
	// Метод заглушка, просто проверяем возврат
	db, _, repo := newMockRepo(t)
	defer db.Close()

	list, count, err := repo.List(context.Background(), 10, 0)
	assert.Nil(t, list)
	assert.Equal(t, 0, count)
	assert.Nil(t, err)
}
