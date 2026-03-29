package service

import "github.com/FeshLig/gophermart/internal/repository"

type Services struct {
	Auth       AuthService
	Order      OrderService
	Balance    BalanceService
	Withdrawal WithdrawalService
}

func NewServices(repo repository.Repository) *Services {

	authService := NewAuthService(repo)
	orderService := NewOrderService(repo)
	balanceService := NewBalanceService(repo, repo)

	withdrawalService := NewWithdrawalService(repo, balanceService)

	return &Services{
		Auth:       authService,
		Order:      orderService,
		Balance:    balanceService,
		Withdrawal: withdrawalService,
	}
}
