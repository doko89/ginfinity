package repository

import (
	"context"

	"gin-boilerplate/internal/domain/entity"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) error

	// FindByID finds a user by ID
	FindByID(ctx context.Context, id string) (*entity.User, error)

	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// FindByProviderID finds a user by provider and provider ID
	FindByProviderID(ctx context.Context, provider entity.Provider, providerID string) (*entity.User, error)

	// Update updates a user
	Update(ctx context.Context, user *entity.User) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, id string) error

	// List returns a list of users with pagination
	List(ctx context.Context, limit, offset int) ([]*entity.User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)

	// EmailExists checks if email already exists
	EmailExists(ctx context.Context, email string) (bool, error)

	// FindByRole finds users by role
	FindByRole(ctx context.Context, role entity.Role, limit, offset int) ([]*entity.User, error)
}