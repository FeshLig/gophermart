package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockWithdrawalRepo struct {
	mock.Mock
}

func (m *MockWithdrawalRepo) GetTotalWithdrawn(ctx context.Context, userID int64) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockWithdrawalRepo) CreateWithdrawal(ctx context.Context, w model.Withdrawal) error {
	args := m.Called(ctx, w)
	return args.Error(0)
}

func (m *MockWithdrawalRepo) GetWithdrawalsByUser(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Withdrawal), args.Error(1)
}

type MockBalanceService struct {
	mock.Mock
}

func (m *MockBalanceService) GetBalance(ctx context.Context, userID int64) (dto.Balance, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(dto.Balance), args.Error(1)
}

func TestWithdraw(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockWithdrawalRepo)
	mockBalance := new(MockBalanceService)

	svc := service.NewWithdrawalService(mockRepo, mockBalance)

	userID := int64(1)
	order := "12345"
	sum := 50.0

	t.Run("success", func(t *testing.T) {
		mockBalance.On("GetBalance", ctx, userID).Return(dto.Balance{Current: 100.0}, nil).Once()
		mockRepo.On("CreateWithdrawal", ctx, mock.AnythingOfType("model.Withdrawal")).Return(nil).Once()

		err := svc.Withdraw(ctx, userID, order, sum)
		require.NoError(t, err)

		mockBalance.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		mockBalance.On("GetBalance", ctx, userID).Return(dto.Balance{Current: 10.0}, nil).Once()

		err := svc.Withdraw(ctx, userID, order, sum)
		require.ErrorIs(t, err, service.ErrInsufficientFunds)

		mockBalance.AssertExpectations(t)
	})

	t.Run("invalid order number", func(t *testing.T) {
		mockBalance.On("GetBalance", ctx, userID).Return(dto.Balance{Current: 100.0}, nil).Once()

		err := svc.Withdraw(ctx, userID, "abc123", sum)
		require.ErrorIs(t, err, service.ErrInvalidOrderNumber)

		mockBalance.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		mockBalance.On("GetBalance", ctx, userID).Return(dto.Balance{Current: 100.0}, nil).Once()
		mockRepo.On("CreateWithdrawal", ctx, mock.AnythingOfType("model.Withdrawal")).Return(errors.New("repo error")).Once()

		err := svc.Withdraw(ctx, userID, order, sum)
		require.Error(t, err)
		require.Contains(t, err.Error(), "repo error")

		mockBalance.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetWithdrawals(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockWithdrawalRepo)
	mockBalance := new(MockBalanceService)

	svc := service.NewWithdrawalService(mockRepo, mockBalance)
	userID := int64(1)

	t.Run("success with withdrawals", func(t *testing.T) {
		withdrawals := []model.Withdrawal{
			{UserID: userID, OrderNumber: 123, Sum: 50.0},
			{UserID: userID, OrderNumber: 124, Sum: 30.0},
		}

		mockRepo.On("GetWithdrawalsByUser", ctx, userID).Return(withdrawals, nil).Once()

		result, err := svc.GetWithdrawals(ctx, userID)
		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, withdrawals, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		mockRepo.On("GetWithdrawalsByUser", ctx, userID).Return([]model.Withdrawal{}, nil).Once()

		result, err := svc.GetWithdrawals(ctx, userID)
		require.NoError(t, err)
		require.Empty(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo.On("GetWithdrawalsByUser", ctx, userID).Return(nil, errors.New("repo error")).Once()

		_, err := svc.GetWithdrawals(ctx, userID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "repo error")

		mockRepo.AssertExpectations(t)
	})
}
