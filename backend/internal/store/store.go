// Package store provides persistence for authenticated user profiles.
package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a user profile does not exist.
var ErrNotFound = errors.New("store: user not found")

// User represents a minimal profile persisted for an OAuth-authenticated user.
type User struct {
	Provider   string    `json:"provider"`
	ProviderID string    `json:"provider_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	AvatarURL  string    `json:"avatar_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UserRepository persists and retrieves user profiles.
type UserRepository interface {
	// Upsert creates or updates a user profile, returning the stored record.
	Upsert(ctx context.Context, u User) (User, error)
	// Get retrieves a user profile by provider and provider-specific ID.
	Get(ctx context.Context, provider, providerID string) (User, error)
}
