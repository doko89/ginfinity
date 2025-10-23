package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gin-boilerplate/internal/application/usecase"
	"gin-boilerplate/internal/domain/service"
	"gin-boilerplate/internal/infrastructure/config"
	"gin-boilerplate/internal/infrastructure/persistence/postgres"
	"gin-boilerplate/internal/infrastructure/redis"
	"gin-boilerplate/internal/infrastructure/storage"
	"gin-boilerplate/internal/infrastructure/redis"
	"gin-boilerplate/internal/interfaces/http/handler"
	httpmiddleware "gin-boilerplate/internal/interfaces/http/middleware"
	"gin-boilerplate/internal/interfaces/http/router"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	_ "gin-boilerplate/docs" // swagger docs
)

// @title Gin Boilerplate API
// @version 1.0
// @description A REST API boilerplate using Gin Framework with DDD architecture, authentication, multi-role authorization, S3-compatible file storage, and document management.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logger
	logger := setupLogger(cfg)

	logger.WithFields(logrus.Fields{
		"version": "1.0.0",
		"env":     cfg.Server.Env,
	}).Info("Starting Gin Boilerplate API")

	// Setup database
	db, err := postgres.NewDatabase(cfg.Database.DSN, cfg.IsDevelopment())
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.WithError(err).Error("Failed to close database connection")
		}
	}()

	// Check database health
	if err := db.Health(); err != nil {
		logger.WithError(err).Fatal("Database health check failed")
	}

	logger.Info("Database connection established successfully")

	// Setup domain services
	passwordService := service.NewPasswordService()
	tokenService := service.NewTokenService(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	// Setup Google OAuth configuration
	googleConfig := config.NewGoogleOAuthConfig(
		cfg.Google.ClientID,
		cfg.Google.ClientSecret,
		cfg.Google.RedirectURL,
	)

	// Setup S3 client
	s3Client, err := storage.NewS3Client(storage.S3Config{
		Endpoint:        cfg.S3.Endpoint,
		AccessKeyID:     cfg.S3.AccessKeyID,
		SecretAccessKey: cfg.S3.SecretAccessKey,
		Region:          cfg.S3.Region,
		Bucket:          cfg.S3.Bucket,
		UseSSL:          cfg.S3.UseSSL,
	})
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize S3 client")
	}

	// Setup Redis client
	redisClient, err := redis.NewRedisClient(redis.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis client")
	}
	defer redisClient.Close()

	// Setup repositories
	userRepo := postgres.NewUserRepository(db.GetDB())
	tokenRepo := postgres.NewTokenRepository(db.GetDB())
	documentRepo := postgres.NewDocumentRepository(db.GetDB())

	// Setup use cases
	registerUseCase := usecase.NewRegisterUseCase(userRepo, passwordService, tokenService)
	loginUseCase := usecase.NewLoginUseCase(userRepo, tokenRepo, passwordService, tokenService)
	refreshTokenUseCase := usecase.NewRefreshTokenUseCase(userRepo, tokenRepo, tokenService)
	logoutUseCase := usecase.NewLogoutUseCase(tokenRepo)
	googleAuthUseCase := usecase.NewGoogleAuthUseCase(userRepo, tokenRepo, tokenService)

	// User management use cases
	getUserProfileUseCase := usecase.NewGetUserProfileUseCase(userRepo)
	updateUserProfileUseCase := usecase.NewUpdateUserProfileUseCase(userRepo)
	listUsersUseCase := usecase.NewListUsersUseCase(userRepo)
	deleteUserUseCase := usecase.NewDeleteUserUseCase(userRepo)
	promoteUserUseCase := usecase.NewPromoteUserUseCase(userRepo)
	demoteUserUseCase := usecase.NewDemoteUserUseCase(userRepo)

	// Document management use cases
	documentUseCase := usecase.NewDocumentUseCase(documentRepo, s3Client)

	// Avatar management use cases
	avatarService := service.NewAvatarService(s3Client)
	avatarUseCase := usecase.NewAvatarUseCase(userRepo, avatarService, s3Client)

	// Setup handlers
	authHandler := handler.NewAuthHandler(
		registerUseCase,
		loginUseCase,
		refreshTokenUseCase,
		logoutUseCase,
		googleAuthUseCase,
		googleConfig,
	)

	userHandler := handler.NewUserHandler(
		getUserProfileUseCase,
		updateUserProfileUseCase,
		listUsersUseCase,
		deleteUserUseCase,
		promoteUserUseCase,
		demoteUserUseCase,
	)

	documentHandler := handler.NewDocumentHandler(documentUseCase)
	avatarHandler := handler.NewAvatarHandler(avatarUseCase)

	// Setup cache service and middleware
	cacheService := service.NewCacheService(redisClient)
	rateLimitMiddleware := httpmiddleware.NewRateLimitMiddleware(cacheService, httpmiddleware.RateLimitConfig{
		RequestsPerWindow: 100,
		WindowDuration:    time.Minute,
	})

	// Setup other middleware
	authMiddleware := httpmiddleware.NewAuthMiddleware(tokenService)
	roleMiddleware := httpmiddleware.NewRoleMiddleware()

	// Setup logger middleware
	loggerMiddleware := func() gin.HandlerFunc {
		return httpmiddleware.LoggerMiddleware(logger)
	}

	// Setup router
	router := router.NewRouter(
		authHandler,
		userHandler,
		documentHandler,
		avatarHandler,
		authMiddleware,
		roleMiddleware,
		rateLimitMiddleware,
		loggerMiddleware,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router.GetEngine(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("port", cfg.Server.Port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("Server shutdown completed")
	}
}

// setupLogger configures the application logger
func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	if cfg.IsDevelopment() {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	} else {
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	// Add file output in production
	if cfg.IsProduction() {
		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger.SetOutput(file)
		} else {
			logger.WithError(err).Warn("Failed to create log file, using stdout")
		}
	}

	return logger
}