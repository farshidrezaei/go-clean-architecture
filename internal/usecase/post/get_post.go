package post

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type GetPostInput struct {
	PostID string
}

type GetPost struct {
	posts repositories.PostReader
}

func NewGetPost(posts repositories.PostReader) *GetPost {
	return &GetPost{posts: posts}
}

func (uc *GetPost) Execute(ctx context.Context, input GetPostInput) (*entities.Post, error) {
	postEntity, err := uc.posts.GetByID(ctx, input.PostID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return nil, shared.WrapNotFound(err, "post was not found")
		}
		return nil, shared.WrapInternal(err, "failed to load post")
	}

	return postEntity, nil
}
