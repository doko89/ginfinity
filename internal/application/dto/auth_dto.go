package dto

import (
	"fmt"
	"strings"
	"gin-boilerplate/internal/domain/entity"
)

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
	Name     string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// GoogleAuthRequest represents Google OAuth callback request
type GoogleAuthRequest struct {
	Code  string `json:"code" binding:"required" example:"auth_code_from_google"`
	State string `json:"state" example:"random_state_string"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Name   string  `json:"name" binding:"omitempty,min=2,max=100" example:"John Doe"`
	Avatar *string `json:"avatar" example:"https://example.com/avatar.jpg"`
}

// AuthResponse represents authentication response with tokens
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string       `json:"token_type" example:"Bearer"`
	ExpiresIn    int64        `json:"expires_in" example:"900"`
}

// UserResponse represents user response
type UserResponse struct {
	ID            string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email         string  `json:"email" example:"user@example.com"`
	Name          string  `json:"name" example:"John Doe"`
	Role          string  `json:"role" example:"USER"`
	Provider      string  `json:"provider" example:"LOCAL"`
	Avatar        *string `json:"avatar" example:"https://example.com/avatar.jpg"`
	EmailVerified bool    `json:"email_verified" example:"true"`
	CreatedAt     string  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt     string  `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// UsersListResponse represents users list response
type UsersListResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
	Limit int            `json:"limit"`
	Offset int           `json:"offset"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error detail
type ErrorDetail struct {
	Code    string `json:"code" example:"INVALID_CREDENTIALS"`
	Message string `json:"message" example:"Email or password is incorrect"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationRequest represents pagination request
type PaginationRequest struct {
	Limit  int `json:"limit" form:"limit" example:"10"`
	Offset int `json:"offset" form:"offset" example:"0"`
}

// ToUserResponse converts entity.User to UserResponse
func ToUserResponse(user *entity.User) UserResponse {
	var avatarURL *string
	if user.Avatar != nil {
		// If it's a Google avatar, return as-is
		if isGoogleAvatar(*user.Avatar) {
			avatarURL = user.Avatar
		} else {
			// For S3 avatars, return API endpoint URL
			apiURL := fmt.Sprintf("/api/v1/users/avatar/%s", user.ID)
			avatarURL = &apiURL
		}
	}

	return UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Role:          string(user.Role),
		Provider:      string(user.Provider),
		Avatar:        avatarURL,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// isGoogleAvatar checks if avatar URL is from Google
func isGoogleAvatar(avatarURL string) bool {
	return strings.Contains(avatarURL, "googleusercontent.com") ||
		strings.Contains(avatarURL, "lh3.googleusercontent.com")
}

// ToUsersListResponse converts users slice to UsersListResponse
func ToUsersListResponse(users []*entity.User, total int64, limit, offset int) UsersListResponse {
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = ToUserResponse(user)
	}

	return UsersListResponse{
		Users:  userResponses,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}

// ToAuthResponse creates AuthResponse from user and tokens
func ToAuthResponse(user *entity.User, accessToken, refreshToken string, expiresIn int64) AuthResponse {
	return AuthResponse{
		User:         ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}
}