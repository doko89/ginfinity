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

// LoginUseCase handles user login
type LoginUseCase struct {
	userRepo        repository.UserRepository
	tokenRepo       repository.TokenRepository
	passwordService service.PasswordService
	tokenService    service.TokenService
}

// NewLoginUseCase creates a new login use case
func NewLoginUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	passwordService service.PasswordService,
	tokenService service.TokenService,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		passwordService: passwordService,
		tokenService:    tokenService,
	}
}

// Execute executes the login use case
func (uc *LoginUseCase) Execute(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by email
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is OAuth user (no password)
	if user.IsOAuthUser() {
		return nil, errors.New("please use OAuth login for this account")
	}

	// Verify password
	if user.Password == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := uc.passwordService.VerifyPassword(req.Password, *user.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Revoke all existing refresh tokens for this user (single session)
	if err := uc.tokenRepo.RevokeAllUserTokens(ctx, user.ID); err != nil {
		// Log error but don't fail login
		// This could be logged in a real application
	}

	// Generate new tokens
	accessToken, err := uc.tokenService.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in database
	refreshTokenEntity := &entity.Token{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(uc.tokenService.GetTokenExpiration(service.TokenTypeRefresh)),
	}

	if err := uc.tokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Calculate token expiration
	expiresIn := int64(uc.tokenService.GetTokenExpiration(service.TokenTypeAccess).Seconds())

	// Create response
	response := dto.ToAuthResponse(user, accessToken, refreshToken, expiresIn)

	return &response, nil
}