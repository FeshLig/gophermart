package repository

import (
	"context"
	"errors"

	"github.com/FeshLig/gophermart/internal/model"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type WithdrawalRepository interface {
	CreateWithdrawal(ctx context.Context, withdrawal model.Withdrawal) error
	GetWithdrawalsByUser(ctx context.Context, userID int64) ([]model.Withdrawal, error)
	GetTotalWithdrawn(ctx context.Context, userID int64) (float64, error)
}

func (p *Postgres) CreateWithdrawal(ctx context.Context, withdrawal model.Withdrawal) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var balance float64

	err = tx.QueryRow(ctx, `
	SELECT 
		COALESCE(SUM(accrual), 0) - COALESCE((
			SELECT SUM(sum) FROM withdrawals WHERE user_id = $1
		), 0)
	FROM orders
	WHERE user_id = $1
	  AND status = 'PROCESSED'
	FOR UPDATE
`, withdrawal.UserID).Scan(&balance)

	if err != nil {
		return err
	}

	if balance < withdrawal.Sum {
		return ErrInsufficientFunds
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO withdrawals (user_id, order_number, sum)
		VALUES ($1, $2, $3)
	`, withdrawal.UserID, withdrawal.OrderNumber, withdrawal.Sum)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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
