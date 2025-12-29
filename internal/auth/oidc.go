package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// OIDCConfig holds OIDC provider configuration
type OIDCConfig struct {
	Enabled      bool     `json:"enabled"`
	IssuerURL    string   `json:"issuer_url"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
}

// OIDCProvider represents an OIDC identity provider
type OIDCProvider struct {
	config       OIDCConfig
	wellKnown    *WellKnownConfig
	sessionStore map[string]*Session
}

// WellKnownConfig holds OIDC discovery document data
type WellKnownConfig struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"`
	JwksURI               string   `json:"jwks_uri"`
	ScopesSupported       []string `json:"scopes_supported"`
}

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	jwt.RegisteredClaims
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	UserID string `json:"user_id"`
}

// NewOIDCProvider creates a new OIDC provider
func NewOIDCProvider(config OIDCConfig) (*OIDCProvider, error) {
	provider := &OIDCProvider{
		config:       config,
		sessionStore: make(map[string]*Session),
	}

	if config.Enabled && config.IssuerURL != "" {
		if err := provider.discoverConfiguration(); err != nil {
			log.Warn().Err(err).Msg("Failed to discover OIDC configuration, will retry on first request")
		}
	}

	return provider, nil
}

// NewOIDCProviderFromEnv creates a provider from environment variables
func NewOIDCProviderFromEnv() (*OIDCProvider, error) {
	config := OIDCConfig{
		Enabled:      os.Getenv("OIDC_ENABLED") == "true",
		IssuerURL:    os.Getenv("OIDC_ISSUER_URL"),
		ClientID:     os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OIDC_REDIRECT_URL"),
		Scopes:       strings.Split(getEnvOrDefault("OIDC_SCOPES", "openid,profile,email"), ","),
	}
	return NewOIDCProvider(config)
}

func (p *OIDCProvider) discoverConfiguration() error {
	wellKnownURL := strings.TrimSuffix(p.config.IssuerURL, "/") + "/.well-known/openid-configuration"

	resp, err := http.Get(wellKnownURL)
	if err != nil {
		return fmt.Errorf("failed to fetch OIDC configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OIDC configuration returned status %d", resp.StatusCode)
	}

	var config WellKnownConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return fmt.Errorf("failed to decode OIDC configuration: %w", err)
	}

	p.wellKnown = &config
	log.Info().Str("issuer", config.Issuer).Msg("OIDC configuration discovered")
	return nil
}

// GetAuthorizationURL returns the URL to redirect users for authentication
func (p *OIDCProvider) GetAuthorizationURL(state string) (string, error) {
	if p.wellKnown == nil {
		if err := p.discoverConfiguration(); err != nil {
			return "", err
		}
	}

	scopes := strings.Join(p.config.Scopes, " ")
	url := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		p.wellKnown.AuthorizationEndpoint,
		p.config.ClientID,
		p.config.RedirectURL,
		scopes,
		state,
	)
	return url, nil
}

// GenerateState generates a random state parameter for OIDC flow
func GenerateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// CreateSession creates a new session for a user
func (p *OIDCProvider) CreateSession(userID, email, name, role string) (*Session, error) {
	sessionID := generateSessionID()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Email:     email,
		Name:      name,
		Role:      role,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	p.sessionStore[sessionID] = session
	return session, nil
}

// GetSession retrieves a session by ID
func (p *OIDCProvider) GetSession(sessionID string) (*Session, bool) {
	session, ok := p.sessionStore[sessionID]
	if !ok {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(p.sessionStore, sessionID)
		return nil, false
	}

	return session, true
}

// DeleteSession removes a session
func (p *OIDCProvider) DeleteSession(sessionID string) {
	delete(p.sessionStore, sessionID)
}

// GenerateJWT generates a JWT token for a session
func (p *OIDCProvider) GenerateJWT(session *Session, secret string) (string, error) {
	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   session.UserID,
			ExpiresAt: jwt.NewNumericDate(session.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "goguard",
		},
		Email:  session.Email,
		Name:   session.Name,
		Role:   session.Role,
		UserID: session.UserID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString, secret string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// AuthMiddleware creates a Gin middleware for authentication
func AuthMiddleware(jwtSecret string, oidcProvider *OIDCProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/ready" {
			c.Next()
			return
		}

		// Check for Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Check for session cookie
			sessionID, err := c.Cookie("goguard_session")
			if err != nil || sessionID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				c.Abort()
				return
			}

			session, ok := oidcProvider.GetSession(sessionID)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
				c.Abort()
				return
			}

			c.Set("user_id", session.UserID)
			c.Set("email", session.Email)
			c.Set("role", session.Role)
			c.Next()
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		claims, err := ValidateJWT(parts[1], jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "no role found"})
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		// Super admin has access to everything
		if role == "super_admin" {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// AuthHandlers provides HTTP handlers for authentication
type AuthHandlers struct {
	provider  *OIDCProvider
	jwtSecret string
}

// NewAuthHandlers creates new auth handlers
func NewAuthHandlers(provider *OIDCProvider, jwtSecret string) *AuthHandlers {
	return &AuthHandlers{
		provider:  provider,
		jwtSecret: jwtSecret,
	}
}

// HandleLogin initiates OIDC login flow
func (h *AuthHandlers) HandleLogin(c *gin.Context) {
	state := GenerateState()
	c.SetCookie("oidc_state", state, 300, "/", "", false, true)

	authURL, err := h.provider.GetAuthorizationURL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate auth URL"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleCallback handles OIDC callback
func (h *AuthHandlers) HandleCallback(c *gin.Context) {
	// In a real implementation, this would:
	// 1. Validate the state parameter
	// 2. Exchange the code for tokens
	// 3. Validate the ID token
	// 4. Create or update the user
	// 5. Create a session

	// For now, return a placeholder
	c.JSON(http.StatusOK, gin.H{
		"message": "OIDC callback - implement token exchange",
	})
}

// HandleLogout handles user logout
func (h *AuthHandlers) HandleLogout(c *gin.Context) {
	sessionID, err := c.Cookie("goguard_session")
	if err == nil && sessionID != "" {
		h.provider.DeleteSession(sessionID)
	}

	c.SetCookie("goguard_session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// HandleMe returns current user info
func (h *AuthHandlers) HandleMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"email":   email,
		"role":    role,
	})
}
