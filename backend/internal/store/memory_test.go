package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryUserRepository_GetNotFound(t *testing.T) {
	repo := NewMemoryUserRepository()
	_, err := repo.Get(context.Background(), "google", "missing")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryUserRepository_UpsertThenGet(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	created, err := repo.Upsert(ctx, User{Provider: "google", ProviderID: "1", Email: "a@example.com", Name: "Ada"})
	require.NoError(t, err)
	require.False(t, created.CreatedAt.IsZero())
	require.Equal(t, created.CreatedAt, created.UpdatedAt)

	updated, err := repo.Upsert(ctx, User{Provider: "google", ProviderID: "1", Email: "a@example.com", Name: "Ada Lovelace"})
	require.NoError(t, err)
	require.Equal(t, created.CreatedAt, updated.CreatedAt, "CreatedAt must be preserved across updates")
	require.Equal(t, "Ada Lovelace", updated.Name)

	got, err := repo.Get(ctx, "google", "1")
	require.NoError(t, err)
	require.Equal(t, "Ada Lovelace", got.Name)
}
