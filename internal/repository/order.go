package repository

import (
	"context"
	"errors"

	"github.com/FeshLig/gophermart/internal/model"
)

var ErrOrderNotFound = errors.New("order not found")

func (p *Postgres) CreateOrder(ctx context.Context, order model.Order) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO orders (number, user_id, status)
		VALUES ($1, $2, $3)
	`, order.Number, order.UserID, order.Status)

	return err
}

func (p *Postgres) GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order

	for rows.Next() {
		var o model.Order

		err := rows.Scan(
			&o.Number,
			&o.UserID,
			&o.Status,
			&o.Accrual,
			&o.UploadedAt,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (p *Postgres) GetOrderByNumber(ctx context.Context, number int64) (model.Order, error) {
	var o model.Order

	err := p.pool.QueryRow(ctx, `
		SELECT number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE number = $1
	`, number).Scan(
		&o.Number,
		&o.UserID,
		&o.Status,
		&o.Accrual,
		&o.UploadedAt,
	)

	if err != nil {
		return model.Order{}, err
	}

	return o, nil
}

func (p *Postgres) UpdateOrderStatus(
	ctx context.Context,
	number int64,
	status model.OrderStatus,
	accrual *float64,
) error {

	res, err := p.pool.Exec(ctx, `
		UPDATE orders
		SET status = $1,
				accrual = COALESCE($2, accrual)
		WHERE number = $3
	`, status, accrual, number)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrOrderNotFound
	}

	return nil
}

func (p *Postgres) GetOrdersByStatus(ctx context.Context, statuses ...model.OrderStatus) ([]model.Order, error) {
	query := `
		SELECT number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE status = ANY($1)
	`

	rows, err := p.pool.Query(ctx, query, statuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order

	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.Number, &o.UserID, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}
