package post

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/usecase/shared"
)

type CreatePostInput struct {
	AuthorID   string
	Title      string
	Content    string
	PublishNow bool
}

type CreatePost struct {
	posts repositories.PostWriter
	users repositories.UserReader
	uow   repositories.UnitOfWork
	ids   shared.IDGenerator
	clock shared.Clock
}

func NewCreatePost(posts repositories.PostWriter, users repositories.UserReader, uow repositories.UnitOfWork, ids shared.IDGenerator, clock shared.Clock) *CreatePost {
	return &CreatePost{posts: posts, users: users, uow: uow, ids: ids, clock: clock}
}

func (uc *CreatePost) Execute(ctx context.Context, input CreatePostInput) (*entities.Post, error) {
	var created *entities.Post

	err := uc.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		if _, err := uc.users.GetByID(txCtx, input.AuthorID); err != nil {
			if errors.Is(err, entities.ErrNotFound) {
				return shared.WrapNotFound(err, "author was not found")
			}
			return shared.WrapInternal(err, "failed to load author")
		}

		postEntity, err := entities.NewPost(uc.ids.NewString(), input.AuthorID, input.Title, input.Content, uc.clock.Now())
		if err != nil {
			return shared.WrapValidation(err, "post payload is invalid")
		}

		if input.PublishNow {
			if err := postEntity.Publish(uc.clock.Now()); err != nil {
				return shared.WrapValidation(err, "post could not be published")
			}
		}

		if err := uc.posts.Create(txCtx, postEntity); err != nil {
			return shared.WrapInternal(err, "failed to persist post")
		}

		created = postEntity
		return nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}
