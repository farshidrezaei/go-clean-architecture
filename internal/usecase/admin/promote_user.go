package admin

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type PromoteUserInput struct {
	ActorRole string
	UserID    string
	Role      string
}

type PromoteUser struct {
	users  repositories.UserRepository
	events shared.EventPublisher
}

func NewPromoteUser(users repositories.UserRepository, events shared.EventPublisher) *PromoteUser {
	return &PromoteUser{users: users, events: events}
}

func (uc *PromoteUser) Execute(ctx context.Context, input PromoteUserInput) (*entities.User, error) {
	if entities.UserRole(input.ActorRole) != entities.RoleAdmin {
		return nil, shared.WrapForbidden(entities.ErrForbidden, "admin role is required")
	}
	userEntity, err := uc.users.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return nil, shared.WrapNotFound(err, "user was not found")
		}
		return nil, shared.WrapInternal(err, "failed to load user")
	}
	userEntity.Role = entities.UserRole(input.Role)
	if err := userEntity.Validate(); err != nil {
		return nil, shared.WrapValidation(err, "role is invalid")
	}
	if err := uc.users.Update(ctx, userEntity); err != nil {
		return nil, shared.WrapInternal(err, "failed to update user role")
	}
	uc.events.Publish(ctx, shared.Event{Name: "admin.user_role_updated", TargetID: userEntity.ID, Metadata: map[string]string{"role": string(userEntity.Role)}})
	return userEntity, nil
}
