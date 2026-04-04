package post

import (
	"context"
	"testing"
	"time"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
)

func TestCreatePostExecute(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)
	posts := &mockPostWriter{}
	users := &mockUserReader{
		user: &entities.User{
			ID:           "user-1",
			Name:         "Farshid",
			Email:        "farshid@example.com",
			PasswordHash: "hashed",
			CreatedAt:    now,
		},
	}
	uow := &mockUnitOfWork{}

	uc := NewCreatePost(posts, users, uow, fixedIDGenerator{id: "post-1"}, fixedClock{now: now})

	postEntity, err := uc.Execute(context.Background(), CreatePostInput{
		AuthorID:   "user-1",
		Title:      "Clean Architecture",
		Content:    "Boundaries keep complexity contained.",
		PublishNow: true,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !uow.called {
		t.Fatal("expected transaction boundary to be used")
	}
	if posts.created == nil {
		t.Fatal("expected post to be persisted")
	}
	if postEntity.Status != entities.PostStatusPublished {
		t.Fatalf("expected status %q, got %q", entities.PostStatusPublished, postEntity.Status)
	}
	if postEntity.PublishedAt == nil {
		t.Fatal("expected PublishedAt to be set")
	}
}

type mockPostWriter struct {
	created *entities.Post
}

func (m *mockPostWriter) Create(_ context.Context, post *entities.Post) error {
	copyValue := *post
	m.created = &copyValue
	return nil
}

func (m *mockPostWriter) Update(context.Context, *entities.Post) error {
	return nil
}

func (m *mockPostWriter) Delete(context.Context, string) error {
	return nil
}

func (m *mockPostWriter) AddLike(context.Context, string, string) error {
	return nil
}

type mockUserReader struct {
	user *entities.User
}

func (m *mockUserReader) GetByID(context.Context, string) (*entities.User, error) {
	return m.user, nil
}

func (m *mockUserReader) GetByEmail(context.Context, string) (*entities.User, error) {
	return m.user, nil
}

func (m *mockUserReader) List(context.Context, repositories.Pagination) (repositories.PageResult[*entities.User], error) {
	return repositories.PageResult[*entities.User]{Items: []*entities.User{m.user}, Page: 1, Limit: 1, Total: 1}, nil
}

type mockUnitOfWork struct {
	called bool
}

func (m *mockUnitOfWork) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	m.called = true
	return fn(ctx)
}

type fixedIDGenerator struct {
	id string
}

func (g fixedIDGenerator) NewString() string {
	return g.id
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}
