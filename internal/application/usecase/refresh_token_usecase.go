package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gin-boilerplate/internal/application/dto"
	"gin-boilerplate/internal/domain/entity"
	"gin-boilerplate/internal/domain/repository"
	"gin-boilerplate/internal/domain/service"
)

// RefreshTokenUseCase handles token refresh
type RefreshTokenUseCase struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.TokenRepository
	tokenService service.TokenService
}

// NewRefreshTokenUseCase creates a new refresh token use case
func NewRefreshTokenUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	tokenService service.TokenService,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenService: tokenService,
	}
}

// Execute executes the refresh token use case
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	// Validate refresh token
	claims, err := uc.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if refresh token exists in database and is valid
	isValid, err := uc.tokenRepo.IsTokenValid(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}
	if !isValid {
		return nil, errors.New("refresh token has been revoked or expired")
	}

	// Find user
	user, err := uc.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Delete old refresh token
	if err := uc.tokenRepo.DeleteByRefreshToken(ctx, req.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to delete old refresh token: %w", err)
	}

	// Generate new tokens
	accessToken, err := uc.tokenService.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := uc.tokenService.GenerateRefreshToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store new refresh token
	refreshTokenEntity := entity.NewToken(
		user.ID,
		newRefreshToken,
		time.Now().Add(uc.tokenService.GetTokenExpiration(service.TokenTypeRefresh)),
	)

	if err := uc.tokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// Calculate token expiration
	expiresIn := int64(uc.tokenService.GetTokenExpiration(service.TokenTypeAccess).Seconds())

	// Create response
	response := dto.ToAuthResponse(user, accessToken, newRefreshToken, expiresIn)

	return &response, nil
}