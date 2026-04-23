package repository

import (
	"context"

	"github.com/FeshLig/gophermart/internal/model"
)

type Repository interface {
	UserRepository
	OrderRepository
	WithdrawalRepository
}

type UserRepository interface {
	CreateUser(ctx context.Context, user model.User) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order model.Order) error
	GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error)
	GetOrderByNumber(ctx context.Context, number int64) (model.Order, error)
	UpdateOrderStatus(ctx context.Context, number int64, status model.OrderStatus, accrual *float64) error
	GetOrdersByStatus(ctx context.Context, statuses ...model.OrderStatus) ([]model.Order, error)
}

type WithdrawalRepository interface {
	CreateWithdrawal(ctx context.Context, withdrawal model.Withdrawal) error
	GetWithdrawalsByUser(ctx context.Context, userID int64) ([]model.Withdrawal, error)
	GetTotalWithdrawn(ctx context.Context, userID int64) (float64, error)
}
