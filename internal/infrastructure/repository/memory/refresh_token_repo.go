package memory

import (
	"context"
	"sort"
	"time"

	"clean_architecture/internal/domain/entities"
)

type RefreshTokenRepository struct {
	store *Store
}

func NewRefreshTokenRepository(store *Store) *RefreshTokenRepository {
	return &RefreshTokenRepository{store: store}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	target, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		target = r.store.refreshTokens
	}
	if _, exists := target[token.TokenHash]; exists {
		return entities.ErrConflict
	}
	copyValue := *token
	target[token.TokenHash] = &copyValue
	return nil
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	target, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.RLock()
		defer r.store.mu.RUnlock()
	}
	token, exists := target[tokenHash]
	if !exists {
		return nil, entities.ErrNotFound
	}
	copyValue := *token
	return &copyValue, nil
}

func (r *RefreshTokenRepository) GetByID(ctx context.Context, id string) (*entities.RefreshToken, error) {
	target, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.RLock()
		defer r.store.mu.RUnlock()
	}
	for _, token := range target {
		if token.ID == id {
			copyValue := *token
			return &copyValue, nil
		}
	}
	return nil, entities.ErrNotFound
}

func (r *RefreshTokenRepository) ListByUserID(ctx context.Context, userID string) ([]*entities.RefreshToken, error) {
	target, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.RLock()
		defer r.store.mu.RUnlock()
	}
	items := make([]*entities.RefreshToken, 0)
	for _, token := range target {
		if token.UserID == userID {
			copyValue := *token
			items = append(items, &copyValue)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	target, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		target = r.store.refreshTokens
	}
	for _, token := range target {
		if token.ID == id {
			token.Revoke(revokedAt)
			return nil
		}
	}
	return entities.ErrNotFound
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, id, replacedByID string, revokedAt time.Time) error {
	target, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		target = r.store.refreshTokens
	}
	for _, token := range target {
		if token.ID == id {
			token.Rotate(replacedByID, revokedAt)
			return nil
		}
	}
	return entities.ErrNotFound
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string, revokedAt time.Time) error {
	target, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		target = r.store.refreshTokens
	}
	for _, token := range target {
		if token.UserID == userID && !token.IsRevoked() {
			token.Revoke(revokedAt)
		}
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeFamily(ctx context.Context, familyID string, revokedAt time.Time) error {
	target, readOnly := r.state(ctx)
	if !readOnly {
		r.store.mu.Lock()
		defer r.store.mu.Unlock()
		target = r.store.refreshTokens
	}
	for _, token := range target {
		if token.FamilyID == familyID && !token.IsRevoked() {
			token.Revoke(revokedAt)
			token.MarkCompromised(revokedAt)
		}
	}
	return nil
}

func (r *RefreshTokenRepository) state(ctx context.Context) (map[string]*entities.RefreshToken, bool) {
	if snap, ok := snapshotFromContext(ctx); ok {
		return snap.refreshTokens, true
	}
	return r.store.refreshTokens, false
}
