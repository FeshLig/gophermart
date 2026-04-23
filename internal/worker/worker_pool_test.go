package worker_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/FeshLig/gophermart/internal/accrual"
	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockRepo struct {
	createOrderFn       func(ctx context.Context, order model.Order) error
	getOrdersByUserFn   func(ctx context.Context, userID int64) ([]model.Order, error)
	getOrderByNumberFn  func(ctx context.Context, number int64) (model.Order, error)
	updateOrderStatusFn func(ctx context.Context, number int64, status model.OrderStatus, accrual *float64) error
	getOrdersByStatusFn func(ctx context.Context, statuses ...model.OrderStatus) ([]model.Order, error)
}

func (m *mockRepo) CreateOrder(ctx context.Context, order model.Order) error {
	if m.createOrderFn != nil {
		return m.createOrderFn(ctx, order)
	}
	return nil
}

func (m *mockRepo) GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	if m.getOrdersByUserFn != nil {
		return m.getOrdersByUserFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockRepo) GetOrderByNumber(ctx context.Context, number int64) (model.Order, error) {
	if m.getOrderByNumberFn != nil {
		return m.getOrderByNumberFn(ctx, number)
	}
	return model.Order{}, nil
}

func (m *mockRepo) UpdateOrderStatus(ctx context.Context, number int64, status model.OrderStatus, accrual *float64) error {
	if m.updateOrderStatusFn != nil {
		return m.updateOrderStatusFn(ctx, number, status, accrual)
	}
	return nil
}

func (m *mockRepo) GetOrdersByStatus(ctx context.Context, statuses ...model.OrderStatus) ([]model.Order, error) {
	if m.getOrdersByStatusFn != nil {
		return m.getOrdersByStatusFn(ctx, statuses...)
	}
	return nil, nil
}

type mockClient struct {
	getOrderFn func(ctx context.Context, number string) (*accrual.Response, int, error)
}

func (m *mockClient) GetOrder(ctx context.Context, number string) (*accrual.Response, int, error) {
	if m.getOrderFn != nil {
		return m.getOrderFn(ctx, number)
	}
	return nil, 0, nil
}

func TestHandle_Success(t *testing.T) {
	ctx := context.Background()

	var called bool

	repo := &mockRepo{
		updateOrderStatusFn: func(ctx context.Context, number int64, status model.OrderStatus, accrual *float64) error {
			called = true
			require.Equal(t, int64(1), number)
			require.Equal(t, model.OrderStatus("PROCESSED"), status)
			require.NotNil(t, accrual)
			require.Equal(t, 100.0, *accrual)
			return nil
		},
	}

	client := &mockClient{
		getOrderFn: func(ctx context.Context, number string) (*accrual.Response, int, error) {
			val := 100.0
			return &accrual.Response{
				Order:   number,
				Status:  "PROCESSED",
				Accrual: &val,
			}, http.StatusOK, nil
		},
	}

	w := worker.NewOrderWorker(repo, client, zap.NewNop())

	w.Handle(ctx, model.Order{Number: 1})

	require.True(t, called)
}

func TestHandle_AccrualError(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{}

	client := &mockClient{
		getOrderFn: func(ctx context.Context, number string) (*accrual.Response, int, error) {
			return nil, 0, assert.AnError
		},
	}

	w := worker.NewOrderWorker(repo, client, zap.NewNop())

	w.Handle(ctx, model.Order{Number: 1})
}

func TestHandle_RateLimit(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{}

	client := &mockClient{
		getOrderFn: func(ctx context.Context, number string) (*accrual.Response, int, error) {
			return nil, http.StatusTooManyRequests, nil
		},
	}

	w := worker.NewOrderWorker(repo, client, zap.NewNop())

	w.Handle(ctx, model.Order{Number: 1})

	select {
	case <-w.PauseCh():
	default:
		t.Fatal("expected pause signal")
	}
}

func TestHandle_NoContent(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{}

	client := &mockClient{
		getOrderFn: func(ctx context.Context, number string) (*accrual.Response, int, error) {
			return nil, http.StatusNoContent, nil
		},
	}

	w := worker.NewOrderWorker(repo, client, zap.NewNop())

	w.Handle(ctx, model.Order{Number: 1})
}

func TestHandle_UpdateError(t *testing.T) {
	ctx := context.Background()

	repo := &mockRepo{
		updateOrderStatusFn: func(ctx context.Context, number int64, status model.OrderStatus, accrual *float64) error {
			return assert.AnError
		},
	}

	client := &mockClient{
		getOrderFn: func(ctx context.Context, number string) (*accrual.Response, int, error) {
			val := 50.0
			return &accrual.Response{
				Status:  "PROCESSED",
				Accrual: &val,
			}, http.StatusOK, nil
		},
	}

	w := worker.NewOrderWorker(repo, client, zap.NewNop())

	w.Handle(ctx, model.Order{Number: 1})
}

func TestProduce(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls int
	var mu sync.Mutex

	done := make(chan struct{})

	repo := &mockRepo{
		getOrdersByStatusFn: func(ctx context.Context, statuses ...model.OrderStatus) ([]model.Order, error) {
			mu.Lock()
			calls++
			mu.Unlock()

			select {
			case done <- struct{}{}:
			default:
			}

			return []model.Order{
				{Number: 1},
			}, nil
		},
	}

	client := &mockClient{
		getOrderFn: func(ctx context.Context, number string) (*accrual.Response, int, error) {
			val := 10.0
			return &accrual.Response{
				Status:  "PROCESSED",
				Accrual: &val,
			}, http.StatusOK, nil
		},
	}

	w := worker.NewOrderWorker(repo, client, zap.NewNop())
	w.SetInterval(5 * time.Millisecond)

	go w.Start(ctx)

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("worker did not produce any jobs")
	}

	cancel()

	mu.Lock()
	defer mu.Unlock()

	require.GreaterOrEqual(t, calls, 1)
}

func TestWorker_ProcessJob(t *testing.T) {
	ctx := context.Background()

	var mu sync.Mutex
	calls := 0

	repo := &mockRepo{
		updateOrderStatusFn: func(ctx context.Context, number int64, status model.OrderStatus, accrual *float64) error {
			mu.Lock()
			defer mu.Unlock()
			calls++
			return nil
		},
	}

	client := &mockClient{
		getOrderFn: func(ctx context.Context, number string) (*accrual.Response, int, error) {
			val := 10.0
			return &accrual.Response{
				Status:  "PROCESSED",
				Accrual: &val,
			}, http.StatusOK, nil
		},
	}

	w := worker.NewOrderWorker(repo, client, zap.NewNop())

	w.Handle(ctx, model.Order{Number: 1})

	require.Equal(t, 1, calls)
}
