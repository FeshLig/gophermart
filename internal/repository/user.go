package repository

import (
	"context"

	"github.com/FeshLig/gophermart/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user model.User) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
}

func (p *Postgres) CreateUser(ctx context.Context, user model.User) (int64, error) {
	var id int64

	err := p.pool.QueryRow(ctx, `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`, user.Login, user.PasswordHash).Scan(&id)

	return id, err
}

func (p *Postgres) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	var user model.User

	err := p.pool.QueryRow(ctx, `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1
	`, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
