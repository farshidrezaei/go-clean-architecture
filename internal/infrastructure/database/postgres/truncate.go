package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Truncate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			post_likes,
			comments,
			posts,
			refresh_tokens,
			audit_logs,
			rate_limit_windows,
			users
		RESTART IDENTITY CASCADE
	`)
	return err
}
