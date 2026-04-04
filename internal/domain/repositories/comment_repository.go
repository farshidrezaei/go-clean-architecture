package repositories

import (
	"context"

	"clean_architecture/internal/domain/entities"
)

type CommentReader interface {
	GetByID(ctx context.Context, id string) (*entities.Comment, error)
	ListByPostID(ctx context.Context, postID string, pagination Pagination) (PageResult[*entities.Comment], error)
}

type CommentWriter interface {
	Create(ctx context.Context, comment *entities.Comment) error
	Update(ctx context.Context, comment *entities.Comment) error
	Delete(ctx context.Context, id string) error
}

type CommentRepository interface {
	CommentReader
	CommentWriter
}
