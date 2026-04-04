package post

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type LikePostInput struct {
	PostID  string
	ActorID string
}

type LikePost struct {
	posts repositories.PostRepository
	uow   repositories.UnitOfWork
}

func NewLikePost(posts repositories.PostRepository, uow repositories.UnitOfWork) *LikePost {
	return &LikePost{posts: posts, uow: uow}
}

func (uc *LikePost) Execute(ctx context.Context, input LikePostInput) error {
	return uc.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		postEntity, err := uc.posts.GetByID(txCtx, input.PostID)
		if err != nil {
			if errors.Is(err, entities.ErrNotFound) {
				return shared.WrapNotFound(err, "post was not found")
			}
			return shared.WrapInternal(err, "failed to load post")
		}

		postEntity.Like()
		if err := uc.posts.AddLike(txCtx, input.PostID, input.ActorID); err != nil {
			if errors.Is(err, entities.ErrAlreadyLiked) {
				return shared.WrapConflict(err, "user already liked this post")
			}
			return shared.WrapInternal(err, "failed to add like")
		}
		if err := uc.posts.Update(txCtx, postEntity); err != nil {
			return shared.WrapInternal(err, "failed to update like count")
		}

		return nil
	})
}
