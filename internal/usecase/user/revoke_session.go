package user

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type RevokeSessionInput struct {
	UserID    string
	SessionID string
}

type RevokeSession struct {
	refreshTokens repositories.RefreshTokenRepository
	clock         shared.Clock
	uow           repositories.UnitOfWork
	events        shared.EventPublisher
}

func NewRevokeSession(refreshTokens repositories.RefreshTokenRepository, clock shared.Clock, uow repositories.UnitOfWork, events shared.EventPublisher) *RevokeSession {
	return &RevokeSession{refreshTokens: refreshTokens, clock: clock, uow: uow, events: events}
}

func (uc *RevokeSession) Execute(ctx context.Context, input RevokeSessionInput) error {
	return uc.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		session, err := uc.refreshTokens.GetByID(txCtx, input.SessionID)
		if err != nil {
			if errors.Is(err, entities.ErrNotFound) {
				return shared.WrapNotFound(err, "session was not found")
			}
			return shared.WrapInternal(err, "failed to load session")
		}
		if session.UserID != input.UserID {
			return shared.WrapForbidden(entities.ErrForbidden, "session does not belong to user")
		}
		if session.IsRevoked() {
			return nil
		}
		if err := uc.refreshTokens.Revoke(txCtx, input.SessionID, uc.clock.Now()); err != nil {
			return shared.WrapInternal(err, "failed to revoke session")
		}
		uc.events.Publish(txCtx, shared.Event{Name: "session.revoked", ActorID: input.UserID, TargetID: input.SessionID})
		return nil
	})
}
