package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"gin-boilerplate/internal/domain/entity"
	"gin-boilerplate/internal/domain"
	"gin-boilerplate/internal/domain/repository"
	"gin-boilerplate/internal/infrastructure/storage"
)

type DocumentUseCase struct {
	documentRepo repository.DocumentRepository
	storage      *storage.S3Client
}

func NewDocumentUseCase(documentRepo repository.DocumentRepository, storage *storage.S3Client) *DocumentUseCase {
	return &DocumentUseCase{
		documentRepo: documentRepo,
		storage:      storage,
	}
}

type UploadDocumentRequest struct {
	Title       string
	Description string
	File        *multipart.FileHeader
	UserID      string
}

type DocumentResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	FileURL     string `json:"file_url"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
	UserID      string `json:"user_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (uc *DocumentUseCase) UploadDocument(ctx context.Context, req *UploadDocumentRequest) (*DocumentResponse, error) {
	// Validate file size (max 10MB)
	const maxFileSize = 10 * 1024 * 1024
	if req.File.Size > maxFileSize {
		return nil, domain.ErrFileTooLarge
	}

	// Validate file type
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif", "application/pdf", "text/plain", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"}
	if !contains(allowedTypes, req.File.Header.Get("Content-Type")) {
		return nil, domain.ErrInvalidFileType
	}

	// Open the uploaded file
	file, err := req.File.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Upload file to S3
	fileURL, err := uc.storage.UploadFile(ctx, file, req.File.Filename, req.File.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrFileUploadFailed, err)
	}

	// Create document entity
	document := entity.NewDocument(
		req.Title,
		req.Description,
		*fileURL,
		req.File.Filename,
		req.File.Size,
		req.File.Header.Get("Content-Type"),
		req.UserID,
	)

	// Validate document
	if err := document.Validate(); err != nil {
		// If validation fails, try to delete the uploaded file
		if fileURL != nil {
			uc.storage.DeleteFile(ctx, *fileURL)
		}
		return nil, err
	}

	// Save document to database
	if err := uc.documentRepo.Create(ctx, document); err != nil {
		// If database save fails, try to delete the uploaded file
		if fileURL != nil {
			uc.storage.DeleteFile(ctx, *fileURL)
		}
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	return uc.toDocumentResponse(document), nil
}

func (uc *DocumentUseCase) GetDocument(ctx context.Context, id, userID string) (*DocumentResponse, error) {
	document, err := uc.documentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	// Check if user owns the document
	if document.UserID != userID {
		return nil, domain.ErrDocumentNotFound
	}

	return uc.toDocumentResponse(document), nil
}

func (uc *DocumentUseCase) GetUserDocuments(ctx context.Context, userID string, limit, offset int) ([]*DocumentResponse, error) {
	documents, err := uc.documentRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find user documents: %w", err)
	}

	responses := make([]*DocumentResponse, len(documents))
	for i, doc := range documents {
		responses[i] = uc.toDocumentResponse(doc)
	}

	return responses, nil
}

func (uc *DocumentUseCase) UpdateDocument(ctx context.Context, id, userID, title, description string) (*DocumentResponse, error) {
	document, err := uc.documentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	// Check if user owns the document
	if document.UserID != userID {
		return nil, domain.ErrDocumentNotFound
	}

	// Update document
	document.Update(title, description)

	// Validate updated document
	if err := document.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := uc.documentRepo.Update(ctx, document); err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return uc.toDocumentResponse(document), nil
}

func (uc *DocumentUseCase) DeleteDocument(ctx context.Context, id, userID string) error {
	document, err := uc.documentRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find document: %w", err)
	}

	// Check if user owns the document
	if document.UserID != userID {
		return domain.ErrDocumentNotFound
	}

	// Delete file from storage
	if err := uc.storage.DeleteFile(ctx, document.FileURL); err != nil {
		// Log error but continue with database deletion
		fmt.Printf("Warning: failed to delete file from storage: %v\n", err)
	}

	// Delete from database
	if err := uc.documentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

func (uc *DocumentUseCase) GetPresignedURL(ctx context.Context, id, userID string) (*string, error) {
	document, err := uc.documentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	// Check if user owns the document
	if document.UserID != userID {
		return nil, domain.ErrDocumentNotFound
	}

	// Generate presigned URL (valid for 1 hour)
	return uc.storage.GetPresignedURL(ctx, document.FileURL, time.Hour)
}

func (uc *DocumentUseCase) toDocumentResponse(doc *entity.Document) *DocumentResponse {
	return &DocumentResponse{
		ID:          doc.ID,
		Title:       doc.Title,
		Description: doc.Description,
		FileURL:     doc.FileURL,
		FileName:    doc.FileName,
		FileSize:    doc.FileSize,
		ContentType: doc.ContentType,
		UserID:      doc.UserID,
		CreatedAt:   doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   doc.UpdatedAt.Format(time.RFC3339),
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.ToLower(s) == strings.ToLower(item) {
			return true
		}
	}
	return false
}