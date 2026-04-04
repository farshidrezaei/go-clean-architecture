package user

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type LogoutSessionInput struct {
	RefreshToken string
}

type LogoutSession struct {
	refreshTokens repositories.RefreshTokenRepository
	tokens        TokenService
	clock         shared.Clock
	uow           repositories.UnitOfWork
	events        shared.EventPublisher
}

func NewLogoutSession(refreshTokens repositories.RefreshTokenRepository, tokens TokenService, clock shared.Clock, uow repositories.UnitOfWork, events shared.EventPublisher) *LogoutSession {
	return &LogoutSession{refreshTokens: refreshTokens, tokens: tokens, clock: clock, uow: uow, events: events}
}

func (uc *LogoutSession) Execute(ctx context.Context, input LogoutSessionInput) error {
	return uc.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		hash := uc.tokens.HashRefreshToken(input.RefreshToken)
		stored, err := uc.refreshTokens.GetByTokenHash(txCtx, hash)
		if err != nil {
			if errors.Is(err, entities.ErrNotFound) {
				return shared.WrapUnauthorized(entities.ErrInvalidToken, "refresh token is invalid")
			}
			return shared.WrapInternal(err, "failed to load refresh session")
		}
		if stored.IsRevoked() {
			return nil
		}
		if err := uc.refreshTokens.Revoke(txCtx, stored.ID, uc.clock.Now()); err != nil {
			return shared.WrapInternal(err, "failed to revoke refresh session")
		}
		uc.events.Publish(txCtx, shared.Event{Name: "session.logged_out", ActorID: stored.UserID, TargetID: stored.ID})
		return nil
	})
}
