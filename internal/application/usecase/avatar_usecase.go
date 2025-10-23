package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"gin-boilerplate/internal/domain/repository"
	"gin-boilerplate/internal/domain/service"
	"gin-boilerplate/internal/infrastructure/storage"
)

type AvatarUseCase struct {
	userRepo      repository.UserRepository
	avatarService *service.AvatarService
	storage       *storage.S3Client
}

func NewAvatarUseCase(userRepo repository.UserRepository, avatarService *service.AvatarService, storage *storage.S3Client) *AvatarUseCase {
	return &AvatarUseCase{
		userRepo:      userRepo,
		avatarService: avatarService,
		storage:       storage,
	}
}

type UploadAvatarRequest struct {
	UserID string
	File   *multipart.FileHeader
}

type UpdateAvatarRequest struct {
	UserID    string
	AvatarURL *string
}

func (uc *AvatarUseCase) UploadAvatar(ctx context.Context, req *UploadAvatarRequest) (*string, error) {
	// Find user
	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Upload new avatar to S3
	newAvatarURL, err := uc.avatarService.UploadAvatar(ctx, req.File, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Delete old avatar from S3 if exists
	if user.Avatar != nil && !uc.isGoogleAvatar(*user.Avatar) {
		if deleteErr := uc.avatarService.DeleteAvatar(ctx, *user.Avatar); deleteErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to delete old avatar: %v\n", deleteErr)
		}
	}

	// Update user avatar in database
	user.Avatar = newAvatarURL
	if err := uc.userRepo.Update(ctx, user); err != nil {
		// Try to rollback S3 upload
		if deleteErr := uc.avatarService.DeleteAvatar(ctx, *newAvatarURL); deleteErr != nil {
			fmt.Printf("Warning: failed to rollback avatar upload: %v\n", deleteErr)
		}
		return nil, fmt.Errorf("failed to update user avatar: %w", err)
	}

	// Return API endpoint URL instead of direct S3 URL
	apiURL := fmt.Sprintf("/api/v1/users/avatar/%s", req.UserID)
	return &apiURL, nil
}

func (uc *AvatarUseCase) RemoveAvatar(ctx context.Context, userID string) error {
	// Find user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Don't remove Google avatars
	if user.Avatar != nil && uc.isGoogleAvatar(*user.Avatar) {
		return fmt.Errorf("cannot remove Google OAuth avatar")
	}

	// Delete avatar from S3 if exists
	if user.Avatar != nil {
		if err := uc.avatarService.DeleteAvatar(ctx, *user.Avatar); err != nil {
			return fmt.Errorf("failed to delete avatar: %w", err)
		}
	}

	// Remove avatar URL from database
	user.Avatar = nil
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (uc *AvatarUseCase) GetAvatarURL(ctx context.Context, userID string) (*string, error) {
	// Find user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.Avatar == nil {
		return nil, nil // No avatar set
	}

	// Return Google avatar directly
	if uc.isGoogleAvatar(*user.Avatar) {
		return user.Avatar, nil
	}

	// Return API endpoint URL for S3 avatars
	apiURL := fmt.Sprintf("/api/v1/users/avatar/%s", userID)
	return &apiURL, nil
}

func (uc *AvatarUseCase) isGoogleAvatar(avatarURL string) bool {
	return strings.Contains(avatarURL, "googleusercontent.com") ||
		strings.Contains(avatarURL, "lh3.googleusercontent.com") ||
		strings.Contains(avatarURL, "graph.facebook.com")
}

func (uc *AvatarUseCase) ServeAvatar(ctx context.Context, userID string) (*string, *string, error) {
	// Find user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	if user.Avatar == nil {
		return nil, nil, fmt.Errorf("user has no avatar")
	}

	// Return redirect URL for Google avatars
	if uc.isGoogleAvatar(*user.Avatar) {
		return user.Avatar, nil, nil
	}

	// For S3 avatars, get presigned URL
	presignedURL, err := uc.storage.GetPresignedURL(ctx, *user.Avatar, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get avatar URL")
	}

	return presignedURL, nil, nil
}