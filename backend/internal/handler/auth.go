package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	Token string `json:"token" binding:"required"`
}

type authResponse struct {
	User  *service.User      `json:"user"`
	Token *service.TokenPair `json:"token"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrCodeInvalidRequest,
			"message": err.Error(),
		})
		return
	}

	user, token, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"code":    "EMAIL_EXISTS",
				"message": "Email already registered",
			})
			return
		}
		// Check for PostgreSQL unique violation
		if isUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{
				"code":    "EMAIL_EXISTS",
				"message": "Email already registered",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "REGISTRATION_FAILED",
			"message": "Failed to register user",
		})
		return
	}

	c.JSON(http.StatusCreated, authResponse{User: user, Token: token})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrCodeInvalidRequest,
			"message": err.Error(),
		})
		return
	}

	user, token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_CREDENTIALS",
				"message": "Invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "LOGIN_FAILED",
			"message": "Failed to login",
		})
		return
	}

	c.JSON(http.StatusOK, authResponse{User: user, Token: token})
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrCodeInvalidRequest,
			"message": err.Error(),
		})
		return
	}

	user, token, err := h.authService.RefreshToken(c.Request.Context(), req.Token)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_TOKEN",
				"message": "Invalid or expired token",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "REFRESH_FAILED",
			"message": "Failed to refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, authResponse{User: user, Token: token})
}

// GitHubAuth redirects to GitHub OAuth with CSRF-safe state
func (h *AuthHandler) GitHubAuth(c *gin.Context) {
	redirectURL := c.Query("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}
	// Validate redirect URL is a relative path to prevent open redirect
	if !strings.HasPrefix(redirectURL, "/") || strings.HasPrefix(redirectURL, "//") {
		redirectURL = "/"
	}

	// Generate cryptographically random state for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "STATE_GENERATION_FAILED",
			"message": "Failed to generate OAuth state",
		})
		return
	}
	state := hex.EncodeToString(stateBytes)

	// Store state and redirect URL in secure HttpOnly cookies
	c.SetCookie("github_oauth_state", state, 600, "/api/auth/github/callback", "", false, true)
	c.SetCookie("github_redirect_url", redirectURL, 600, "/api/auth/github/callback", "", false, true)

	url := h.authService.GetGitHubAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubCallback handles the GitHub OAuth callback
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrCodeInvalidRequest,
			"message": "Missing code parameter",
		})
		return
	}

	// CSRF validation: compare state from query with state from cookie
	queryState := c.Query("state")
	cookieState, err := c.Cookie("github_oauth_state")
	if err != nil || cookieState == "" || queryState != cookieState {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_STATE",
			"message": "Invalid OAuth state parameter",
		})
		return
	}

	// Clear the state cookie
	c.SetCookie("github_oauth_state", "", -1, "/api/auth/github/callback", "", false, true)

	user, token, authErr := h.authService.GitHubCallback(c.Request.Context(), code)
	if authErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "GITHUB_AUTH_FAILED",
			"message": "GitHub authentication failed",
		})
		return
	}

	// Get redirect URL from cookie (validated on initial request)
	redirectURL, _ := c.Cookie("github_redirect_url")
	c.SetCookie("github_redirect_url", "", -1, "/api/auth/github/callback", "", false, true)
	if redirectURL == "" || !strings.HasPrefix(redirectURL, "/") || strings.HasPrefix(redirectURL, "//") {
		redirectURL = "/"
	}

	// Use URL fragment (#) instead of query params to avoid token in server logs/browser history
	c.Redirect(http.StatusTemporaryRedirect,
		redirectURL+"#access_token="+token.AccessToken+"&user_id="+user.ID)
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "Not authenticated",
		})
		return
	}

	// Safely extract Bearer token
	authHeader := c.GetHeader("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader || tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "Invalid authorization header",
		})
		return
	}

	// Query user from database via auth service
	user, _, err := h.authService.RefreshToken(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "USER_FETCH_FAILED",
			"message": "Failed to fetch user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// isUniqueViolation checks if a PostgreSQL error is a unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "unique constraint")
}
