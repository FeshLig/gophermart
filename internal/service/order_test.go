package service_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) CreateOrder(ctx context.Context, order model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *MockOrderRepo) GetOrderByNumber(ctx context.Context, number int64) (model.Order, error) {
	args := m.Called(ctx, number)
	return args.Get(0).(model.Order), args.Error(1)
}

func (m *MockOrderRepo) UpdateOrderStatus(ctx context.Context, number int64, status model.OrderStatus, accrual *float64) error {
	args := m.Called(ctx, number, status, accrual)
	return args.Error(0)
}

func (m *MockOrderRepo) GetOrdersByStatus(ctx context.Context, statuses ...model.OrderStatus) ([]model.Order, error) {
	args := m.Called(ctx, statuses)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Order), args.Error(1)
}

func TestUploadOrder(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockOrderRepo)
	svc := service.NewOrderService(mockRepo)
	userID := int64(1)
	orderNumber := "79927398713"

	t.Run("success upload", func(t *testing.T) {
		mockRepo.On("CreateOrder", ctx, mock.AnythingOfType("model.Order")).Return(nil).Once()

		err := svc.UploadOrder(ctx, userID, orderNumber)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid order number (Luhn)", func(t *testing.T) {
		err := svc.UploadOrder(ctx, userID, "1234567890")
		require.ErrorIs(t, err, service.ErrInvalidOrderNumber)
	})

	t.Run("unique violation, same user", func(t *testing.T) {
		orderNum, _ := strconv.ParseInt(orderNumber, 10, 64)
		mockRepo.On("CreateOrder", ctx, mock.AnythingOfType("model.Order")).
			Return(service.ErrUniqueViolation).Once()
		mockRepo.On("GetOrderByNumber", ctx, orderNum).
			Return(model.Order{UserID: userID, Number: orderNum}, nil).Once()

		err := svc.UploadOrder(ctx, userID, orderNumber)
		require.ErrorIs(t, err, service.ErrOrderAlreadyExists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unique violation, another user", func(t *testing.T) {
		orderNum, _ := strconv.ParseInt(orderNumber, 10, 64)
		mockRepo.On("CreateOrder", ctx, mock.AnythingOfType("model.Order")).
			Return(service.ErrUniqueViolation).Once()
		mockRepo.On("GetOrderByNumber", ctx, orderNum).
			Return(model.Order{UserID: 2, Number: orderNum}, nil).Once()

		err := svc.UploadOrder(ctx, userID, orderNumber)
		require.ErrorIs(t, err, service.ErrOrderConflict)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create order error", func(t *testing.T) {
		mockRepo.On("CreateOrder", ctx, mock.AnythingOfType("model.Order")).Return(errors.New("db error")).Once()

		err := svc.UploadOrder(ctx, userID, orderNumber)
		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")
		mockRepo.AssertExpectations(t)
	})
}

func TestGetOrders(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockOrderRepo)
	svc := service.NewOrderService(mockRepo)
	userID := int64(1)

	t.Run("success with orders", func(t *testing.T) {
		orders := []model.Order{
			{Number: 123, UserID: userID, Status: model.StatusNew},
			{Number: 124, UserID: userID, Status: model.StatusProcessing},
		}
		mockRepo.On("GetOrdersByUser", ctx, userID).Return(orders, nil).Once()

		result, err := svc.GetOrders(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, orders, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		mockRepo.On("GetOrdersByUser", ctx, userID).Return([]model.Order{}, nil).Once()

		result, err := svc.GetOrders(ctx, userID)
		require.NoError(t, err)
		require.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo.On("GetOrdersByUser", ctx, userID).Return(nil, errors.New("repo error")).Once()

		_, err := svc.GetOrders(ctx, userID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "repo error")
		mockRepo.AssertExpectations(t)
	})
}
