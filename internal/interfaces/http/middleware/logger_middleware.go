package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// responseBodyWriter is a wrapper around gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// LoggerMiddleware returns a logging middleware
func LoggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Capture response body
		w := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log request
		fields := logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"status":     c.Writer.Status(),
			"duration":   duration,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"size":       c.Writer.Size(),
		}

		// Add user information if available
		if userID, exists := c.Get("user_id"); exists {
			fields["user_id"] = userID
		}
		if userEmail, exists := c.Get("user_email"); exists {
			fields["user_email"] = userEmail
		}
		if userRole, exists := c.Get("user_role"); exists {
			fields["user_role"] = userRole
		}

		// Add request/response body for debugging in development
		if gin.Mode() == gin.DebugMode {
			if len(requestBody) > 0 && len(requestBody) < 1024 {
				fields["request_body"] = string(requestBody)
			}
			if w.body.Len() > 0 && w.body.Len() < 1024 {
				fields["response_body"] = w.body.String()
			}
		}

		// Log based on status code
		switch {
		case c.Writer.Status() >= 500:
			logger.WithFields(fields).Error("Internal server error")
		case c.Writer.Status() >= 400:
			logger.WithFields(fields).Warn("Client error")
		case c.Writer.Status() >= 300:
			logger.WithFields(fields).Info("Redirection")
		default:
			logger.WithFields(fields).Info("Request completed")
		}
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// In production, you might want to use UUID or more sophisticated method
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based random if crypto/rand fails
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		}
		return string(b)
	}
	return hex.EncodeToString(bytes)[:length]
}