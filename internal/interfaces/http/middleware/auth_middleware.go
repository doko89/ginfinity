package middleware

import (
	"net/http"
	"strings"

	"gin-boilerplate/internal/application/dto"
	"gin-boilerplate/internal/domain/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	tokenService service.TokenService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(tokenService service.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// RequireAuth middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "MISSING_TOKEN",
					Message: "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "INVALID_TOKEN_FORMAT",
					Message: "Authorization header must be in format: Bearer <token>",
				},
			})
			c.Abort()
			return
		}

		accessToken := tokenParts[1]

		// Validate access token
		claims, err := m.tokenService.ValidateAccessToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "INVALID_TOKEN",
					Message: "Invalid or expired access token",
				},
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// OptionalAuth middleware that optionally extracts user information if token is provided
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		accessToken := tokenParts[1]

		claims, err := m.tokenService.ValidateAccessToken(accessToken)
		if err != nil {
			c.Next()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}