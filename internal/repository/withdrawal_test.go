package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/repository"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func TestCreateWithdrawal(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			name: "success",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()

				rows := pgxmock.NewRows([]string{"balance"}).
					AddRow(100.0)

				mock.ExpectQuery("SELECT").
					WithArgs(int64(1)).
					WillReturnRows(rows)

				mock.ExpectExec("INSERT INTO withdrawals").
					WithArgs(int64(1), int64(123), 50.0).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectCommit()
			},
		},
		{
			name: "begin error",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			expectedError: errors.New("begin error"),
		},
		{
			name: "select error",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()

				mock.ExpectQuery("SELECT").
					WithArgs(int64(1)).
					WillReturnError(errors.New("select error"))

				mock.ExpectRollback()
			},
			expectedError: errors.New("select error"),
		},
		{
			name: "insufficient funds",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()

				rows := pgxmock.NewRows([]string{"balance"}).
					AddRow(10.0)

				mock.ExpectQuery("SELECT").
					WithArgs(int64(1)).
					WillReturnRows(rows)

				mock.ExpectRollback()
			},
			expectedError: repository.ErrInsufficientFunds,
		},
		{
			name: "insert error",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()

				rows := pgxmock.NewRows([]string{"balance"}).
					AddRow(100.0)

				mock.ExpectQuery("SELECT").
					WithArgs(int64(1)).
					WillReturnRows(rows)

				mock.ExpectExec("INSERT INTO withdrawals").
					WithArgs(int64(1), int64(123), 50.0).
					WillReturnError(errors.New("insert error"))

				mock.ExpectRollback()
			},
			expectedError: errors.New("insert error"),
		},
		{
			name: "commit error",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()

				rows := pgxmock.NewRows([]string{"balance"}).
					AddRow(100.0)

				mock.ExpectQuery("SELECT").
					WithArgs(int64(1)).
					WillReturnRows(rows)

				mock.ExpectExec("INSERT INTO withdrawals").
					WithArgs(int64(1), int64(123), 50.0).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectedError: errors.New("commit error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockSetup(mock)

			repo := repository.NewPostgresWithDB(mock)

			err = repo.CreateWithdrawal(ctx, model.Withdrawal{
				UserID:      1,
				OrderNumber: 123,
				Sum:         50,
			})

			if tt.expectedError != nil {
				require.Error(t, err)

				if errors.Is(tt.expectedError, repository.ErrInsufficientFunds) {
					require.ErrorIs(t, err, repository.ErrInsufficientFunds)
				}
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetWithdrawalsByUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()

		repo := repository.NewPostgresWithDB(mock)

		now := time.Now()

		rows := pgxmock.NewRows([]string{
			"id", "user_id", "order_number", "sum", "processed_at",
		}).
			AddRow(int64(1), int64(1), int64(123), 50.0, now).
			AddRow(int64(2), int64(1), int64(456), 30.0, now)

		mock.ExpectQuery("SELECT").
			WithArgs(int64(1)).
			WillReturnRows(rows)

		result, err := repo.GetWithdrawalsByUser(ctx, 1)

		require.NoError(t, err)
		require.Len(t, result, 2)
	})

	t.Run("empty result", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()

		repo := repository.NewPostgresWithDB(mock)

		rows := pgxmock.NewRows([]string{
			"id", "user_id", "order_number", "sum", "processed_at",
		})

		mock.ExpectQuery("SELECT").
			WithArgs(int64(1)).
			WillReturnRows(rows)

		result, err := repo.GetWithdrawalsByUser(ctx, 1)

		require.NoError(t, err)
		require.Empty(t, result)
	})

	t.Run("query error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()

		repo := repository.NewPostgresWithDB(mock)

		mock.ExpectQuery("SELECT").
			WithArgs(int64(1)).
			WillReturnError(errors.New("query error"))

		_, err := repo.GetWithdrawalsByUser(ctx, 1)

		require.Error(t, err)
	})

	t.Run("scan error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()

		repo := repository.NewPostgresWithDB(mock)

		rows := pgxmock.NewRows([]string{
			"id", "user_id", "order_number", "sum", "processed_at",
		}).
			AddRow("wrong", int64(1), "123", 50.0, time.Now())

		mock.ExpectQuery("SELECT").
			WithArgs(int64(1)).
			WillReturnRows(rows)

		_, err := repo.GetWithdrawalsByUser(ctx, 1)

		require.Error(t, err)
	})
}

func TestGetTotalWithdrawn(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()

		repo := repository.NewPostgresWithDB(mock)

		rows := pgxmock.NewRows([]string{"sum"}).AddRow(150.0)

		mock.ExpectQuery("SELECT").
			WithArgs(int64(1)).
			WillReturnRows(rows)

		total, err := repo.GetTotalWithdrawn(ctx, 1)

		require.NoError(t, err)
		require.Equal(t, 150.0, total)
	})

	t.Run("zero result", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()

		repo := repository.NewPostgresWithDB(mock)

		rows := pgxmock.NewRows([]string{"sum"}).AddRow(0.0)

		mock.ExpectQuery("SELECT").
			WithArgs(int64(1)).
			WillReturnRows(rows)

		total, err := repo.GetTotalWithdrawn(ctx, 1)

		require.NoError(t, err)
		require.Equal(t, 0.0, total)
	})

	t.Run("query error", func(t *testing.T) {
		mock, _ := pgxmock.NewPool()
		defer mock.Close()

		repo := repository.NewPostgresWithDB(mock)

		mock.ExpectQuery("SELECT").
			WithArgs(int64(1)).
			WillReturnError(errors.New("query error"))

		_, err := repo.GetTotalWithdrawn(ctx, 1)

		require.Error(t, err)
	})
}
