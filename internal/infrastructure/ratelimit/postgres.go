package ratelimit

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresService struct {
	pool *pgxpool.Pool
}

func NewPostgresService(pool *pgxpool.Pool) *PostgresService {
	return &PostgresService{pool: pool}
}

func (s *PostgresService) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(window)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM rate_limit_windows WHERE expires_at <= $1`, now); err != nil {
		return false, err
	}

	var count int
	var currentExpiry time.Time
	err = tx.QueryRow(ctx, `
		SELECT count, expires_at
		FROM rate_limit_windows
		WHERE key = $1
		FOR UPDATE
	`, key).Scan(&count, &currentExpiry)
	if err != nil {
		if err != pgx.ErrNoRows {
			return false, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO rate_limit_windows(key, count, expires_at)
			VALUES ($1, 1, $2)
			ON CONFLICT (key) DO NOTHING
		`, key, expiresAt); err != nil {
			return false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return true, nil
	}

	if now.After(currentExpiry) {
		if _, err := tx.Exec(ctx, `UPDATE rate_limit_windows SET count = 1, expires_at = $2 WHERE key = $1`, key, expiresAt); err != nil {
			return false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return true, nil
	}

	if count >= limit {
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return false, nil
	}

	if _, err := tx.Exec(ctx, `UPDATE rate_limit_windows SET count = count + 1 WHERE key = $1`, key); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}
