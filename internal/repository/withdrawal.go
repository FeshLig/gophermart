package repository

import (
	"context"
	"errors"

	"github.com/FeshLig/gophermart/internal/model"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

func (p *Postgres) CreateWithdrawal(ctx context.Context, withdrawal model.Withdrawal) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			p.logger.Warn("failed to rollback transaction",
				zap.Int64("user_id", withdrawal.UserID),
				zap.Int64("order_number", withdrawal.OrderNumber),
				zap.Error(err),
			)
		}
	}()

	var ok bool

	err = tx.QueryRow(ctx, `
		WITH balance AS (
			SELECT 
				COALESCE(SUM(o.accrual), 0)
				- COALESCE(
					(SELECT SUM(w.sum) FROM withdrawals w WHERE w.user_id = $1),
					0
				) AS available
			FROM orders o
			WHERE o.user_id = $1
			  AND o.status = 'PROCESSED'
		)
		SELECT available >= $2
		FROM balance
	`, withdrawal.UserID, withdrawal.Sum).Scan(&ok)

	if err != nil {
		return err
	}

	if !ok {
		return ErrInsufficientFunds
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO withdrawals (user_id, order_number, sum)
		VALUES ($1, $2, $3)
	`, withdrawal.UserID, withdrawal.OrderNumber, withdrawal.Sum)

	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetWithdrawalsByUser(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, user_id, order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Withdrawal

	for rows.Next() {
		var w model.Withdrawal

		err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.OrderNumber,
			&w.Sum,
			&w.ProcessedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (p *Postgres) GetTotalWithdrawn(ctx context.Context, userID int64) (float64, error) {
	var total float64

	err := p.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(sum), 0)
		FROM withdrawals
		WHERE user_id = $1
	`, userID).Scan(&total)

	if err != nil {
		return 0, err
	}

	return total, nil
}
