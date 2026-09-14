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

// MemoryReadingRepository is an in-memory ReadingRepository used by unit tests.
type MemoryReadingRepository struct {
	mu       sync.Mutex
	Readings []Reading
}

// NewMemoryReadingRepository builds an empty in-memory repository.
func NewMemoryReadingRepository() *MemoryReadingRepository {
	return &MemoryReadingRepository{}
}

// ListRecentByOwner returns up to limit readings for ownerUserID with a
// timestamp at or after since, most recent first.
func (r *MemoryReadingRepository) ListRecentByOwner(_ context.Context, ownerUserID string, limit int32, since time.Time) ([]Reading, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	matches := make([]Reading, 0, len(r.Readings))
	for _, reading := range r.Readings {
		if reading.OwnerUserID != ownerUserID {
			continue
		}
		ts, err := time.Parse(time.RFC3339, reading.Timestamp)
		if err != nil || ts.Before(since) {
			continue
		}
		matches = append(matches, reading)
	}
	if int32(len(matches)) > limit {
		matches = matches[:limit]
	}
	return matches, nil
}
