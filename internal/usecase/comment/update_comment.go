package comment

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type UpdateCommentInput struct {
	CommentID string
	ActorID   string
	ActorRole string
	Body      string
}

type UpdateComment struct {
	comments repositories.CommentRepository
	clock    shared.Clock
}

func NewUpdateComment(comments repositories.CommentRepository, clock shared.Clock) *UpdateComment {
	return &UpdateComment{comments: comments, clock: clock}
}

func (uc *UpdateComment) Execute(ctx context.Context, input UpdateCommentInput) (*entities.Comment, error) {
	commentEntity, err := uc.comments.GetByID(ctx, input.CommentID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return nil, shared.WrapNotFound(err, "comment was not found")
		}
		return nil, shared.WrapInternal(err, "failed to load comment")
	}

	if !commentEntity.CanBeManagedBy(input.ActorID, entities.UserRole(input.ActorRole)) {
		return nil, shared.WrapForbidden(entities.ErrForbidden, "only the author can update the comment")
	}

	if err := commentEntity.UpdateBody(input.Body, uc.clock.Now()); err != nil {
		return nil, shared.WrapValidation(err, "comment payload is invalid")
	}

	if err := uc.comments.Update(ctx, commentEntity); err != nil {
		return nil, shared.WrapInternal(err, "failed to persist comment")
	}

	return commentEntity, nil
}
