package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTokenIssuer_IssueAndVerify(t *testing.T) {
	issuer := NewTokenIssuer("test-signing-key", time.Hour)

	token, err := issuer.Issue("google#123", Claims{
		Provider: "google",
		Email:    "user@example.com",
		Name:     "Test User",
	})
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := issuer.Verify(token)
	require.NoError(t, err)
	require.Equal(t, "google#123", claims.Subject)
	require.Equal(t, "google", claims.Provider)
	require.Equal(t, "user@example.com", claims.Email)
}

func TestTokenIssuer_Verify_RejectsExpiredToken(t *testing.T) {
	issuer := NewTokenIssuer("test-signing-key", -time.Minute)

	token, err := issuer.Issue("github#1", Claims{Provider: "github"})
	require.NoError(t, err)

	_, err = issuer.Verify(token)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestTokenIssuer_Verify_RejectsWrongKey(t *testing.T) {
	issuer := NewTokenIssuer("key-a", time.Hour)
	other := NewTokenIssuer("key-b", time.Hour)

	token, err := issuer.Issue("github#1", Claims{Provider: "github"})
	require.NoError(t, err)

	_, err = other.Verify(token)
	require.ErrorIs(t, err, ErrInvalidToken)
}

func TestTokenIssuer_Verify_RejectsGarbage(t *testing.T) {
	issuer := NewTokenIssuer("key", time.Hour)

	_, err := issuer.Verify("not-a-jwt")
	require.ErrorIs(t, err, ErrInvalidToken)
}
