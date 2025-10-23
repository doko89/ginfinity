package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       string    `json:"user_id" gorm:"type:uuid;not null;index"`
	RefreshToken string    `json:"refresh_token" gorm:"type:text;not null;uniqueIndex"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewToken creates a new refresh token
func NewToken(userID, refreshToken string, expiresAt time.Time) *Token {
	return &Token{
		ID:           uuid.New().String(),
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Validate validates the token entity
func (t *Token) Validate() error {
	if t.UserID == "" {
		return errors.New("user ID is required")
	}

	if t.RefreshToken == "" {
		return errors.New("refresh token is required")
	}

	if t.ExpiresAt.IsZero() {
		return errors.New("expiration time is required")
	}

	if t.ExpiresAt.Before(time.Now()) {
		return errors.New("token has already expired")
	}

	return nil
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid checks if the token is valid (not expired)
func (t *Token) IsValid() bool {
	return !t.IsExpired()
}

// UpdateRefreshToken updates the refresh token and extends expiration
func (t *Token) UpdateRefreshToken(newRefreshToken string, newExpiresAt time.Time) {
	t.RefreshToken = newRefreshToken
	t.ExpiresAt = newExpiresAt
	t.UpdatedAt = time.Now()
}

// Revoke marks the token as revoked by setting expiration to past
func (t *Token) Revoke() {
	t.ExpiresAt = time.Now().Add(-1 * time.Hour)
	t.UpdatedAt = time.Now()
}