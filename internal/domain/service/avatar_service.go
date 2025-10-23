package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"gin-boilerplate/internal/domain"
	"gin-boilerplate/internal/infrastructure/storage"
)

type AvatarService struct {
	storage *storage.S3Client
}

func NewAvatarService(storage *storage.S3Client) *AvatarService {
	return &AvatarService{
		storage: storage,
	}
}

func (s *AvatarService) UploadAvatar(ctx context.Context, file *multipart.FileHeader, userID string) (*string, error) {
	// Validate file size (max 2MB for avatar)
	const maxAvatarSize = 2 * 1024 * 1024
	if file.Size > maxAvatarSize {
		return nil, domain.ErrFileTooLarge
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	allowedTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	if !s.contains(allowedTypes, contentType) {
		return nil, domain.ErrInvalidFileType
	}

	// Open the uploaded file
	fileReader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open avatar file: %w", err)
	}
	defer fileReader.Close()

	// Generate unique filename with user ID
	filename := s.generateAvatarFilename(file.Filename, userID)

	// Upload to S3
	fileURL, err := s.storage.UploadFile(ctx, fileReader, filename, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	return fileURL, nil
}

func (s *AvatarService) DeleteAvatar(ctx context.Context, avatarURL string) error {
	if avatarURL == "" {
		return nil // Nothing to delete
	}

	return s.storage.DeleteFile(ctx, avatarURL)
}

func (s *AvatarService) generateAvatarFilename(originalFilename, userID string) string {
	// Extract file extension
	ext := ""
	if dotIndex := strings.LastIndex(originalFilename, "."); dotIndex != -1 {
		ext = originalFilename[dotIndex:]
	}

	return fmt.Sprintf("avatars/%s/avatar%s", userID, ext)
}

func (s *AvatarService) contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.ToLower(s) == strings.ToLower(item) {
			return true
		}
	}
	return false
}