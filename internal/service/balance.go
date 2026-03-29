package service

import (
	"context"

	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/mapper"
	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/repository"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID int64) (dto.Balance, error)
}

type BalanceServiceImpl struct {
	orderRepo      repository.OrderRepository
	withdrawalRepo repository.WithdrawalRepository
}

func NewBalanceService(orderRepo repository.OrderRepository, withdrawalRepo repository.WithdrawalRepository) *BalanceServiceImpl {
	return &BalanceServiceImpl{
		orderRepo:      orderRepo,
		withdrawalRepo: withdrawalRepo,
	}
}

func (s *BalanceServiceImpl) GetBalance(ctx context.Context, userID int64) (dto.Balance, error) {
	orders, err := s.orderRepo.GetOrdersByUser(ctx, userID)
	if err != nil {
		return dto.Balance{}, err
	}

	var current float64
	for _, o := range orders {
		if o.Accrual != nil && o.Status == model.StatusProcessed {
			current += *o.Accrual
		}
	}

	withdrawn, err := s.withdrawalRepo.GetTotalWithdrawn(ctx, userID)
	if err != nil {
		return dto.Balance{}, err
	}

	current -= withdrawn

	return mapper.ToBalanceDTO(current, withdrawn), nil
}
