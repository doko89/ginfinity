package usecase

import (
	"context"
	"errors"
	"fmt"

	"gin-boilerplate/internal/application/dto"
	"gin-boilerplate/internal/domain/entity"
	"gin-boilerplate/internal/domain/repository"
	"gin-boilerplate/internal/domain/service"
)

// RegisterUseCase handles user registration
type RegisterUseCase struct {
	userRepo        repository.UserRepository
	passwordService service.PasswordService
	tokenService    service.TokenService
}

// NewRegisterUseCase creates a new register use case
func NewRegisterUseCase(
	userRepo repository.UserRepository,
	passwordService service.PasswordService,
	tokenService service.TokenService,
) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo:        userRepo,
		passwordService: passwordService,
		tokenService:    tokenService,
	}
}

// Execute executes the register use case
func (uc *RegisterUseCase) Execute(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if email already exists
	exists, err := uc.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := uc.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := entity.NewUser(req.Email, req.Name, entity.RoleUser)
	user.SetPassword(hashedPassword)

	// Validate user
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("invalid user data: %w", err)
	}

	// Save user to database
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := uc.tokenService.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Calculate token expiration
	expiresIn := int64(uc.tokenService.GetTokenExpiration(service.TokenTypeAccess).Seconds())

	// Create response
	response := dto.ToAuthResponse(user, accessToken, refreshToken, expiresIn)

	return &response, nil
}