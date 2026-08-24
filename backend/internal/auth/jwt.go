// Package auth implements OAuth2 login flows (Google, GitHub) and issuance
// of the application's own JWT session tokens.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when a JWT fails validation.
var ErrInvalidToken = errors.New("auth: invalid token")

// Claims is the payload embedded in application-issued JWTs.
type Claims struct {
	jwt.RegisteredClaims
	Provider  string `json:"provider"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// TokenIssuer issues and verifies application JWTs signed with HS256.
type TokenIssuer struct {
	signingKey []byte
	ttl        time.Duration
	now        func() time.Time
}

// NewTokenIssuer builds a TokenIssuer using the given signing key and TTL.
func NewTokenIssuer(signingKey string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{signingKey: []byte(signingKey), ttl: ttl, now: time.Now}
}

// Issue creates a signed JWT for the given subject (provider#providerID) and claims.
func (t *TokenIssuer) Issue(subject string, claims Claims) (string, error) {
	now := t.now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.signingKey)
}

// Verify parses and validates a signed JWT, returning its claims.
func (t *TokenIssuer) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return t.signingKey, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
