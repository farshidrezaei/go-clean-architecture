package comment

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type AddCommentInput struct {
	PostID  string
	ActorID string
	Body    string
}

type AddComment struct {
	comments repositories.CommentWriter
	posts    repositories.PostReader
	users    repositories.UserReader
	uow      repositories.UnitOfWork
	ids      shared.IDGenerator
	clock    shared.Clock
}

func NewAddComment(comments repositories.CommentWriter, posts repositories.PostReader, users repositories.UserReader, uow repositories.UnitOfWork, ids shared.IDGenerator, clock shared.Clock) *AddComment {
	return &AddComment{comments: comments, posts: posts, users: users, uow: uow, ids: ids, clock: clock}
}

func (uc *AddComment) Execute(ctx context.Context, input AddCommentInput) (*entities.Comment, error) {
	var created *entities.Comment

	err := uc.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		if _, err := uc.users.GetByID(txCtx, input.ActorID); err != nil {
			if errors.Is(err, entities.ErrNotFound) {
				return shared.WrapNotFound(err, "author was not found")
			}
			return shared.WrapInternal(err, "failed to load author")
		}

		if _, err := uc.posts.GetByID(txCtx, input.PostID); err != nil {
			if errors.Is(err, entities.ErrNotFound) {
				return shared.WrapNotFound(err, "post was not found")
			}
			return shared.WrapInternal(err, "failed to load post")
		}

		commentEntity, err := entities.NewComment(uc.ids.NewString(), input.PostID, input.ActorID, input.Body, uc.clock.Now())
		if err != nil {
			return shared.WrapValidation(err, "comment payload is invalid")
		}

		if err := uc.comments.Create(txCtx, commentEntity); err != nil {
			return shared.WrapInternal(err, "failed to persist comment")
		}

		created = commentEntity
		return nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}
