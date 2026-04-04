package post

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type DeletePostInput struct {
	PostID    string
	ActorID   string
	ActorRole string
}

type DeletePost struct {
	posts repositories.PostRepository
}

func NewDeletePost(posts repositories.PostRepository) *DeletePost {
	return &DeletePost{posts: posts}
}

func (uc *DeletePost) Execute(ctx context.Context, input DeletePostInput) error {
	postEntity, err := uc.posts.GetByID(ctx, input.PostID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return shared.WrapNotFound(err, "post was not found")
		}
		return shared.WrapInternal(err, "failed to load post")
	}

	if !postEntity.CanBeEditedBy(input.ActorID, entities.UserRole(input.ActorRole)) {
		return shared.WrapForbidden(entities.ErrForbidden, "only the author can delete the post")
	}

	if err := uc.posts.Delete(ctx, input.PostID); err != nil {
		return shared.WrapInternal(err, "failed to delete post")
	}

	return nil
}
