package repositories

import "context"

type UnitOfWork interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
