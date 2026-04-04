package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey struct{}

type UnitOfWork struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{pool: pool}
}

func (u *UnitOfWork) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, contextKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func querierFromContext(ctx context.Context, pool *pgxpool.Pool) interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
} {
	if tx, ok := ctx.Value(contextKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}
