package dto

// UploadDocumentRequest represents a document upload request
type UploadDocumentRequest struct {
	Title       string                `form:"title" binding:"required" json:"title" example:"My Document"`
	Description string                `form:"description" json:"description" example:"A sample document"`
	File        string                `form:"file" binding:"required" json:"file" example:"document.pdf"`
}

// UpdateDocumentRequest represents a document update request
type UpdateDocumentRequest struct {
	Title       string `form:"title" binding:"required" json:"title" example:"Updated Document"`
	Description string `form:"description" json:"description" example:"Updated description"`
}

// DocumentResponse represents a document response
type DocumentResponse struct {
	ID          string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title       string `json:"title" example:"My Document"`
	Description string `json:"description" example:"A sample document"`
	FileURL     string `json:"file_url" example:"https://s3.amazonaws.com/bucket/uploads/file.pdf"`
	FileName    string `json:"file_name" example:"document.pdf"`
	FileSize    int64  `json:"file_size" example:"1024000"`
	ContentType string `json:"content_type" example:"application/pdf"`
	UserID      string `json:"user_id" example:"user123"`
	CreatedAt   string `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   string `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// PresignedURLResponse represents a presigned URL response
type PresignedURLResponse struct {
	URL     string `json:"url" example:"https://s3.amazonaws.com/bucket/file.pdf?signature=..."`
	Expires string `json:"expires" example:"2023-01-01T01:00:00Z"`
}