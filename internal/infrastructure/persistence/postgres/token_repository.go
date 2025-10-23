package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gin-boilerplate/internal/domain/entity"
	"gin-boilerplate/internal/domain/repository"

	"gorm.io/gorm"
)

type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new PostgreSQL token repository
func NewTokenRepository(db *gorm.DB) repository.TokenRepository {
	return &tokenRepository{
		db: db,
	}
}

// Create creates a new refresh token
func (r *tokenRepository) Create(ctx context.Context, token *entity.Token) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}
	return nil
}

// FindByRefreshToken finds a token by refresh token
func (r *tokenRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*entity.Token, error) {
	var token entity.Token
	if err := r.db.WithContext(ctx).Where("refresh_token = ?", refreshToken).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find token by refresh token: %w", err)
	}
	return &token, nil
}

// FindByUserID finds tokens by user ID
func (r *tokenRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Token, error) {
	var tokens []*entity.Token
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to find tokens by user ID: %w", err)
	}
	return tokens, nil
}

// Update updates a token
func (r *tokenRepository) Update(ctx context.Context, token *entity.Token) error {
	if err := r.db.WithContext(ctx).Save(token).Error; err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}
	return nil
}

// Delete deletes a token by ID
func (r *tokenRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Token{}).Error; err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

// DeleteByRefreshToken deletes a token by refresh token
func (r *tokenRepository) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	if err := r.db.WithContext(ctx).
		Where("refresh_token = ?", refreshToken).
		Delete(&entity.Token{}).Error; err != nil {
		return fmt.Errorf("failed to delete token by refresh token: %w", err)
	}
	return nil
}

// DeleteByUserID deletes all tokens for a user (logout from all devices)
func (r *tokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.Token{}).Error; err != nil {
		return fmt.Errorf("failed to delete tokens by user ID: %w", err)
	}
	return nil
}

// DeleteExpiredTokens deletes all expired tokens
func (r *tokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.Token{}).Error; err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}
	return nil
}

// RevokeToken revokes a token by setting expiration to past
func (r *tokenRepository) RevokeToken(ctx context.Context, refreshToken string) error {
	if err := r.db.WithContext(ctx).
		Model(&entity.Token{}).
		Where("refresh_token = ?", refreshToken).
		Update("expires_at", time.Now().Add(-1*time.Hour)).Error; err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

// RevokeAllUserTokens revokes all tokens for a user
func (r *tokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).
		Model(&entity.Token{}).
		Where("user_id = ?", userID).
		Update("expires_at", time.Now().Add(-1*time.Hour)).Error; err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}
	return nil
}

// IsTokenValid checks if a refresh token is valid and not expired
func (r *tokenRepository) IsTokenValid(ctx context.Context, refreshToken string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Token{}).
		Where("refresh_token = ? AND expires_at > ?", refreshToken, time.Now()).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check token validity: %w", err)
	}
	return count > 0, nil
}