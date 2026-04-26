package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/repository"
)

type OrderService interface {
	UploadOrder(ctx context.Context, userID int64, number string) error
	GetOrders(ctx context.Context, userID int64) ([]model.Order, error)
}

type OrderServiceImpl struct {
	repo repository.OrderRepository
}

func NewOrderService(orderRepo repository.OrderRepository) *OrderServiceImpl {
	return &OrderServiceImpl{
		repo: orderRepo,
	}
}

var (
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrOrderConflict      = errors.New("order belongs to another user")
)

func (s *OrderServiceImpl) UploadOrder(ctx context.Context, userID int64, number string) error {

	if !isValidLuhn(number) {
		return ErrInvalidOrderNumber
	}

	num, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		return ErrInvalidOrderNumber
	}

	order := model.Order{
		Number: num,
		UserID: userID,
		Status: model.StatusNew,
	}

	err = s.repo.CreateOrder(ctx, order)
	if err == nil {
		return nil
	}

	if isUniqueViolation(err) {
		existing, err := s.repo.GetOrderByNumber(ctx, num)
		if err != nil {
			return err
		}

		if existing.UserID == userID {
			return ErrOrderAlreadyExists
		}

		return ErrOrderConflict
	}

	return err
}

func isValidLuhn(number string) bool {
	var sum int
	var alternate bool

	for i := len(number) - 1; i >= 0; i-- {

		n := int(number[i] - '0')

		if n < 0 || n > 9 {
			return false
		}

		if alternate {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}

		sum += n
		alternate = !alternate

	}

	return sum%10 == 0
}

func (s *OrderServiceImpl) GetOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	orders, err := s.repo.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if orders == nil {
		return []model.Order{}, nil
	}

	return orders, nil
}
