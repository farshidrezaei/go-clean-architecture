package post

import (
	"context"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type ListPostsInput struct {
	Page  int
	Limit int
}

type ListPostsOutput struct {
	Items []*entities.Post
	Page  int
	Limit int
	Total int
}

type ListPosts struct {
	posts repositories.PostReader
}

func NewListPosts(posts repositories.PostReader) *ListPosts {
	return &ListPosts{posts: posts}
}

func (uc *ListPosts) Execute(ctx context.Context, input ListPostsInput) (*ListPostsOutput, error) {
	result, err := uc.posts.List(ctx, repositories.NewPagination(input.Page, input.Limit))
	if err != nil {
		return nil, shared.WrapInternal(err, "failed to load posts")
	}

	return &ListPostsOutput{
		Items: result.Items,
		Page:  result.Page,
		Limit: result.Limit,
		Total: result.Total,
	}, nil
}
