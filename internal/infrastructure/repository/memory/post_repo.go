package memory

import (
	"context"
	"sort"
	"time"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
)

type PostRepository struct {
	store *Store
}

func NewPostRepository(store *Store) *PostRepository {
	return &PostRepository{store: store}
}

func (r *PostRepository) Create(ctx context.Context, post *entities.Post) error {
	targetPosts, _, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		targetPosts = r.store.posts
	}
	if _, exists := targetPosts[post.ID]; exists {
		return entities.ErrConflict
	}
	copyValue := *post
	targetPosts[post.ID] = &copyValue
	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (*entities.Post, error) {
	targetPosts, _, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.RLock()
		defer r.store.mu.RUnlock()
	}
	post, exists := targetPosts[id]
	if !exists {
		return nil, entities.ErrNotFound
	}
	copyValue := *post
	return &copyValue, nil
}

func (r *PostRepository) Update(ctx context.Context, post *entities.Post) error {
	targetPosts, _, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		targetPosts = r.store.posts
	}
	if _, exists := targetPosts[post.ID]; !exists {
		return entities.ErrNotFound
	}
	copyValue := *post
	targetPosts[post.ID] = &copyValue
	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id string) error {
	if snap, ok := snapshotFromContext(ctx); ok {
		if _, exists := snap.posts[id]; !exists {
			return entities.ErrNotFound
		}
		delete(snap.posts, id)
		for commentID, comment := range snap.comments {
			if comment.PostID == id {
				delete(snap.comments, commentID)
			}
		}
		delete(snap.likes, id)
		return nil
	}

	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.posts[id]; !exists {
		return entities.ErrNotFound
	}
	delete(r.store.posts, id)
	for commentID, comment := range r.store.comments {
		if comment.PostID == id {
			delete(r.store.comments, commentID)
		}
	}
	delete(r.store.likes, id)
	return nil
}

func (r *PostRepository) AddLike(ctx context.Context, postID, userID string) error {
	_, likes, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		likes = r.store.likes
	}
	if _, exists := likes[postID]; !exists {
		likes[postID] = make(map[string]time.Time)
	}
	if _, exists := likes[postID][userID]; exists {
		return entities.ErrAlreadyLiked
	}
	likes[postID][userID] = time.Now().UTC()
	return nil
}

func (r *PostRepository) List(ctx context.Context, pagination repositories.Pagination) (repositories.PageResult[*entities.Post], error) {
	postsMap, _, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.RLock()
		defer r.store.mu.RUnlock()
	}

	items := make([]*entities.Post, 0, len(postsMap))
	for _, post := range postsMap {
		copyValue := *post
		items = append(items, &copyValue)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
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

	return repositories.PageResult[*entities.Post]{
		Items: items[start:end],
		Page:  pagination.Page,
		Limit: pagination.Limit,
		Total: total,
	}, nil
}

func (r *PostRepository) state(ctx context.Context) (map[string]*entities.Post, map[string]map[string]time.Time, bool) {
	if snap, ok := snapshotFromContext(ctx); ok {
		return snap.posts, snap.likes, true
	}
	return r.store.posts, r.store.likes, false
}
