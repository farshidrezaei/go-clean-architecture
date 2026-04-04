package memory

import (
	"context"
	"sync"
	"time"

	"clean_architecture/internal/domain/entities"
)

type Store struct {
	mu            sync.RWMutex
	users         map[string]*entities.User
	byEmail       map[string]string
	posts         map[string]*entities.Post
	comments      map[string]*entities.Comment
	likes         map[string]map[string]time.Time
	refreshTokens map[string]*entities.RefreshToken
}

func NewStore() *Store {
	return &Store{
		users:         make(map[string]*entities.User),
		byEmail:       make(map[string]string),
		posts:         make(map[string]*entities.Post),
		comments:      make(map[string]*entities.Comment),
		likes:         make(map[string]map[string]time.Time),
		refreshTokens: make(map[string]*entities.RefreshToken),
	}
}

type txKey struct{}

type snapshot struct {
	users         map[string]*entities.User
	byEmail       map[string]string
	posts         map[string]*entities.Post
	comments      map[string]*entities.Comment
	likes         map[string]map[string]time.Time
	refreshTokens map[string]*entities.RefreshToken
}

func withSnapshot(ctx context.Context, snap *snapshot) context.Context {
	return context.WithValue(ctx, txKey{}, snap)
}

func snapshotFromContext(ctx context.Context) (*snapshot, bool) {
	snap, ok := ctx.Value(txKey{}).(*snapshot)
	return snap, ok
}

func (s *Store) cloneLocked() *snapshot {
	users := make(map[string]*entities.User, len(s.users))
	for k, v := range s.users {
		copyValue := *v
		users[k] = &copyValue
	}

	byEmail := make(map[string]string, len(s.byEmail))
	for k, v := range s.byEmail {
		byEmail[k] = v
	}

	posts := make(map[string]*entities.Post, len(s.posts))
	for k, v := range s.posts {
		copyValue := *v
		posts[k] = &copyValue
	}

	comments := make(map[string]*entities.Comment, len(s.comments))
	for k, v := range s.comments {
		copyValue := *v
		comments[k] = &copyValue
	}

	likes := make(map[string]map[string]time.Time, len(s.likes))
	for postID, set := range s.likes {
		cloned := make(map[string]time.Time, len(set))
		for userID, createdAt := range set {
			cloned[userID] = createdAt
		}
		likes[postID] = cloned
	}

	refreshTokens := make(map[string]*entities.RefreshToken, len(s.refreshTokens))
	for k, v := range s.refreshTokens {
		copyValue := *v
		refreshTokens[k] = &copyValue
	}

	return &snapshot{users: users, byEmail: byEmail, posts: posts, comments: comments, likes: likes, refreshTokens: refreshTokens}
}

func (s *Store) commitLocked(snap *snapshot) {
	s.users = snap.users
	s.byEmail = snap.byEmail
	s.posts = snap.posts
	s.comments = snap.comments
	s.likes = snap.likes
	s.refreshTokens = snap.refreshTokens
}
