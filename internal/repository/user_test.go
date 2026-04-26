package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/repository"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	user := model.User{
		Login:        "testuser",
		PasswordHash: "123456",
	}

	mock.ExpectQuery("INSERT INTO users.*RETURNING id").
		WithArgs(user.Login, user.PasswordHash).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	id, err := repo.CreateUser(ctx, user)
	require.NoError(t, err)
	require.Equal(t, int64(1), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_QueryError(t *testing.T) {
	ctx := context.Background()

	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	user := model.User{
		Login:        "failuser",
		PasswordHash: "654321",
	}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(user.Login, user.PasswordHash).
		WillReturnError(errors.New("query error"))

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	_, err := repo.CreateUser(ctx, user)
	require.Error(t, err)
	require.Contains(t, err.Error(), "query error")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByLogin(t *testing.T) {
	ctx := context.Background()

	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	now := time.Now()
	mockUser := model.User{
		ID:           1,
		Login:        "testuser",
		PasswordHash: "123456",
		CreatedAt:    now,
	}

	rows := pgxmock.NewRows([]string{"id", "login", "password_hash", "created_at"}).
		AddRow(mockUser.ID, mockUser.Login, mockUser.PasswordHash, mockUser.CreatedAt)

	mock.ExpectQuery("SELECT .* FROM users WHERE login").
		WithArgs(mockUser.Login).
		WillReturnRows(rows)

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	user, err := repo.GetUserByLogin(ctx, mockUser.Login)
	require.NoError(t, err)
	require.Equal(t, mockUser.ID, user.ID)
	require.Equal(t, mockUser.Login, user.Login)
	require.Equal(t, mockUser.PasswordHash, user.PasswordHash)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByLogin_NotFound(t *testing.T) {
	ctx := context.Background()

	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	mock.ExpectQuery("SELECT .* FROM users WHERE login").
		WithArgs("unknown").
		WillReturnError(errors.New("no rows in result set"))

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	_, err := repo.GetUserByLogin(ctx, "unknown")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no rows")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByLogin_ScanError(t *testing.T) {
	ctx := context.Background()

	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	mock.ExpectQuery("SELECT .* FROM users WHERE login").
		WithArgs("user").
		WillReturnError(errors.New("scan error"))

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	_, err := repo.GetUserByLogin(ctx, "user")
	require.Error(t, err)
	require.Contains(t, err.Error(), "scan error")
	require.NoError(t, mock.ExpectationsWereMet())
}
