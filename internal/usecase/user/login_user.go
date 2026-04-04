package user

import (
	"context"
	"errors"
	"strings"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type LoginUserInput struct {
	Email      string
	Password   string
	DeviceName string
	UserAgent  string
	IPAddress  string
}

type LoginResult struct {
	User         *entities.User
	AccessToken  string
	RefreshToken string
}

type LoginUser struct {
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

func NewLoginUser(users repositories.UserReader, refreshTokens repositories.RefreshTokenRepository, hasher PasswordHasher, tokens TokenService, ids shared.IDGenerator, clock shared.Clock, ttl RefreshTokenLifetime, uow repositories.UnitOfWork, events shared.EventPublisher) *LoginUser {
	return &LoginUser{
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

func (uc *LoginUser) Execute(ctx context.Context, input LoginUserInput) (*LoginResult, error) {
	userEntity, err := uc.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return nil, shared.WrapUnauthorized(err, "invalid credentials")
		}
		return nil, shared.WrapInternal(err, "failed to load user")
	}

	if err := uc.hasher.Compare(userEntity.PasswordHash, input.Password); err != nil {
		return nil, shared.WrapUnauthorized(entities.ErrUnauthorized, "invalid credentials")
	}

	var result *LoginResult
	err = uc.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		accessToken, err := uc.tokens.GenerateAccessToken(txCtx, userEntity.ID, userEntity.Email, string(userEntity.Role))
		if err != nil {
			return shared.WrapInternal(err, "failed to issue access token")
		}

		refreshPlain, refreshHash, err := uc.tokens.NewRefreshToken()
		if err != nil {
			return shared.WrapInternal(err, "failed to issue refresh token")
		}

		refreshID := uc.ids.NewString()
		refreshEntity, err := entities.NewRefreshToken(
			refreshID,
			refreshID,
			userEntity.ID,
			refreshHash,
			input.DeviceName,
			input.UserAgent,
			input.IPAddress,
			uc.clock.Now().Add(uc.ttl.RefreshTokenTTL()),
			uc.clock.Now(),
		)
		if err != nil {
			return shared.WrapInternal(err, "failed to create refresh session")
		}

		if err := uc.refreshTokens.Create(txCtx, refreshEntity); err != nil {
			return shared.WrapInternal(err, "failed to persist refresh session")
		}

		result = &LoginResult{
			User:         userEntity,
			AccessToken:  accessToken,
			RefreshToken: refreshPlain,
		}
		uc.events.Publish(txCtx, shared.Event{
			Name:     "user.logged_in",
			ActorID:  userEntity.ID,
			TargetID: refreshID,
			Metadata: map[string]string{
				"device_name": input.DeviceName,
				"ip_address":  input.IPAddress,
			},
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
