package post

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type UpdatePostInput struct {
	PostID    string
	ActorID   string
	ActorRole string
	Title     string
	Content   string
}

type UpdatePost struct {
	posts repositories.PostRepository
	clock shared.Clock
}

func NewUpdatePost(posts repositories.PostRepository, clock shared.Clock) *UpdatePost {
	return &UpdatePost{posts: posts, clock: clock}
}

func (uc *UpdatePost) Execute(ctx context.Context, input UpdatePostInput) (*entities.Post, error) {
	postEntity, err := uc.posts.GetByID(ctx, input.PostID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return nil, shared.WrapNotFound(err, "post was not found")
		}
		return nil, shared.WrapInternal(err, "failed to load post")
	}

	if !postEntity.CanBeEditedBy(input.ActorID, entities.UserRole(input.ActorRole)) {
		return nil, shared.WrapForbidden(entities.ErrForbidden, "only the author can update the post")
	}

	if err := postEntity.UpdateContent(input.Title, input.Content, uc.clock.Now()); err != nil {
		return nil, shared.WrapValidation(err, "post payload is invalid")
	}

	if err := uc.posts.Update(ctx, postEntity); err != nil {
		return nil, shared.WrapInternal(err, "failed to persist post")
	}

	return postEntity, nil
}
