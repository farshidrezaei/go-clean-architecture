package repositories

import (
	"context"

	"clean_architecture/internal/domain/entities"
)

type UserReader interface {
	GetByID(ctx context.Context, id string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	List(ctx context.Context, pagination Pagination) (PageResult[*entities.User], error)
}

type UserWriter interface {
	Create(ctx context.Context, user *entities.User) error
	Update(ctx context.Context, user *entities.User) error
}

type UserRepository interface {
	UserReader
	UserWriter
}
