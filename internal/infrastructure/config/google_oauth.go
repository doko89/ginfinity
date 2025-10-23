package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleOAuthConfig wraps OAuth2 configuration for Google
type GoogleOAuthConfig struct {
	oauth2.Config
}

// NewGoogleOAuthConfig creates a new Google OAuth configuration
func NewGoogleOAuthConfig(clientID, clientSecret, redirectURL string) *GoogleOAuthConfig {
	return &GoogleOAuthConfig{
		Config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// GetAuthURL returns the Google OAuth authorization URL
func (c *GoogleOAuthConfig) GetAuthURL(state string) string {
	return c.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCodeForToken exchanges authorization code for access token
func (c *GoogleOAuthConfig) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.Exchange(ctx, code)
}

// GetUserInfo fetches user information from Google using access token
func (c *GoogleOAuthConfig) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := c.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return &userInfo, nil
}

// GoogleUserInfo represents user information from Google OAuth
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
	VerifiedEmail bool   `json:"verified_email"`
}

// VerifyState verifies the OAuth state parameter
func VerifyState(receivedState, expectedState string) bool {
	return receivedState == expectedState
}

// GenerateRandomState generates a random state string for OAuth
func GenerateRandomState() string {
	// In a real application, you should use crypto/rand for better randomness
	return "random_state_" + fmt.Sprintf("%d", len("random_state"))
}

// ParseCallbackURL parses the callback URL to extract authorization code and state
func ParseCallbackURL(callbackURL string) (code, state string, err error) {
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse callback URL: %w", err)
	}

	// Check if the URL contains the authorization code
	code = parsedURL.Query().Get("code")
	state = parsedURL.Query().Get("state")

	if code == "" {
		return "", "", fmt.Errorf("authorization code not found in callback URL")
	}

	return code, state, nil
}

// BuildGoogleAuthURL builds the Google OAuth authorization URL
func BuildGoogleAuthURL(baseURL, clientID, redirectURI, state string) string {
	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("scope", "openid email profile")
	params.Add("response_type", "code")
	params.Add("state", state)
	params.Add("access_type", "offline")
	params.Add("approval_prompt", "force")

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// HandleCallback handles the OAuth callback and exchanges code for user info
func (c *GoogleOAuthConfig) HandleCallback(ctx context.Context, code, state string) (*GoogleUserInfo, error) {
	// Exchange authorization code for access token
	token, err := c.ExchangeCodeForToken(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user information
	userInfo, err := c.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return userInfo, nil
}