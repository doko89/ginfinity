package repository

import (
	"context"

	"gin-boilerplate/internal/domain/entity"
)

// TokenRepository defines the interface for token data operations
type TokenRepository interface {
	// Create creates a new refresh token
	Create(ctx context.Context, token *entity.Token) error

	// FindByRefreshToken finds a token by refresh token
	FindByRefreshToken(ctx context.Context, refreshToken string) (*entity.Token, error)

	// FindByUserID finds tokens by user ID
	FindByUserID(ctx context.Context, userID string) ([]*entity.Token, error)

	// Update updates a token
	Update(ctx context.Context, token *entity.Token) error

	// Delete deletes a token by ID
	Delete(ctx context.Context, id string) error

	// DeleteByRefreshToken deletes a token by refresh token
	DeleteByRefreshToken(ctx context.Context, refreshToken string) error

	// DeleteByUserID deletes all tokens for a user (logout from all devices)
	DeleteByUserID(ctx context.Context, userID string) error

	// DeleteExpiredTokens deletes all expired tokens
	DeleteExpiredTokens(ctx context.Context) error

	// RevokeToken revokes a token by setting expiration to past
	RevokeToken(ctx context.Context, refreshToken string) error

	// RevokeAllUserTokens revokes all tokens for a user
	RevokeAllUserTokens(ctx context.Context, userID string) error

	// IsTokenValid checks if a refresh token is valid and not expired
	IsTokenValid(ctx context.Context, refreshToken string) (bool, error)
}