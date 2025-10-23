package middleware

import (
	"net/http"

	"gin-boilerplate/internal/application/dto"
	"gin-boilerplate/internal/domain/entity"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware handles role-based access control
type RoleMiddleware struct{}

// NewRoleMiddleware creates a new role middleware
func NewRoleMiddleware() *RoleMiddleware {
	return &RoleMiddleware{}
}

// RequireRole middleware that requires specific role
func (m *RoleMiddleware) RequireRole(role entity.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "UNAUTHORIZED",
					Message: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		if userRole.(string) != string(role) {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "INSUFFICIENT_PERMISSIONS",
					Message: "Insufficient permissions to access this resource",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin middleware that requires admin role
func (m *RoleMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole(entity.RoleAdmin)
}

// RequireUser middleware that requires user role (or higher)
func (m *RoleMiddleware) RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "UNAUTHORIZED",
					Message: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		// Both USER and ADMIN roles can access user endpoints
		role := userRole.(string)
		if role != string(entity.RoleUser) && role != string(entity.RoleAdmin) {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "INSUFFICIENT_PERMISSIONS",
					Message: "Insufficient permissions to access this resource",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// HasRole checks if user has specific role
func HasRole(c *gin.Context, role entity.Role) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}
	return userRole.(string) == string(role)
}

// IsAdmin checks if user is admin
func IsAdmin(c *gin.Context) bool {
	return HasRole(c, entity.RoleAdmin)
}

// IsUser checks if user has at least user role
func IsUser(c *gin.Context) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}
	role := userRole.(string)
	return role == string(entity.RoleUser) || role == string(entity.RoleAdmin)
}