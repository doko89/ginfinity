package handler

import (
	"net/http"
	"strings"

	"gin-boilerplate/internal/application/dto"
	"gin-boilerplate/internal/application/usecase"
	"gin-boilerplate/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	registerUseCase  *usecase.RegisterUseCase
	loginUseCase     *usecase.LoginUseCase
	refreshUseCase   *usecase.RefreshTokenUseCase
	logoutUseCase    *usecase.LogoutUseCase
	googleAuthUseCase *usecase.GoogleAuthUseCase
	googleConfig     *config.GoogleOAuthConfig
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	registerUseCase *usecase.RegisterUseCase,
	loginUseCase *usecase.LoginUseCase,
	refreshUseCase *usecase.RefreshTokenUseCase,
	logoutUseCase *usecase.LogoutUseCase,
	googleAuthUseCase *usecase.GoogleAuthUseCase,
	googleConfig *config.GoogleOAuthConfig,
) *AuthHandler {
	return &AuthHandler{
		registerUseCase:   registerUseCase,
		loginUseCase:      loginUseCase,
		refreshUseCase:    refreshUseCase,
		logoutUseCase:     logoutUseCase,
		googleAuthUseCase: googleAuthUseCase,
		googleConfig:      googleConfig,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	response, err := h.registerUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "email already exists") {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "EMAIL_EXISTS",
					Message: "Email already exists",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "REGISTRATION_FAILED",
				Message: "Failed to register user",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	response, err := h.loginUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "INVALID_CREDENTIALS",
					Message: "Email or password is incorrect",
				},
			})
			return
		}

		if strings.Contains(err.Error(), "OAuth login") {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "OAUTH_REQUIRED",
					Message: "Please use OAuth login for this account",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "LOGIN_FAILED",
				Message: "Failed to login",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	response, err := h.refreshUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid refresh token") || strings.Contains(err.Error(), "revoked") {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorDetail{
					Code:    "INVALID_REFRESH_TOKEN",
					Message: "Invalid or expired refresh token",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "TOKEN_REFRESH_FAILED",
				Message: "Failed to refresh token",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	err := h.logoutUseCase.Execute(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "LOGOUT_FAILED",
				Message: "Failed to logout",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Successfully logged out",
	})
}

// LogoutAll handles logout from all devices
func (h *AuthHandler) LogoutAll(c *gin.Context) {
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

	err := h.logoutUseCase.ExecuteAll(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "LOGOUT_ALL_FAILED",
				Message: "Failed to logout from all devices",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Successfully logged out from all devices",
	})
}

// GoogleAuth redirects to Google OAuth
func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	state := config.GenerateRandomState()

	// Store state in session or cookie (simplified for this example)
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)

	authURL := h.googleConfig.GetAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleCallback handles Google OAuth callback
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Get state from cookie
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_STATE",
				Message: "OAuth state not found",
			},
		})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Verify state
	receivedState := c.Query("state")
	if !config.VerifyState(receivedState, stateCookie) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_STATE",
				Message: "Invalid OAuth state",
			},
		})
		return
	}

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "MISSING_CODE",
				Message: "Authorization code not found",
			},
		})
		return
	}

	// Exchange code for user info
	userInfo, err := h.googleConfig.HandleCallback(c.Request.Context(), code, receivedState)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "GOOGLE_AUTH_FAILED",
				Message: "Failed to authenticate with Google",
			},
		})
		return
	}

	// Convert to our GoogleUserInfo type
	googleUser := &usecase.GoogleUserInfo{
		ID:            userInfo.ID,
		Email:         userInfo.Email,
		Name:          userInfo.Name,
		Avatar:        userInfo.Picture,
		VerifiedEmail: userInfo.VerifiedEmail,
	}

	// Authenticate user
	response, err := h.googleAuthUseCase.Execute(c.Request.Context(), googleUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "GOOGLE_LOGIN_FAILED",
				Message: "Failed to login with Google",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}