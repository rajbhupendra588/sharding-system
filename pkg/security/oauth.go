package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"go.uber.org/zap"
)

// OAuthProvider represents an OAuth provider type
type OAuthProvider string

const (
	ProviderGoogle   OAuthProvider = "google"
	ProviderGitHub   OAuthProvider = "github"
	ProviderFacebook OAuthProvider = "facebook"
)

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ID       string
	Email    string
	Name     string
	Picture  string
	Provider OAuthProvider
}

// OAuthConfig holds OAuth configuration for all providers
type OAuthConfig struct {
	Google   *ProviderConfig
	GitHub   *ProviderConfig
	Facebook *ProviderConfig
	BaseURL  string // Base URL for redirect callbacks
	logger   *zap.Logger
}

// ProviderConfig holds configuration for a single OAuth provider
type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	Enabled      bool
}

// NewOAuthConfig creates a new OAuth configuration
func NewOAuthConfig(baseURL string, logger *zap.Logger) *OAuthConfig {
	return &OAuthConfig{
		BaseURL: baseURL,
		logger:  logger,
	}
}

// SetGoogleConfig sets Google OAuth configuration
func (c *OAuthConfig) SetGoogleConfig(clientID, clientSecret string) {
	c.Google = &ProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Enabled:      clientID != "" && clientSecret != "",
	}
}

// SetGitHubConfig sets GitHub OAuth configuration
func (c *OAuthConfig) SetGitHubConfig(clientID, clientSecret string) {
	c.GitHub = &ProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Enabled:      clientID != "" && clientSecret != "",
	}
}

// SetFacebookConfig sets Facebook OAuth configuration
func (c *OAuthConfig) SetFacebookConfig(clientID, clientSecret string) {
	c.Facebook = &ProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Enabled:      clientID != "" && clientSecret != "",
	}
}

// GetOAuth2Config returns the OAuth2 config for a provider
func (c *OAuthConfig) GetOAuth2Config(provider OAuthProvider) (*oauth2.Config, error) {
	var config *ProviderConfig
	var endpoint oauth2.Endpoint
	var scopes []string

	switch provider {
	case ProviderGoogle:
		if c.Google == nil || !c.Google.Enabled {
			return nil, errors.New("Google OAuth is not configured")
		}
		config = c.Google
		endpoint = google.Endpoint
		scopes = []string{"openid", "profile", "email"}

	case ProviderGitHub:
		if c.GitHub == nil || !c.GitHub.Enabled {
			return nil, errors.New("GitHub OAuth is not configured")
		}
		config = c.GitHub
		endpoint = github.Endpoint
		scopes = []string{"user:email"}

	case ProviderFacebook:
		if c.Facebook == nil || !c.Facebook.Enabled {
			return nil, errors.New("Facebook OAuth is not configured")
		}
		config = c.Facebook
		endpoint = facebook.Endpoint
		scopes = []string{"email", "public_profile"}

	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	return &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s/api/v1/auth/oauth/%s/callback", c.BaseURL, provider),
		Scopes:       scopes,
		Endpoint:     endpoint,
	}, nil
}

// GenerateState generates a random state token for OAuth flow
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthURL returns the authorization URL for a provider
func (c *OAuthConfig) GetAuthURL(provider OAuthProvider, state string) (string, error) {
	oauth2Config, err := c.GetOAuth2Config(provider)
	if err != nil {
		return "", err
	}
	return oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

// ExchangeCode exchanges an authorization code for a token
func (c *OAuthConfig) ExchangeCode(provider OAuthProvider, code string) (*oauth2.Token, error) {
	oauth2Config, err := c.GetOAuth2Config(provider)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return oauth2Config.Exchange(ctx, code)
}

// GetUserInfo retrieves user information from the OAuth provider
func (c *OAuthConfig) GetUserInfo(provider OAuthProvider, token *oauth2.Token) (*OAuthUserInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ts := oauth2.StaticTokenSource(token)
	client := oauth2.NewClient(ctx, ts)

	switch provider {
	case ProviderGoogle:
		return c.getGoogleUserInfo(client)
	case ProviderGitHub:
		return c.getGitHubUserInfo(client)
	case ProviderFacebook:
		return c.getFacebookUserInfoWithToken(client, token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// getGoogleUserInfo retrieves user info from Google
func (c *OAuthConfig) getGoogleUserInfo(client *http.Client) (*OAuthUserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &OAuthUserInfo{
		ID:       userInfo.ID,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Picture:  userInfo.Picture,
		Provider: ProviderGoogle,
	}, nil
}

// getGitHubUserInfo retrieves user info from GitHub
func (c *OAuthConfig) getGitHubUserInfo(client *http.Client) (*OAuthUserInfo, error) {
	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userInfo struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// If email is not public, fetch it from emails endpoint
	if userInfo.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.NewDecoder(emailResp.Body).Decode(&emails) == nil {
				for _, e := range emails {
					if e.Primary {
						userInfo.Email = e.Email
						break
					}
				}
				if userInfo.Email == "" && len(emails) > 0 {
					userInfo.Email = emails[0].Email
				}
			}
		}
	}

	if userInfo.Name == "" {
		userInfo.Name = userInfo.Login
	}

	return &OAuthUserInfo{
		ID:       fmt.Sprintf("%d", userInfo.ID),
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Picture:  userInfo.AvatarURL,
		Provider: ProviderGitHub,
	}, nil
}

// getFacebookUserInfoWithToken retrieves user info from Facebook using the token directly
func (c *OAuthConfig) getFacebookUserInfoWithToken(client *http.Client, token *oauth2.Token) (*OAuthUserInfo, error) {
	resp, err := client.Get(fmt.Sprintf("https://graph.facebook.com/me?fields=id,name,email,picture&access_token=%s", token.AccessToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &OAuthUserInfo{
		ID:       userInfo.ID,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Picture:  userInfo.Picture.Data.URL,
		Provider: ProviderFacebook,
	}, nil
}

// IsProviderEnabled checks if a provider is enabled
func (c *OAuthConfig) IsProviderEnabled(provider OAuthProvider) bool {
	switch provider {
	case ProviderGoogle:
		return c.Google != nil && c.Google.Enabled
	case ProviderGitHub:
		return c.GitHub != nil && c.GitHub.Enabled
	case ProviderFacebook:
		return c.Facebook != nil && c.Facebook.Enabled
	default:
		return false
	}
}

// GetEnabledProviders returns a list of enabled providers
func (c *OAuthConfig) GetEnabledProviders() []OAuthProvider {
	var providers []OAuthProvider
	if c.IsProviderEnabled(ProviderGoogle) {
		providers = append(providers, ProviderGoogle)
	}
	if c.IsProviderEnabled(ProviderGitHub) {
		providers = append(providers, ProviderGitHub)
	}
	if c.IsProviderEnabled(ProviderFacebook) {
		providers = append(providers, ProviderFacebook)
	}
	return providers
}

