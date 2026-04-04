package admin

import (
	"context"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type ListUsersInput struct {
	ActorRole string
	Page      int
	Limit     int
}

type ListUsersOutput struct {
	Items []*entities.User
	Page  int
	Limit int
	Total int
}

type ListUsers struct {
	users repositories.UserReader
}

func NewListUsers(users repositories.UserReader) *ListUsers {
	return &ListUsers{users: users}
}

func (uc *ListUsers) Execute(ctx context.Context, input ListUsersInput) (*ListUsersOutput, error) {
	if entities.UserRole(input.ActorRole) != entities.RoleAdmin {
		return nil, shared.WrapForbidden(entities.ErrForbidden, "admin role is required")
	}
	result, err := uc.users.List(ctx, repositories.NewPagination(input.Page, input.Limit))
	if err != nil {
		return nil, shared.WrapInternal(err, "failed to list users")
	}
	return &ListUsersOutput{Items: result.Items, Page: result.Page, Limit: result.Limit, Total: result.Total}, nil
}
