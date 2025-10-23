package domain

import "errors"

// Document errors
var (
	ErrDocumentNotFound       = errors.New("document not found")
	ErrDocumentTitleRequired  = errors.New("document title is required")
	ErrDocumentFileURLRequired = errors.New("document file URL is required")
	ErrDocumentUserIDRequired = errors.New("document user ID is required")
	ErrFileUploadFailed       = errors.New("file upload failed")
	ErrInvalidFileType        = errors.New("invalid file type")
	ErrFileTooLarge           = errors.New("file too large")
)