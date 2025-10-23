package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService handles password-related operations
type PasswordService interface {
	// HashPassword hashes a password using bcrypt
	HashPassword(password string) (string, error)

	// VerifyPassword verifies a password against its hash
	VerifyPassword(password, hash string) error

	// ValidatePassword validates password strength
	ValidatePassword(password string) error
}

type passwordService struct {
	cost int
}

// NewPasswordService creates a new password service
func NewPasswordService() PasswordService {
	return &passwordService{
		cost: bcrypt.DefaultCost, // Can be configured
	}
}

// NewPasswordServiceWithCost creates a new password service with custom cost
func NewPasswordServiceWithCost(cost int) PasswordService {
	return &passwordService{
		cost: cost,
	}
}

// HashPassword hashes a password using bcrypt
func (s *passwordService) HashPassword(password string) (string, error) {
	if err := s.ValidatePassword(password); err != nil {
		return "", fmt.Errorf("password validation failed: %w", err)
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash
func (s *passwordService) VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidatePassword validates password strength
func (s *passwordService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be less than 128 characters long")
	}

	// Can add more validation rules here
	// - At least one uppercase letter
	// - At least one lowercase letter
	// - At least one number
	// - At least one special character

	return nil
}