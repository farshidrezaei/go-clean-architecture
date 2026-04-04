package user

import (
	"context"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type ListSessionsInput struct {
	UserID string
}

type ListSessions struct {
	refreshTokens repositories.RefreshTokenRepository
}

func NewListSessions(refreshTokens repositories.RefreshTokenRepository) *ListSessions {
	return &ListSessions{refreshTokens: refreshTokens}
}

func (uc *ListSessions) Execute(ctx context.Context, input ListSessionsInput) ([]*entities.RefreshToken, error) {
	items, err := uc.refreshTokens.ListByUserID(ctx, input.UserID)
	if err != nil {
		return nil, shared.WrapInternal(err, "failed to list sessions")
	}
	return items, nil
}
