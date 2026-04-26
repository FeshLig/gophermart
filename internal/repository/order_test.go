package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/repository"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateOrder(t *testing.T) {
	ctx := context.Background()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	order := model.Order{
		Number:  123,
		UserID:  1,
		Status:  model.StatusNew,
		Accrual: nil,
	}

	mock.ExpectExec("INSERT INTO orders").
		WithArgs(order.Number, order.UserID, order.Status).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	err = repo.CreateOrder(ctx, order)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrdersByUser(t *testing.T) {
	ctx := context.Background()

	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	now := time.Now()
	rows := pgxmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at"}).
		AddRow(int64(123), int64(1), model.StatusNew, nil, now).
		AddRow(int64(456), int64(1), model.StatusProcessing, nil, now)

	mock.ExpectQuery("SELECT .* FROM orders WHERE user_id").
		WithArgs(int64(1)).
		WillReturnRows(rows)

	result, err := repo.GetOrdersByUser(ctx, 1)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, int64(123), result[0].Number)
	require.Equal(t, int64(456), result[1].Number)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderByNumber(t *testing.T) {
	ctx := context.Background()

	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	now := time.Now()
	row := pgxmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at"}).
		AddRow(int64(123), int64(1), model.StatusNew, nil, now)

	mock.ExpectQuery("SELECT .* FROM orders WHERE number").
		WithArgs(int64(123)).
		WillReturnRows(row)

	result, err := repo.GetOrderByNumber(ctx, 123)
	require.NoError(t, err)
	require.Equal(t, int64(123), result.Number)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrderStatus(t *testing.T) {
	ctx := context.Background()

	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	status := model.StatusProcessed
	accrual := 50.0

	mock.ExpectExec("UPDATE orders").
		WithArgs(status, &accrual, int64(123)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.UpdateOrderStatus(ctx, 123, status, &accrual)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrderStatus_NotFound(t *testing.T) {
	ctx := context.Background()

	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	status := model.StatusProcessed
	accrual := 50.0

	mock.ExpectExec("UPDATE orders").
		WithArgs(status, &accrual, int64(999)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := repo.UpdateOrderStatus(ctx, 999, status, &accrual)
	require.ErrorIs(t, err, repository.ErrOrderNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrdersByStatus(t *testing.T) {
	ctx := context.Background()

	mock, _ := pgxmock.NewPool()
	defer mock.Close()

	repo := repository.NewPostgresWithDB(mock, zap.NewNop())

	now := time.Now()
	rows := pgxmock.NewRows([]string{"number", "user_id", "status", "accrual", "uploaded_at"}).
		AddRow(int64(123), int64(1), model.StatusNew, nil, now).
		AddRow(int64(456), int64(2), model.StatusNew, nil, now)

	statuses := []model.OrderStatus{model.StatusNew}

	mock.ExpectQuery("SELECT .* FROM orders WHERE status").
		WithArgs(statuses).
		WillReturnRows(rows)

	result, err := repo.GetOrdersByStatus(ctx, statuses...)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, int64(123), result[0].Number)
	require.NoError(t, mock.ExpectationsWereMet())
}
