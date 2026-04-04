package ratelimit

import (
	"context"
	"sync"
	"time"
)

type Service interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type MemoryService struct {
	mu      sync.Mutex
	entries map[string]*entry
}

type entry struct {
	count     int
	expiresAt time.Time
}

func NewMemoryService() *MemoryService {
	return &MemoryService{entries: make(map[string]*entry)}
}

func (s *MemoryService) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	current, exists := s.entries[key]
	if !exists || now.After(current.expiresAt) {
		s.entries[key] = &entry{count: 1, expiresAt: now.Add(window)}
		return true, nil
	}
	if current.count >= limit {
		return false, nil
	}
	current.count++
	return true, nil
}
