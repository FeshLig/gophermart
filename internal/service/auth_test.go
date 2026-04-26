package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) CreateUser(ctx context.Context, user model.User) (int64, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepo) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return model.User{}, args.Error(1)
	}
	return args.Get(0).(model.User), args.Error(1)
}

func TestRegister(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepo)
	svc := service.NewAuthService(mockRepo)

	login := "testuser"
	password := "password123"

	t.Run("success", func(t *testing.T) {
		mockRepo.
			On("CreateUser", ctx, mock.AnythingOfType("model.User")).
			Return(int64(1), nil).
			Once()

		token, err := svc.Register(ctx, login, password)

		require.NoError(t, err)
		require.NotEmpty(t, token)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user already exists (unique violation)", func(t *testing.T) {
		mockRepo.
			On("CreateUser", ctx, mock.AnythingOfType("model.User")).
			Return(int64(0), service.ErrUniqueViolation).
			Once()

		_, err := svc.Register(ctx, login, password)

		require.ErrorIs(t, err, service.ErrUserAlreadyExists)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo.
			On("CreateUser", ctx, mock.AnythingOfType("model.User")).
			Return(int64(0), errors.New("db error")).
			Once()

		_, err := svc.Register(ctx, login, password)

		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")

		mockRepo.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepo)
	svc := service.NewAuthService(mockRepo)

	login := "testuser"
	password := "password123"

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	t.Run("success", func(t *testing.T) {
		mockRepo.
			On("GetUserByLogin", ctx, login).
			Return(model.User{
				ID:           1,
				Login:        login,
				PasswordHash: string(hash),
			}, nil).
			Once()

		token, err := svc.Login(ctx, login, password)

		require.NoError(t, err)
		require.NotEmpty(t, token)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.
			On("GetUserByLogin", ctx, login).
			Return(nil, pgx.ErrNoRows).
			Once()

		_, err := svc.Login(ctx, login, password)

		require.ErrorIs(t, err, service.ErrInvalidCredentials)

		mockRepo.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockRepo.
			On("GetUserByLogin", ctx, login).
			Return(model.User{
				ID:           1,
				Login:        login,
				PasswordHash: string(hash),
			}, nil).
			Once()

		_, err := svc.Login(ctx, login, "wrongpassword")

		require.ErrorIs(t, err, service.ErrInvalidCredentials)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo.
			On("GetUserByLogin", ctx, login).
			Return(nil, errors.New("db error")).
			Once()

		_, err := svc.Login(ctx, login, password)

		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")

		mockRepo.AssertExpectations(t)
	})
}
