package user

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type RefreshSessionInput struct {
	RefreshToken string
}

type RefreshSessionResult struct {
	User         *entities.User
	AccessToken  string
	RefreshToken string
}

type RefreshSession struct {
	users         repositories.UserReader
	refreshTokens repositories.RefreshTokenRepository
	hasher        PasswordHasher
	tokens        TokenService
	ids           shared.IDGenerator
	clock         shared.Clock
	ttl           RefreshTokenLifetime
	uow           repositories.UnitOfWork
	events        shared.EventPublisher
}

func NewRefreshSession(users repositories.UserReader, refreshTokens repositories.RefreshTokenRepository, hasher PasswordHasher, tokens TokenService, ids shared.IDGenerator, clock shared.Clock, ttl RefreshTokenLifetime, uow repositories.UnitOfWork, events shared.EventPublisher) *RefreshSession {
	return &RefreshSession{
		users:         users,
		refreshTokens: refreshTokens,
		hasher:        hasher,
		tokens:        tokens,
		ids:           ids,
		clock:         clock,
		ttl:           ttl,
		uow:           uow,
		events:        events,
	}
}

func (uc *RefreshSession) Execute(ctx context.Context, input RefreshSessionInput) (*RefreshSessionResult, error) {
	var result *RefreshSessionResult

	err := uc.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		refreshHash := uc.tokens.HashRefreshToken(input.RefreshToken)

		stored, err := uc.refreshTokens.GetByTokenHash(txCtx, refreshHash)
		if err != nil {
			if errors.Is(err, entities.ErrNotFound) {
				return shared.WrapUnauthorized(entities.ErrInvalidToken, "refresh token is invalid")
			}
			return shared.WrapInternal(err, "failed to load refresh session")
		}

		now := uc.clock.Now()
		if stored.IsRevoked() {
			if stored.WasRotated() {
				_ = uc.refreshTokens.RevokeFamily(txCtx, stored.FamilyID, now)
				uc.events.Publish(txCtx, shared.Event{Name: "session.reuse_detected", ActorID: stored.UserID, TargetID: stored.FamilyID})
				return shared.WrapUnauthorized(entities.ErrInvalidToken, "refresh token reuse detected")
			}
			return shared.WrapUnauthorized(entities.ErrInvalidToken, "refresh token is revoked")
		}
		if stored.IsExpired(now) {
			_ = uc.refreshTokens.Revoke(txCtx, stored.ID, now)
			return shared.WrapUnauthorized(entities.ErrTokenExpired, "refresh token has expired")
		}

		userEntity, err := uc.users.GetByID(txCtx, stored.UserID)
		if err != nil {
			if errors.Is(err, entities.ErrNotFound) {
				return shared.WrapUnauthorized(entities.ErrUnauthorized, "user no longer exists")
			}
			return shared.WrapInternal(err, "failed to load user")
		}

		nextPlain, nextHash, err := uc.tokens.NewRefreshToken()
		if err != nil {
			return shared.WrapInternal(err, "failed to issue refresh token")
		}

		nextID := uc.ids.NewString()
		nextEntity, err := entities.NewRefreshToken(
			nextID,
			stored.FamilyID,
			userEntity.ID,
			nextHash,
			stored.DeviceName,
			stored.UserAgent,
			stored.IPAddress,
			now.Add(uc.ttl.RefreshTokenTTL()),
			now,
		)
		if err != nil {
			return shared.WrapInternal(err, "failed to create refresh session")
		}
		if err := uc.refreshTokens.Create(txCtx, nextEntity); err != nil {
			return shared.WrapInternal(err, "failed to persist refresh session")
		}
		if err := uc.refreshTokens.Rotate(txCtx, stored.ID, nextID, now); err != nil {
			return shared.WrapInternal(err, "failed to rotate refresh session")
		}

		accessToken, err := uc.tokens.GenerateAccessToken(txCtx, userEntity.ID, userEntity.Email, string(userEntity.Role))
		if err != nil {
			return shared.WrapInternal(err, "failed to issue access token")
		}

		result = &RefreshSessionResult{
			User:         userEntity,
			AccessToken:  accessToken,
			RefreshToken: nextPlain,
		}
		uc.events.Publish(txCtx, shared.Event{Name: "session.refreshed", ActorID: userEntity.ID, TargetID: nextID})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
