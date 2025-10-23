package router

import (
	"gin-boilerplate/internal/interfaces/http/handler"
	"gin-boilerplate/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

// Router wraps Gin router with all routes
type Router struct {
	engine *gin.Engine
}

// NewRouter creates a new router with all routes
func NewRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	authMiddleware *middleware.AuthMiddleware,
	roleMiddleware *middleware.RoleMiddleware,
	loggerMiddleware func() gin.HandlerFunc,
) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// Add global middleware
	engine.Use(gin.Recovery())
	engine.Use(loggerMiddleware())
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.RequestIDMiddleware())

	router := &Router{
		engine: engine,
	}

	router.setupRoutes(authHandler, userHandler, authMiddleware, roleMiddleware)

	return router
}

// setupRoutes configures all application routes
func (r *Router) setupRoutes(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	authMiddleware *middleware.AuthMiddleware,
	roleMiddleware *middleware.RoleMiddleware,
) {
	// Health check endpoint
	r.engine.GET("/health", r.healthCheck)

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := v1.Group("/")
		{
			r.setupPublicRoutes(public, authHandler)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(authMiddleware.RequireAuth())
		{
			r.setupProtectedRoutes(protected, authHandler, userHandler, roleMiddleware)
		}

		// Admin routes (admin role required)
		admin := v1.Group("/")
		admin.Use(authMiddleware.RequireAuth())
		admin.Use(roleMiddleware.RequireAdmin())
		{
			r.setupAdminRoutes(admin, userHandler)
		}
	}
}

// setupPublicRoutes configures public routes
func (r *Router) setupPublicRoutes(group *gin.RouterGroup, authHandler *handler.AuthHandler) {
	// Authentication routes
	auth := group.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.GET("/google", authHandler.GoogleAuth)
		auth.GET("/google/callback", authHandler.GoogleCallback)
	}
}

// setupProtectedRoutes configures protected routes
func (r *Router) setupProtectedRoutes(
	group *gin.RouterGroup,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	roleMiddleware *middleware.RoleMiddleware,
) {
	// Authentication routes (require valid token)
	auth := group.Group("/auth")
	{
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/logout-all", authHandler.LogoutAll)
	}

	// User routes (authenticated users)
	users := group.Group("/users")
	{
		// Current user endpoints
		users.GET("/me", userHandler.GetMe)
		users.PUT("/me", userHandler.UpdateMe)
	}
}

// setupAdminRoutes configures admin routes
func (r *Router) setupAdminRoutes(group *gin.RouterGroup, userHandler *handler.UserHandler) {
	// Admin user management
	users := group.Group("/users")
	{
		users.GET("", userHandler.ListUsers)           // List all users
		users.GET("/:id", userHandler.GetUser)         // Get user by ID
		users.DELETE("/:id", userHandler.DeleteUser)   // Delete user
		users.POST("/:id/promote", userHandler.PromoteUser) // Promote to admin
		users.POST("/:id/demote", userHandler.DemoteUser)   // Demote from admin
	}
}

// healthCheck returns server health status
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "ok",
		"timestamp": gin.H{},
		"version":   "1.0.0",
	})
}

// GetEngine returns the Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}