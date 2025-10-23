package handler

import (
	"net/http"
	"strconv"
	"strings"

	"gin-boilerplate/internal/application/usecase"
	"gin-boilerplate/internal/interfaces/dto"

	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	documentUseCase *usecase.DocumentUseCase
}

func NewDocumentHandler(documentUseCase *usecase.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{
		documentUseCase: documentUseCase,
	}
}

// UploadDocument godoc
// @Summary Upload a new document
// @Description Upload a document with file
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Document title"
// @Param description formData string false "Document description"
// @Param file formData file true "Document file"
// @Security BearerAuth
// @Success 200 {object} dto.DocumentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 413 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/upload [post]
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get form data
	title := c.PostForm("title")
	description := c.PostForm("description")

	// Get file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Create upload request
	req := &usecase.UploadDocumentRequest{
		Title:       title,
		Description: description,
		File:        file,
		UserID:      userID,
	}

	document, err := h.documentUseCase.UploadDocument(c.Request.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "file too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File too large (max 10MB)"})
			return
		}
		if strings.Contains(err.Error(), "invalid file type") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload document"})
		return
	}

	c.JSON(http.StatusOK, document)
}

// GetDocument godoc
// @Summary Get a document by ID
// @Description Get a specific document by ID
// @Tags documents
// @Produce json
// @Param id path string true "Document ID"
// @Security BearerAuth
// @Success 200 {object} dto.DocumentResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id} [get]
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	document, err := h.documentUseCase.GetDocument(c.Request.Context(), documentID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get document"})
		return
	}

	c.JSON(http.StatusOK, document)
}

// GetUserDocuments godoc
// @Summary Get user's documents
// @Description Get all documents for the authenticated user
// @Tags documents
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents [get]
func (h *DocumentHandler) GetUserDocuments(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	documents, err := h.documentUseCase.GetUserDocuments(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get documents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
		"page":      page,
		"limit":     limit,
		"total":     len(documents),
	})
}

// UpdateDocument godoc
// @Summary Update a document
// @Description Update document title and description
// @Tags documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Param request body dto.UpdateDocumentRequest true "Update request"
// @Security BearerAuth
// @Success 200 {object} dto.DocumentResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id} [put]
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	var req dto.UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	document, err := h.documentUseCase.UpdateDocument(
		c.Request.Context(),
		documentID,
		userID,
		req.Title,
		req.Description,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document"})
		return
	}

	c.JSON(http.StatusOK, document)
}

// DeleteDocument godoc
// @Summary Delete a document
// @Description Delete a document and its file
// @Tags documents
// @Produce json
// @Param id path string true "Document ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id} [delete]
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	err := h.documentUseCase.DeleteDocument(c.Request.Context(), documentID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

// GetPresignedURL godoc
// @Summary Get presigned URL for document download
// @Description Get a temporary download URL for a document
// @Tags documents
// @Produce json
// @Param id path string true "Document ID"
// @Security BearerAuth
// @Success 200 {object} dto.PresignedURLResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /documents/{id}/download [get]
func (h *DocumentHandler) GetPresignedURL(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	url, err := h.documentUseCase.GetPresignedURL(c.Request.Context(), documentID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate download URL"})
		return
	}

	c.JSON(http.StatusOK, dto.PresignedURLResponse{
		URL: *url,
	})
}