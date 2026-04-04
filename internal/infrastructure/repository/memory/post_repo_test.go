package memory

import (
	"context"
	"testing"
	"time"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
)

func TestPostRepositoryListAndLike(t *testing.T) {
	t.Parallel()

	store := NewStore()
	repo := NewPostRepository(store)
	ctx := context.Background()

	first, err := entities.NewPost("post-1", "user-1", "First", "Alpha", time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewPost() error = %v", err)
	}
	second, err := entities.NewPost("post-2", "user-1", "Second", "Beta", time.Date(2026, 4, 4, 11, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewPost() error = %v", err)
	}

	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	if err := repo.Create(ctx, second); err != nil {
		t.Fatalf("Create(second) error = %v", err)
	}

	page, err := repo.List(ctx, repositories.NewPagination(1, 1))
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if page.Total != 2 {
		t.Fatalf("expected total 2, got %d", page.Total)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "post-2" {
		t.Fatalf("expected newest post first, got %+v", page.Items)
	}

	if err := repo.AddLike(ctx, "post-1", "user-9"); err != nil {
		t.Fatalf("AddLike() error = %v", err)
	}
	if err := repo.AddLike(ctx, "post-1", "user-9"); err != entities.ErrAlreadyLiked {
		t.Fatalf("expected ErrAlreadyLiked, got %v", err)
	}
}
