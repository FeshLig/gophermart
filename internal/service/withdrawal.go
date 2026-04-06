package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/repository"
)

type WithdrawalService interface {
	Withdraw(ctx context.Context, userID int64, order string, sum float64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

type WithdrawalServiceImpl struct {
	withdrawalRepo repository.WithdrawalRepository
	balanceService BalanceService
}

var ErrInsufficientFunds = errors.New("insufficient funds")

func NewWithdrawalService(repo repository.WithdrawalRepository, balance BalanceService) WithdrawalService {
	return &WithdrawalServiceImpl{
		withdrawalRepo: repo,
		balanceService: balance,
	}
}

func (s *WithdrawalServiceImpl) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {

	balanceDTO, err := s.balanceService.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	if balanceDTO.Current < sum {
		return ErrInsufficientFunds
	}

	orderNumber, err := strconv.ParseInt(order, 10, 64)
	if err != nil {
		return ErrInvalidOrderNumber
	}

	withdrawal := model.Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Sum:         sum,
	}

	return s.withdrawalRepo.CreateWithdrawal(ctx, withdrawal)
}

func (s *WithdrawalServiceImpl) GetWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	withdrawals, err := s.withdrawalRepo.GetWithdrawalsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if withdrawals == nil {
		return []model.Withdrawal{}, nil
	}

	return withdrawals, nil
}
