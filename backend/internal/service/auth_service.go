package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	GitHubLogin string    `json:"github_login,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuthService handles user authentication and JWT tokens
type AuthService struct {
	pool               *pgxpool.Pool
	jwtSecret          []byte
	jwtExpiry          time.Duration
	githubClientID     string
	githubClientSecret string
	baseURL            string
	httpClient         *http.Client
}

// NewAuthService creates a new AuthService
func NewAuthService(pool *pgxpool.Pool, jwtSecret string, jwtExpiry time.Duration, githubClientID, githubClientSecret, baseURL string) *AuthService {
	return &AuthService{
		pool:               pool,
		jwtSecret:          []byte(jwtSecret),
		jwtExpiry:          jwtExpiry,
		githubClientID:     githubClientID,
		githubClientSecret: githubClientSecret,
		baseURL:            baseURL,
		httpClient:         &http.Client{Timeout: 15 * time.Second},
	}
}

// TokenPair contains access and refresh token information
type TokenPair struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

// JWTClaims defines the JWT payload
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Register creates a new user with email and password
func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (*User, *TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	var user User
	err = s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, display_name)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, display_name, COALESCE(avatar_url, '') as avatar_url, created_at`,
		email, string(hash), displayName,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("create user: %w", err)
	}

	token, err := s.generateToken(&user)
	if err != nil {
		return nil, nil, err
	}

	return &user, token, nil
}

// Login authenticates a user with email and password
func (s *AuthService) Login(ctx context.Context, email, password string) (*User, *TokenPair, error) {
	var user User
	var passwordHash string
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, display_name, COALESCE(avatar_url, '') as avatar_url, COALESCE(password_hash, '') as password_hash, created_at
		 FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &passwordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("query user: %w", err)
	}

	if passwordHash == "" {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(&user)
	if err != nil {
		return nil, nil, err
	}

	return &user, token, nil
}

// GitHubCallback handles the GitHub OAuth callback
func (s *AuthService) GitHubCallback(ctx context.Context, code string) (*User, *TokenPair, error) {
	// Exchange code for access token
	accessToken, err := s.exchangeGitHubCode(code)
	if err != nil {
		return nil, nil, fmt.Errorf("github exchange: %w", err)
	}

	// Get GitHub user info
	ghUser, err := s.getGitHubUser(accessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("github user: %w", err)
	}

	// Upsert user
	var user User
	err = s.pool.QueryRow(ctx,
		`INSERT INTO users (email, github_id, github_login, avatar_url, display_name)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (github_id) DO UPDATE SET
		   github_login = EXCLUDED.github_login,
		   avatar_url = EXCLUDED.avatar_url,
		   updated_at = NOW()
		 RETURNING id, email, display_name, COALESCE(avatar_url, '') as avatar_url, COALESCE(github_login, '') as github_login, created_at`,
		ghUser.Email, ghUser.ID, ghUser.Login, ghUser.AvatarURL, ghUser.Name,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.GitHubLogin, &user.CreatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("upsert github user: %w", err)
	}

	token, err := s.generateToken(&user)
	if err != nil {
		return nil, nil, err
	}

	return &user, token, nil
}

// RefreshToken generates a new token pair from a valid existing token
func (s *AuthService) RefreshToken(ctx context.Context, tokenString string) (*User, *TokenPair, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}

	var user User
	err = s.pool.QueryRow(ctx,
		`SELECT id, email, display_name, COALESCE(avatar_url, '') as avatar_url, created_at
		 FROM users WHERE id = $1`, claims.UserID,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrInvalidToken
		}
		return nil, nil, fmt.Errorf("query user: %w", err)
	}

	token, err := s.generateToken(&user)
	if err != nil {
		return nil, nil, err
	}

	return &user, token, nil
}

// ValidateToken parses and validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetGitHubAuthURL returns the GitHub OAuth authorization URL
func (s *AuthService) GetGitHubAuthURL(state string) string {
	return fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s/api/auth/github/callback&scope=user:email&state=%s",
		s.githubClientID, s.baseURL, state,
	)
}

func (s *AuthService) generateToken(user *User) (*TokenPair, error) {
	expiresAt := time.Now().Add(s.jwtExpiry)
	claims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	return &TokenPair{
		AccessToken: tokenString,
		ExpiresAt:   expiresAt.Unix(),
	}, nil
}

// GitHub OAuth types
type githubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func (s *AuthService) exchangeGitHubCode(code string) (string, error) {
	bodyMap := map[string]string{
		"client_id":     s.githubClientID,
		"client_secret": s.githubClientSecret,
		"code":          code,
	}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return "", fmt.Errorf("marshal github request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("github oauth: %s", result.Error)
	}

	return result.AccessToken, nil
}

func (s *AuthService) getGitHubUser(accessToken string) (*githubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user githubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// If email is not public, fetch from /user/emails
	if user.Email == "" {
		user.Email, _ = s.getGitHubPrimaryEmail(accessToken)
	}

	if user.Name == "" {
		user.Name = user.Login
	}

	return &user, nil
}

func (s *AuthService) getGitHubPrimaryEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", fmt.Errorf("no email found")
}

// Auth error types
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrEmailAlreadyExists = errors.New("email already registered")
)
