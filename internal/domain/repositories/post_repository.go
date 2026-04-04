package repositories

import (
	"context"

	"clean_architecture/internal/domain/entities"
)

type PostReader interface {
	GetByID(ctx context.Context, id string) (*entities.Post, error)
	List(ctx context.Context, pagination Pagination) (PageResult[*entities.Post], error)
}

type PostWriter interface {
	Create(ctx context.Context, post *entities.Post) error
	Update(ctx context.Context, post *entities.Post) error
	Delete(ctx context.Context, id string) error
	AddLike(ctx context.Context, postID, userID string) error
}

type PostRepository interface {
	PostReader
	PostWriter
}
