package memory

import (
	"context"
	"sort"
	"strings"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
)

type UserRepository struct {
	store *Store
}

func NewUserRepository(store *Store) *UserRepository {
	return &UserRepository{store: store}
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	if snap, ok := snapshotFromContext(ctx); ok {
		if _, exists := snap.users[user.ID]; exists {
			return entities.ErrConflict
		}
		if _, exists := snap.byEmail[user.Email]; exists {
			return entities.ErrConflict
		}
		copyValue := *user
		snap.users[user.ID] = &copyValue
		snap.byEmail[user.Email] = user.ID
		return nil
	}

	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.users[user.ID]; exists {
		return entities.ErrConflict
	}
	if _, exists := r.store.byEmail[user.Email]; exists {
		return entities.ErrConflict
	}
	copyValue := *user
	r.store.users[user.ID] = &copyValue
	r.store.byEmail[user.Email] = user.ID
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	if snap, ok := snapshotFromContext(ctx); ok {
		user, exists := snap.users[id]
		if !exists {
			return nil, entities.ErrNotFound
		}
		copyValue := *user
		return &copyValue, nil
	}

	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	user, exists := r.store.users[id]
	if !exists {
		return nil, entities.ErrNotFound
	}
	copyValue := *user
	return &copyValue, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if snap, ok := snapshotFromContext(ctx); ok {
		id, exists := snap.byEmail[normalized]
		if !exists {
			return nil, entities.ErrNotFound
		}
		copyValue := *snap.users[id]
		return &copyValue, nil
	}

	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	id, exists := r.store.byEmail[normalized]
	if !exists {
		return nil, entities.ErrNotFound
	}
	copyValue := *r.store.users[id]
	return &copyValue, nil
}

func (r *UserRepository) List(ctx context.Context, pagination repositories.Pagination) (repositories.PageResult[*entities.User], error) {
	if snap, ok := snapshotFromContext(ctx); ok {
		return listUsersMap(snap.users, pagination), nil
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return listUsersMap(r.store.users, pagination), nil
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	if snap, ok := snapshotFromContext(ctx); ok {
		if _, exists := snap.users[user.ID]; !exists {
			return entities.ErrNotFound
		}
		copyValue := *user
		snap.users[user.ID] = &copyValue
		snap.byEmail[user.Email] = user.ID
		return nil
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if _, exists := r.store.users[user.ID]; !exists {
		return entities.ErrNotFound
	}
	copyValue := *user
	r.store.users[user.ID] = &copyValue
	r.store.byEmail[user.Email] = user.ID
	return nil
}

func listUsersMap(source map[string]*entities.User, pagination repositories.Pagination) repositories.PageResult[*entities.User] {
	items := make([]*entities.User, 0, len(source))
	for _, user := range source {
		copyValue := *user
		items = append(items, &copyValue)
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
	return repositories.PageResult[*entities.User]{Items: items[start:end], Page: pagination.Page, Limit: pagination.Limit, Total: total}
}
