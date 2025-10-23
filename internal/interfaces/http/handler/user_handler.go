package handler

import (
	"net/http"
	"strconv"

	"gin-boilerplate/internal/application/dto"
	"gin-boilerplate/internal/application/usecase"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related endpoints
type UserHandler struct {
	getProfileUseCase  *usecase.GetUserProfileUseCase
	updateProfileUseCase *usecase.UpdateUserProfileUseCase
	listUsersUseCase   *usecase.ListUsersUseCase
	deleteUserUseCase  *usecase.DeleteUserUseCase
	promoteUserUseCase *usecase.PromoteUserUseCase
	demoteUserUseCase  *usecase.DemoteUserUseCase
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	getProfileUseCase *usecase.GetUserProfileUseCase,
	updateProfileUseCase *usecase.UpdateUserProfileUseCase,
	listUsersUseCase *usecase.ListUsersUseCase,
	deleteUserUseCase *usecase.DeleteUserUseCase,
	promoteUserUseCase *usecase.PromoteUserUseCase,
	demoteUserUseCase *usecase.DemoteUserUseCase,
) *UserHandler {
	return &UserHandler{
		getProfileUseCase:    getProfileUseCase,
		updateProfileUseCase: updateProfileUseCase,
		listUsersUseCase:     listUsersUseCase,
		deleteUserUseCase:    deleteUserUseCase,
		promoteUserUseCase:   promoteUserUseCase,
		demoteUserUseCase:    demoteUserUseCase,
	}
}

// GetMe handles getting current user profile
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

	response, err := h.getProfileUseCase.Execute(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "GET_PROFILE_FAILED",
				Message: "Failed to get user profile",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateMe handles updating current user profile
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	response, err := h.updateProfileUseCase.Execute(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "UPDATE_PROFILE_FAILED",
				Message: "Failed to update profile",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListUsers handles listing all users (admin only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	req := dto.PaginationRequest{}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	response, err := h.listUsersUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "LIST_USERS_FAILED",
				Message: "Failed to list users",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUser handles getting user by ID (admin only)
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_USER_ID",
				Message: "User ID is required",
			},
		})
		return
	}

	response, err := h.getProfileUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "USER_NOT_FOUND",
					Message: "User not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "GET_USER_FAILED",
				Message: "Failed to get user",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteUser handles deleting a user (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_USER_ID",
				Message: "User ID is required",
			},
		})
		return
	}

	err := h.deleteUserUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "USER_NOT_FOUND",
					Message: "User not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "DELETE_USER_FAILED",
				Message: "Failed to delete user",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "User deleted successfully",
	})
}

// PromoteUser handles promoting a user to admin (admin only)
func (h *UserHandler) PromoteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_USER_ID",
				Message: "User ID is required",
			},
		})
		return
	}

	response, err := h.promoteUserUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "USER_NOT_FOUND",
					Message: "User not found",
				},
			})
			return
		}

		if err.Error() == "user is already an admin" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "USER_ALREADY_ADMIN",
					Message: "User is already an admin",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "PROMOTE_USER_FAILED",
				Message: "Failed to promote user",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DemoteUser handles demoting an admin to user (admin only)
func (h *UserHandler) DemoteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_USER_ID",
				Message: "User ID is required",
			},
		})
		return
	}

	response, err := h.demoteUserUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "USER_NOT_FOUND",
					Message: "User not found",
				},
			})
			return
		}

		if err.Error() == "user is not an admin" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "USER_NOT_ADMIN",
					Message: "User is not an admin",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "DEMOTE_USER_FAILED",
				Message: "Failed to demote user",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}