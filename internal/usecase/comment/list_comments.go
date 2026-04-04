package comment

import (
	"context"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type ListCommentsInput struct {
	PostID string
	Page   int
	Limit  int
}

type ListCommentsOutput struct {
	Items []*entities.Comment
	Page  int
	Limit int
	Total int
}

type ListComments struct {
	comments repositories.CommentReader
}

func NewListComments(comments repositories.CommentReader) *ListComments {
	return &ListComments{comments: comments}
}

func (uc *ListComments) Execute(ctx context.Context, input ListCommentsInput) (*ListCommentsOutput, error) {
	result, err := uc.comments.ListByPostID(ctx, input.PostID, repositories.NewPagination(input.Page, input.Limit))
	if err != nil {
		return nil, shared.WrapInternal(err, "failed to load comments")
	}

	return &ListCommentsOutput{
		Items: result.Items,
		Page:  result.Page,
		Limit: result.Limit,
		Total: result.Total,
	}, nil
}
