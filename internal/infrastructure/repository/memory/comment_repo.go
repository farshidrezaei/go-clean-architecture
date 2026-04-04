package memory

import (
	"context"
	"sort"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
)

type CommentRepository struct {
	store *Store
}

func NewCommentRepository(store *Store) *CommentRepository {
	return &CommentRepository{store: store}
}

func (r *CommentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	targetComments, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		targetComments = r.store.comments
	}
	if _, exists := targetComments[comment.ID]; exists {
		return entities.ErrConflict
	}
	copyValue := *comment
	targetComments[comment.ID] = &copyValue
	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id string) (*entities.Comment, error) {
	targetComments, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.RLock()
		defer r.store.mu.RUnlock()
	}
	comment, exists := targetComments[id]
	if !exists {
		return nil, entities.ErrNotFound
	}
	copyValue := *comment
	return &copyValue, nil
}

func (r *CommentRepository) Update(ctx context.Context, comment *entities.Comment) error {
	targetComments, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		targetComments = r.store.comments
	}
	if _, exists := targetComments[comment.ID]; !exists {
		return entities.ErrNotFound
	}
	copyValue := *comment
	targetComments[comment.ID] = &copyValue
	return nil
}

func (r *CommentRepository) Delete(ctx context.Context, id string) error {
	targetComments, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		targetComments = r.store.comments
	}
	if _, exists := targetComments[id]; !exists {
		return entities.ErrNotFound
	}
	delete(targetComments, id)
	return nil
}

func (r *CommentRepository) ListByPostID(ctx context.Context, postID string, pagination repositories.Pagination) (repositories.PageResult[*entities.Comment], error) {
	targetComments, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.RLock()
		defer r.store.mu.RUnlock()
	}

	items := make([]*entities.Comment, 0)
	for _, comment := range targetComments {
		if comment.PostID == postID {
			copyValue := *comment
			items = append(items, &copyValue)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	total := len(items)
	start := pagination.Offset
	if start > total {
		start = total
	}
	end := start + pagination.Limit
	if end > total {
		end = total
	}

	return repositories.PageResult[*entities.Comment]{
		Items: items[start:end],
		Page:  pagination.Page,
		Limit: pagination.Limit,
		Total: total,
	}, nil
}

func (r *CommentRepository) state(ctx context.Context) (map[string]*entities.Comment, bool) {
	if snap, ok := snapshotFromContext(ctx); ok {
		return snap.comments, true
	}
	return r.store.comments, false
}
