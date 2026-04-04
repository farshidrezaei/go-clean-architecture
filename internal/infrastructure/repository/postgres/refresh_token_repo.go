package postgres

import (
	"context"
	"errors"
	"time"

	"clean_architecture/internal/domain/entities"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	q := querierFromContext(ctx, r.pool)
	_, err := q.Exec(ctx, `
		INSERT INTO refresh_tokens (id, family_id, user_id, token_hash, device_name, user_agent, ip_address, expires_at, revoked_at, replaced_by_id, compromised_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, token.ID, token.FamilyID, token.UserID, token.TokenHash, token.DeviceName, token.UserAgent, token.IPAddress, token.ExpiresAt, token.RevokedAt, token.ReplacedByID, token.CompromisedAt, token.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entities.ErrConflict
		}
		return err
	}
	return nil
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	q := querierFromContext(ctx, r.pool)
	var token entities.RefreshToken
	err := q.QueryRow(ctx, `
		SELECT id, family_id, user_id, token_hash, device_name, user_agent, ip_address, expires_at, revoked_at, replaced_by_id, compromised_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(&token.ID, &token.FamilyID, &token.UserID, &token.TokenHash, &token.DeviceName, &token.UserAgent, &token.IPAddress, &token.ExpiresAt, &token.RevokedAt, &token.ReplacedByID, &token.CompromisedAt, &token.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenRepository) GetByID(ctx context.Context, id string) (*entities.RefreshToken, error) {
	q := querierFromContext(ctx, r.pool)
	var token entities.RefreshToken
	err := q.QueryRow(ctx, `
		SELECT id, family_id, user_id, token_hash, device_name, user_agent, ip_address, expires_at, revoked_at, replaced_by_id, compromised_at, created_at
		FROM refresh_tokens
		WHERE id = $1
	`, id).Scan(&token.ID, &token.FamilyID, &token.UserID, &token.TokenHash, &token.DeviceName, &token.UserAgent, &token.IPAddress, &token.ExpiresAt, &token.RevokedAt, &token.ReplacedByID, &token.CompromisedAt, &token.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenRepository) ListByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error) {
	q := querierFromContext(ctx, r.pool)
	rows, err := q.Query(ctx, `
		SELECT id, family_id, user_id, token_hash, device_name, user_agent, ip_address, expires_at, revoked_at, replaced_by_id, compromised_at, created_at
		FROM refresh_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]*entities.RefreshToken, 0)
	for rows.Next() {
		var token entities.RefreshToken
		if err := rows.Scan(&token.ID, &token.FamilyID, &token.UserID, &token.TokenHash, &token.DeviceName, &token.UserAgent, &token.IPAddress, &token.ExpiresAt, &token.RevokedAt, &token.ReplacedByID, &token.CompromisedAt, &token.CreatedAt); err != nil {
			return nil, err
		}
		copyValue := token
		items = append(items, &copyValue)
	}
	return items, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	q := querierFromContext(ctx, r.pool)
	tag, err := q.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1`, id, revokedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string, revokedAt time.Time) error {
	q := querierFromContext(ctx, r.pool)
	_, err := q.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL`, userID, revokedAt)
	return err
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, id, replacedByID string, revokedAt time.Time) error {
	q := querierFromContext(ctx, r.pool)
	tag, err := q.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = $2, replaced_by_id = $3
		WHERE id = $1
	`, id, revokedAt, replacedByID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeFamily(ctx context.Context, familyID string, revokedAt time.Time) error {
	q := querierFromContext(ctx, r.pool)
	_, err := q.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = COALESCE(revoked_at, $2), compromised_at = $2
		WHERE family_id = $1
	`, familyID, revokedAt)
	return err
}
