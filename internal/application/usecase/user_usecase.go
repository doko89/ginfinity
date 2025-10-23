package usecase

import (
	"context"
	"fmt"

	"gin-boilerplate/internal/application/dto"
	"gin-boilerplate/internal/domain/repository"
)

// GetUserProfileUseCase handles getting user profile
type GetUserProfileUseCase struct {
	userRepo repository.UserRepository
}

// NewGetUserProfileUseCase creates a new get user profile use case
func NewGetUserProfileUseCase(userRepo repository.UserRepository) *GetUserProfileUseCase {
	return &GetUserProfileUseCase{
		userRepo: userRepo,
	}
}

// Execute executes the get user profile use case
func (uc *GetUserProfileUseCase) Execute(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

// UpdateUserProfileUseCase handles updating user profile
type UpdateUserProfileUseCase struct {
	userRepo repository.UserRepository
}

// NewUpdateUserProfileUseCase creates a new update user profile use case
func NewUpdateUserProfileUseCase(userRepo repository.UserRepository) *UpdateUserProfileUseCase {
	return &UpdateUserProfileUseCase{
		userRepo: userRepo,
	}
}

// Execute executes the update user profile use case
func (uc *UpdateUserProfileUseCase) Execute(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update profile
	user.UpdateProfile(req.Name, req.Avatar)

	// Validate updated user
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("invalid user data: %w", err)
	}

	// Save changes
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

// ListUsersUseCase handles listing users (admin only)
type ListUsersUseCase struct {
	userRepo repository.UserRepository
}

// NewListUsersUseCase creates a new list users use case
func NewListUsersUseCase(userRepo repository.UserRepository) *ListUsersUseCase {
	return &ListUsersUseCase{
		userRepo: userRepo,
	}
}

// Execute executes the list users use case
func (uc *ListUsersUseCase) Execute(ctx context.Context, req dto.PaginationRequest) (*dto.UsersListResponse, error) {
	// Set default pagination values
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 10
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get users and total count
	users, err := uc.userRepo.List(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := uc.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	response := dto.ToUsersListResponse(users, total, req.Limit, req.Offset)
	return &response, nil
}

// DeleteUserUseCase handles deleting a user (admin only)
type DeleteUserUseCase struct {
	userRepo repository.UserRepository
}

// NewDeleteUserUseCase creates a new delete user use case
func NewDeleteUserUseCase(userRepo repository.UserRepository) *DeleteUserUseCase {
	return &DeleteUserUseCase{
		userRepo: userRepo,
	}
}

// Execute executes the delete user use case
func (uc *DeleteUserUseCase) Execute(ctx context.Context, targetUserID string) error {
	// Check if user exists
	user, err := uc.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Delete user
	if err := uc.userRepo.Delete(ctx, targetUserID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// PromoteUserUseCase handles promoting a user to admin (admin only)
type PromoteUserUseCase struct {
	userRepo repository.UserRepository
}

// NewPromoteUserUseCase creates a new promote user use case
func NewPromoteUserUseCase(userRepo repository.UserRepository) *PromoteUserUseCase {
	return &PromoteUserUseCase{
		userRepo: userRepo,
	}
}

// Execute executes the promote user use case
func (uc *PromoteUserUseCase) Execute(ctx context.Context, targetUserID string) (*dto.UserResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.IsAdmin() {
		return nil, fmt.Errorf("user is already an admin")
	}

	// Promote user
	user.PromoteToAdmin()

	// Save changes
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to promote user: %w", err)
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

// DemoteUserUseCase handles demoting an admin to user (admin only)
type DemoteUserUseCase struct {
	userRepo repository.UserRepository
}

// NewDemoteUserUseCase creates a new demote user use case
func NewDemoteUserUseCase(userRepo repository.UserRepository) *DemoteUserUseCase {
	return &DemoteUserUseCase{
		userRepo: userRepo,
	}
}

// Execute executes the demote user use case
func (uc *DemoteUserUseCase) Execute(ctx context.Context, targetUserID string) (*dto.UserResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.IsAdmin() {
		return nil, fmt.Errorf("user is not an admin")
	}

	// Demote user
	user.DemoteToUser()

	// Save changes
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to demote user: %w", err)
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}