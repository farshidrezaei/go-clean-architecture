package user

import (
	"context"

	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type LogoutAllSessionsInput struct {
	UserID string
}

type LogoutAllSessions struct {
	refreshTokens repositories.RefreshTokenRepository
	clock         shared.Clock
	events        shared.EventPublisher
}

func NewLogoutAllSessions(refreshTokens repositories.RefreshTokenRepository, clock shared.Clock, events shared.EventPublisher) *LogoutAllSessions {
	return &LogoutAllSessions{refreshTokens: refreshTokens, clock: clock, events: events}
}

func (uc *LogoutAllSessions) Execute(ctx context.Context, input LogoutAllSessionsInput) error {
	if err := uc.refreshTokens.RevokeAllByUserID(ctx, input.UserID, uc.clock.Now()); err != nil {
		return shared.WrapInternal(err, "failed to revoke user sessions")
	}
	uc.events.Publish(ctx, shared.Event{Name: "session.logged_out_all", ActorID: input.UserID, TargetID: input.UserID})
	return nil
}
