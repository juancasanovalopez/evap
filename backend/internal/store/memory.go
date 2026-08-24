package store

import (
	"context"
	"sync"
	"time"
)

// MemoryUserRepository is an in-memory UserRepository used by unit tests.
type MemoryUserRepository struct {
	mu    sync.Mutex
	users map[string]User
	Now   func() time.Time
}

// NewMemoryUserRepository builds an empty in-memory repository.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{users: make(map[string]User), Now: time.Now}
}

// Get retrieves a user profile by provider and provider-specific ID.
func (r *MemoryUserRepository) Get(_ context.Context, provider, providerID string) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pk, _ := userKey(provider, providerID)
	u, ok := r.users[pk]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

// Upsert creates or updates a user profile, preserving the original CreatedAt.
func (r *MemoryUserRepository) Upsert(_ context.Context, u User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pk, _ := userKey(u.Provider, u.ProviderID)
	now := r.Now().UTC()
	if existing, ok := r.users[pk]; ok {
		u.CreatedAt = existing.CreatedAt
	} else {
		u.CreatedAt = now
	}
	u.UpdatedAt = now
	r.users[pk] = u
	return u, nil
}
