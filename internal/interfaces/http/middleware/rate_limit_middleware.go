package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"gin-boilerplate/internal/domain/service"
)

type RateLimitConfig struct {
	RequestsPerWindow int
	WindowDuration    time.Duration
}

type RateLimitMiddleware struct {
	cacheService    *service.CacheService
	config        RateLimitConfig
}

func NewRateLimitMiddleware(cacheService *service.CacheService, config RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		cacheService: cacheService,
		config:      config,
	}
}

// RateLimiter tracks request counts per key
type RateLimiter struct {
	mu         sync.Mutex
	requests   int
	windowStart time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		windowStart: time.Now(),
	}
}

func (rl *RateLimiter) IsAllowed(config RateLimitConfig) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Reset if window has passed
	if now.Sub(rl.windowStart) >= config.WindowDuration {
		rl.requests = 0
		rl.windowStart = now
	}

	rl.requests++
	return rl.requests <= config.RequestsPerWindow
}

// RateLimit creates a rate limiting middleware
func (m *RateLimitMiddleware) RateLimit(identifier string) gin.HandlerFunc {
	// Create rate limiter for this identifier
	limiter := NewRateLimiter()

	return func(c *gin.Context) {
		// Get client identifier (IP, user ID, etc.)
		key := service.RateLimitCacheKey(identifier)

		// Try to increment counter in cache
		incremented, err := m.cacheService.SetNX(c.Request.Context(), key, 1, m.config.WindowDuration)
		if err != nil {
			// Log error but don't block the request
			c.Next()
			return
		}

		if !incremented {
			// Key exists, get current count
			countStr, err := m.cacheService.GetString(c.Request.Context(), key)
			if err != nil {
				c.Next()
				return
			}

			if countStr != "" {
				if count, err := strconv.Atoi(countStr); err == nil {
					if count > m.config.RequestsPerWindow {
						c.JSON(http.StatusTooManyRequests, gin.H{
							"error": "Rate limit exceeded",
							"retry_after": m.config.WindowDuration.Seconds(),
						})
						c.Abort()
						return
					}
				}
			}
		}

		// Allow the request
		c.Next()
	}
}

// RateLimitByIP creates rate limiting middleware by IP address
func (m *RateLimitMiddleware) RateLimitByIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		limiter := NewRateLimiter()

		key := service.RateLimitCacheKey("ip:" + clientIP)

		// Check current rate in cache
		countStr, err := m.cacheService.GetString(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		currentCount := 0
		if countStr != "" {
			if parsed, err := strconv.Atoi(countStr); err == nil {
				currentCount = parsed
			}
		}

		if currentCount >= m.config.RequestsPerWindow {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": m.config.WindowDuration.Seconds(),
			})
			c.Abort()
			return
		}

		// Increment counter
		if err := m.cacheService.Increment(c.Request.Context(), key); err != nil {
			// Log error but allow request
			c.Next()
			return
		}

		// Set expiry for the key if not already set
		if countStr == "" {
			m.cacheService.SetWithExpiration(
				c.Request.Context(),
				key,
				"1",
				m.config.WindowDuration,
			)
		}

		c.Next()
	}
}

// RateLimitByUser creates rate limiting middleware by user ID
func (m *RateLimitMiddleware) RateLimitByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		limiter := NewRateLimiter()
		key := service.RateLimitCacheKey("user:" + userID)

		// Check current rate in cache
		countStr, err := m.cacheService.GetString(c.Request.Context(), key)
		if err != nil {
			c.Next()
			return
		}

		currentCount := 0
		if countStr != "" {
			if parsed, err := strconv.Atoi(countStr); err == nil {
				currentCount = parsed
			}
		}

		if currentCount >= m.config.RequestsPerWindow {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": m.config.WindowDuration.Seconds(),
			})
			c.Abort()
			return
		}

		// Increment counter
		if err := m.cacheService.Increment(c.Request.Context(), key); err != nil {
			// Log error but allow request
			c.Next()
			return
		}

		// Set expiry for the key if not already set
		if countStr == "" {
			m.cacheService.SetWithExpiration(
				c.Request.Context(),
				key,
				"1",
				m.config.WindowDuration,
			)
		}

		c.Next()
	}
}