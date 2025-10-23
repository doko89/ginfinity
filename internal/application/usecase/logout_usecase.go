package usecase

import (
	"context"
	"fmt"

	"gin-boilerplate/internal/domain/repository"
)

// LogoutUseCase handles user logout
type LogoutUseCase struct {
	tokenRepo repository.TokenRepository
}

// NewLogoutUseCase creates a new logout use case
func NewLogoutUseCase(tokenRepo repository.TokenRepository) *LogoutUseCase {
	return &LogoutUseCase{
		tokenRepo: tokenRepo,
	}
}

// Execute executes the logout use case (logout from current device)
func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	if err := uc.tokenRepo.DeleteByRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// ExecuteAll executes the logout from all devices
func (uc *LogoutUseCase) ExecuteAll(ctx context.Context, userID string) error {
	if err := uc.tokenRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	return nil
}