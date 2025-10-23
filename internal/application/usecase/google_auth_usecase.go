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

// GoogleUserInfo represents user information from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Avatar        string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

// GoogleAuthUseCase handles Google OAuth authentication
type GoogleAuthUseCase struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.TokenRepository
	tokenService service.TokenService
}

// NewGoogleAuthUseCase creates a new Google auth use case
func NewGoogleAuthUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	tokenService service.TokenService,
) *GoogleAuthUseCase {
	return &GoogleAuthUseCase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenService: tokenService,
	}
}

// Execute executes the Google OAuth authentication
func (uc *GoogleAuthUseCase) Execute(ctx context.Context, googleUser *GoogleUserInfo) (*dto.AuthResponse, error) {
	if googleUser == nil {
		return nil, errors.New("google user info is required")
	}

	if !googleUser.VerifiedEmail {
		return nil, errors.New("email is not verified")
	}

	// Try to find existing user by Google ID first
	user, err := uc.userRepo.FindByProviderID(ctx, entity.ProviderGoogle, googleUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by provider ID: %w", err)
	}

	// If not found by provider ID, try by email (for merging accounts)
	if user == nil {
		user, err = uc.userRepo.FindByEmail(ctx, googleUser.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to find user by email: %w", err)
		}

		// If user exists with same email but different provider, merge accounts
		if user != nil && user.Provider != entity.ProviderGoogle {
			// Update user to include Google provider info
			user.Provider = entity.ProviderGoogle
			user.ProviderID = &googleUser.ID
			if googleUser.Avatar != "" {
				user.Avatar = &googleUser.Avatar
			}
			user.EmailVerified = true
			user.UpdatedAt = time.Now()

			if err := uc.userRepo.Update(ctx, user); err != nil {
				return nil, fmt.Errorf("failed to merge user account: %w", err)
			}
		}
	}

	// If user still doesn't exist, create new one
	if user == nil {
		var avatar *string
		if googleUser.Avatar != "" {
			avatar = &googleUser.Avatar
		}

		user = entity.NewOAuthUser(
			googleUser.Email,
			googleUser.Name,
			googleUser.ID,
			entity.ProviderGoogle,
			avatar,
		)

		if err := user.Validate(); err != nil {
			return nil, fmt.Errorf("invalid user data: %w", err)
		}

		if err := uc.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Revoke all existing refresh tokens for this user
	if err := uc.tokenRepo.RevokeAllUserTokens(ctx, user.ID); err != nil {
		// Log error but don't fail login
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
	refreshTokenEntity := entity.NewToken(
		user.ID,
		refreshToken,
		time.Now().Add(uc.tokenService.GetTokenExpiration(service.TokenTypeRefresh)),
	)

	if err := uc.tokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Calculate token expiration
	expiresIn := int64(uc.tokenService.GetTokenExpiration(service.TokenTypeAccess).Seconds())

	// Create response
	response := dto.ToAuthResponse(user, accessToken, refreshToken, expiresIn)

	return &response, nil
}