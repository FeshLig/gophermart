package service_test

import (
	"context"
	"testing"

	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGetBalance(t *testing.T) {
	ctx := context.Background()

	mockOrderRepo := new(MockOrderRepo)
	mockWithdrawalRepo := new(MockWithdrawalRepo)

	svc := service.NewBalanceService(mockOrderRepo, mockWithdrawalRepo)

	userID := int64(1)

	t.Run("success", func(t *testing.T) {
		accrual1 := 100.0
		accrual2 := 50.0

		orders := []model.Order{
			{UserID: userID, Status: model.StatusProcessed, Accrual: &accrual1},
			{UserID: userID, Status: model.StatusProcessed, Accrual: &accrual2},
			{UserID: userID, Status: model.StatusNew, Accrual: &accrual1},
			{UserID: userID, Status: model.StatusProcessed, Accrual: nil},
		}

		mockOrderRepo.
			On("GetOrdersByUser", ctx, userID).
			Return(orders, nil).
			Once()

		mockWithdrawalRepo.
			On("GetTotalWithdrawn", ctx, userID).
			Return(80.0, nil).
			Once()

		balance, err := svc.GetBalance(ctx, userID)

		require.NoError(t, err)
		require.Equal(t, dto.Balance{
			Current:   70.0,
			Withdrawn: 80.0,
		}, balance)

		mockOrderRepo.AssertExpectations(t)
		mockWithdrawalRepo.AssertExpectations(t)
	})

}
