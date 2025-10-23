package entity

import (
	"time"

	"gin-boilerplate/internal/domain"
	"github.com/google/uuid"
)

type Document struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FileURL     string    `json:"file_url"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewDocument(title, description, fileURL, fileName string, fileSize int64, contentType, userID string) *Document {
	now := time.Now()
	return &Document{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		FileURL:     fileURL,
		FileName:    fileName,
		FileSize:    fileSize,
		ContentType: contentType,
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (d *Document) Validate() error {
	if d.Title == "" {
		return domain.ErrDocumentTitleRequired
	}
	if d.FileURL == "" {
		return domain.ErrDocumentFileURLRequired
	}
	if d.UserID == "" {
		return domain.ErrDocumentUserIDRequired
	}
	return nil
}

func (d *Document) Update(title, description string) {
	d.Title = title
	d.Description = description
	d.UpdatedAt = time.Now()
}