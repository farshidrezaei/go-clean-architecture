package user

import (
	"context"
	"errors"
	"strings"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
}

type RegisterUser struct {
	users  repositories.UserRepository
	hasher PasswordHasher
	ids    shared.IDGenerator
	clock  shared.Clock
}

func NewRegisterUser(users repositories.UserRepository, hasher PasswordHasher, ids shared.IDGenerator, clock shared.Clock) *RegisterUser {
	return &RegisterUser{users: users, hasher: hasher, ids: ids, clock: clock}
}

func (uc *RegisterUser) Execute(ctx context.Context, input RegisterUserInput) (*entities.User, error) {
	if len(strings.TrimSpace(input.Password)) < 8 {
		return nil, shared.WrapValidation(entities.ErrWeakPassword, "password does not meet policy")
	}

	existing, err := uc.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil && !errors.Is(err, entities.ErrNotFound) {
		return nil, shared.WrapInternal(err, "failed to check existing user")
	}
	if existing != nil {
		return nil, shared.WrapConflict(entities.ErrConflict, "email is already registered")
	}

	hash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, shared.WrapInternal(err, "failed to protect password")
	}

	userEntity, err := entities.NewUser(uc.ids.NewString(), input.Name, input.Email, hash, entities.RoleUser, uc.clock.Now())
	if err != nil {
		return nil, shared.WrapValidation(err, "user payload is invalid")
	}

	if err := uc.users.Create(ctx, userEntity); err != nil {
		return nil, shared.WrapInternal(err, "failed to persist user")
	}

	return userEntity, nil
}
