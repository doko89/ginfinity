package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of token
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// TokenClaims represents JWT claims
type TokenClaims struct {
	UserID   string                 `json:"user_id"`
	Email    string                 `json:"email"`
	Role     string                 `json:"role"`
	TokenType TokenType             `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenService handles JWT token operations
type TokenService interface {
	// GenerateAccessToken generates an access token
	GenerateAccessToken(userID, email, role string) (string, error)

	// GenerateRefreshToken generates a refresh token
	GenerateRefreshToken(userID, email, role string) (string, error)

	// ValidateAccessToken validates an access token
	ValidateAccessToken(tokenString string) (*TokenClaims, error)

	// ValidateRefreshToken validates a refresh token
	ValidateRefreshToken(tokenString string) (*TokenClaims, error)

	// GetTokenExpiration returns the expiration time for a token type
	GetTokenExpiration(tokenType TokenType) time.Duration
}

type tokenService struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewTokenService creates a new token service
func NewTokenService(secretKey string, accessExpiry, refreshExpiry time.Duration) TokenService {
	return &tokenService{
		secretKey:     []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken generates an access token
func (s *tokenService) GenerateAccessToken(userID, email, role string) (string, error) {
	claims := &TokenClaims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GenerateRefreshToken generates a refresh token
func (s *tokenService) GenerateRefreshToken(userID, email, role string) (string, error) {
	claims := &TokenClaims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateAccessToken validates an access token
func (s *tokenService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	return s.validateToken(tokenString, TokenTypeAccess)
}

// ValidateRefreshToken validates a refresh token
func (s *tokenService) ValidateRefreshToken(tokenString string) (*TokenClaims, error) {
	return s.validateToken(tokenString, TokenTypeRefresh)
}

// validateToken validates a token and returns claims
func (s *tokenService) validateToken(tokenString string, expectedType TokenType) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}

// GetTokenExpiration returns the expiration time for a token type
func (s *tokenService) GetTokenExpiration(tokenType TokenType) time.Duration {
	switch tokenType {
	case TokenTypeAccess:
		return s.accessExpiry
	case TokenTypeRefresh:
		return s.refreshExpiry
	default:
		return s.accessExpiry
	}
}