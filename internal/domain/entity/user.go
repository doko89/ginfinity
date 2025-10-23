package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

type Provider string

const (
	ProviderLocal  Provider = "LOCAL"
	ProviderGoogle Provider = "GOOGLE"
)

type User struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	Password     *string   `json:"-" gorm:"null"` // nullable for OAuth users
	Name         string    `json:"name" gorm:"not null"`
	Role         Role      `json:"role" gorm:"type:varchar(10);default:'USER'"`
	Provider     Provider  `json:"provider" gorm:"type:varchar(10);default:'LOCAL'"`
	ProviderID   *string   `json:"-" gorm:"null"` // nullable for local users
	Avatar       *string   `json:"avatar" gorm:"null"`
	EmailVerified bool     `json:"email_verified" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewUser creates a new user instance
func NewUser(email, name string, role Role) *User {
	return &User{
		ID:            uuid.New().String(),
		Email:         strings.ToLower(strings.TrimSpace(email)),
		Name:          strings.TrimSpace(name),
		Role:          role,
		Provider:      ProviderLocal,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// NewOAuthUser creates a new user from OAuth provider
func NewOAuthUser(email, name, providerID string, provider Provider, avatar *string) *User {
	return &User{
		ID:            uuid.New().String(),
		Email:         strings.ToLower(strings.TrimSpace(email)),
		Name:          strings.TrimSpace(name),
		Role:          RoleUser,
		Provider:      provider,
		ProviderID:    &providerID,
		Avatar:        avatar,
		EmailVerified: true, // OAuth users are considered verified
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Validate validates the user entity
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}

	if !u.IsValidEmail() {
		return errors.New("invalid email format")
	}

	if u.Name == "" {
		return errors.New("name is required")
	}

	if len(u.Name) < 2 || len(u.Name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}

	// For local users, password is required
	if u.Provider == ProviderLocal && (u.Password == nil || *u.Password == "") {
		return errors.New("password is required for local users")
	}

	// For OAuth users, provider ID is required
	if u.Provider != ProviderLocal && (u.ProviderID == nil || *u.ProviderID == "") {
		return errors.New("provider ID is required for OAuth users")
	}

	return nil
}

// IsValidEmail checks if email format is valid
func (u *User) IsValidEmail() bool {
	email := u.Email
	return strings.Contains(email, "@") && strings.Contains(email, ".") &&
		len(email) > 5 && len(email) <= 255
}

// HasRole checks if user has specific role
func (u *User) HasRole(role Role) bool {
	return u.Role == role
}

// IsAdmin checks if user is admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsOAuthUser checks if user is from OAuth provider
func (u *User) IsOAuthUser() bool {
	return u.Provider != ProviderLocal
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(name string, avatar *string) {
	u.Name = strings.TrimSpace(name)
	u.Avatar = avatar
	u.UpdatedAt = time.Now()
}

// SetPassword sets the password for local users
func (u *User) SetPassword(hashedPassword string) {
	if u.Provider == ProviderLocal {
		u.Password = &hashedPassword
		u.UpdatedAt = time.Now()
	}
}

// VerifyEmail marks email as verified
func (u *User) VerifyEmail() {
	u.EmailVerified = true
	u.UpdatedAt = time.Now()
}

// PromoteToAdmin promotes user to admin role
func (u *User) PromoteToAdmin() {
	u.Role = RoleAdmin
	u.UpdatedAt = time.Now()
}

// DemoteToUser demotes admin to user role
func (u *User) DemoteToUser() {
	u.Role = RoleUser
	u.UpdatedAt = time.Now()
}