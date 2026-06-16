// Package jwt issues and verifies access tokens for students and admins.
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// PrincipalType distinguishes student vs admin tokens.
type PrincipalType string

const (
	Student PrincipalType = "student"
	Admin   PrincipalType = "admin"
)

// Claims are embedded in every access token.
type Claims struct {
	UserID      uuid.UUID     `json:"sub"`
	Type        PrincipalType `json:"type"`
	CollegeID   uuid.UUID     `json:"college_id"`
	Role        string        `json:"role,omitempty"`
	Permissions []string      `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// Manager creates and validates JWTs.
type Manager struct {
	accessSecret []byte
	accessTTL    time.Duration
}

func NewManager(accessSecret string, accessTTL time.Duration) *Manager {
	return &Manager{accessSecret: []byte(accessSecret), accessTTL: accessTTL}
}

// GenerateAccess signs a new access token for the given principal.
func (m *Manager) GenerateAccess(c Claims) (string, error) {
	now := time.Now()
	c.RegisteredClaims = jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		ID:        uuid.NewString(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return tok.SignedString(m.accessSecret)
}

// Parse validates a token string and returns its claims.
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.accessSecret, nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
