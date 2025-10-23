package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"gin-boilerplate/internal/application/usecase"

	"github.com/gin-gonic/gin"
)

type AvatarHandler struct {
	avatarUseCase *usecase.AvatarUseCase
}

func NewAvatarHandler(avatarUseCase *usecase.AvatarUseCase) *AvatarHandler {
	return &AvatarHandler{
		avatarUseCase: avatarUseCase,
	}
}

// UploadAvatar godoc
// @Summary Upload user avatar
// @Description Upload a new avatar image for the authenticated user
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar image file (max 2MB, supported: JPEG, PNG, GIF, WebP)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 413 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/avatar [post]
func (h *AvatarHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Avatar file is required"})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	validExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			validExt = true
			break
		}
	}

	if !validExt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Supported: JPEG, PNG, GIF, WebP"})
		return
	}

	// Upload avatar
	req := &usecase.UploadAvatarRequest{
		UserID: userID,
		File:   file,
	}

	apiURL, err := h.avatarUseCase.UploadAvatar(c.Request.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "file too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File too large (max 2MB)"})
			return
		}
		if strings.Contains(err.Error(), "invalid file type") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Avatar uploaded successfully",
		"avatar_url": *apiURL,
	})
}

// RemoveAvatar godoc
// @Summary Remove user avatar
// @Description Remove the current avatar image for the authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/avatar [delete]
func (h *AvatarHandler) RemoveAvatar(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err := h.avatarUseCase.RemoveAvatar(c.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "cannot remove Google") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove Google OAuth avatar"})
			return
		}
		if strings.Contains(err.Error(), "user not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if strings.Contains(err.Error(), "has no avatar") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User has no avatar to remove"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Avatar removed successfully",
	})
}

// ServeAvatar godoc
// @Summary Serve user avatar
// @Description Serve user avatar image (proxies S3 or redirects to Google)
// @Tags users
// @Produce image/jpeg,image/png,image/gif,image/webp
// @Param id path string true "User ID"
// @Success 200 {file} binary
// @Failure 302 {string} string "Redirect to Google avatar"
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/avatar/{id} [get]
func (h *AvatarHandler) ServeAvatar(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	avatarURL, _, err := h.avatarUseCase.ServeAvatar(c.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if strings.Contains(err.Error(), "has no avatar") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User has no avatar"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get avatar"})
		return
	}

	if avatarURL == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User has no avatar"})
		return
	}

	// If it's a Google avatar, redirect
	if strings.Contains(*avatarURL, "googleusercontent.com") {
		c.Redirect(http.StatusFound, *avatarURL)
		return
	}

	// For S3 avatars, redirect to presigned URL
	c.Redirect(http.StatusFound, *avatarURL)
}