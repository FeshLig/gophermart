package handler

import "github.com/FeshLig/gophermart/internal/service"

type Handlers struct {
	Auth       *AuthHandler
	Order      *OrderHandler
	Balance    *BalanceHandler
	Withdrawal *WithdrawalHandler
}

func NewHandlers(
	authService service.AuthService,
	orderService service.OrderService,
	balanceService service.BalanceService,
	withdrawalService service.WithdrawalService,
) *Handlers {
	return &Handlers{
		Auth:       NewAuthHandler(authService),
		Order:      NewOrderHandler(orderService),
		Balance:    NewBalanceHandler(balanceService, withdrawalService),
		Withdrawal: NewWithdrawalHandler(withdrawalService),
	}
}
