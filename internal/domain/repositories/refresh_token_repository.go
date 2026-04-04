package repositories

import (
	"context"
	"time"

	"clean_architecture/internal/domain/entities"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entities.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)
	GetByID(ctx context.Context, id string) (*entities.RefreshToken, error)
	ListByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error)
	Revoke(ctx context.Context, id string, revokedAt time.Time) error
	Rotate(ctx context.Context, id, replacedByID string, revokedAt time.Time) error
	RevokeAllByUserID(ctx context.Context, userID string, revokedAt time.Time) error
	RevokeFamily(ctx context.Context, familyID string, revokedAt time.Time) error
}
