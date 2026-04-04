package comment

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type DeleteCommentInput struct {
	CommentID string
	ActorID   string
	ActorRole string
}

type DeleteComment struct {
	comments repositories.CommentRepository
}

func NewDeleteComment(comments repositories.CommentRepository) *DeleteComment {
	return &DeleteComment{comments: comments}
}

func (uc *DeleteComment) Execute(ctx context.Context, input DeleteCommentInput) error {
	commentEntity, err := uc.comments.GetByID(ctx, input.CommentID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return shared.WrapNotFound(err, "comment was not found")
		}
		return shared.WrapInternal(err, "failed to load comment")
	}

	if !commentEntity.CanBeManagedBy(input.ActorID, entities.UserRole(input.ActorRole)) {
		return shared.WrapForbidden(entities.ErrForbidden, "only the author can delete the comment")
	}

	if err := uc.comments.Delete(ctx, input.CommentID); err != nil {
		return shared.WrapInternal(err, "failed to delete comment")
	}

	return nil
}
