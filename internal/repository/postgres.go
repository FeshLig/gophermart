package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Close()
}

type Postgres struct {
	pool DB
}

func NewPostgresWithDB(db DB) *Postgres {
	return &Postgres{pool: db}
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("db init failed: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *Postgres) Close() {
	p.pool.Close()
}
