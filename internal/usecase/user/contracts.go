package user

import (
	"context"
	"time"
)

type TokenService interface {
	GenerateAccessToken(ctx context.Context, userID, email string, role string) (string, error)
	NewRefreshToken() (plain string, hashed string, err error)
	HashRefreshToken(plain string) string
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, plain string) error
}

type RefreshTokenLifetime interface {
	RefreshTokenTTL() time.Duration
}
