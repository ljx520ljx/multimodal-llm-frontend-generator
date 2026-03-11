package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// newTestAuthService creates an AuthService for testing.
// Uses a fixed JWT secret so we can generate and validate tokens.
func newTestAuthService() *service.AuthService {
	// AuthService with nil pool — only ValidateToken/GenerateToken are used, which don't touch DB
	return service.NewAuthService(
		(*pgxpool.Pool)(nil),
		"test-jwt-secret-for-unit-tests",
		1*time.Hour,
		"", "", "http://localhost:8080",
	)
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	authSvc := newTestAuthService()

	router := gin.New()
	router.Use(JWTAuth(authSvc))
	router.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	authSvc := newTestAuthService()

	router := gin.New()
	router.Use(JWTAuth(authSvc))
	router.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	authSvc := newTestAuthService()

	router := gin.New()
	router.Use(JWTAuth(authSvc))
	router.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestOptionalJWTAuth_NoHeader(t *testing.T) {
	authSvc := newTestAuthService()

	router := gin.New()
	router.Use(OptionalJWTAuth(authSvc))
	router.GET("/optional", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			c.String(200, "user:"+userID.(string))
		} else {
			c.String(200, "anonymous")
		}
	})

	req := httptest.NewRequest("GET", "/optional", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "anonymous" {
		t.Errorf("expected 'anonymous', got %q", w.Body.String())
	}
}

func TestOptionalJWTAuth_InvalidToken_StillPasses(t *testing.T) {
	authSvc := newTestAuthService()

	router := gin.New()
	router.Use(OptionalJWTAuth(authSvc))
	router.GET("/optional", func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if exists {
			c.String(200, "authenticated")
		} else {
			c.String(200, "anonymous")
		}
	})

	req := httptest.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "anonymous" {
		t.Errorf("expected 'anonymous', got %q", w.Body.String())
	}
}

func TestOptionalJWTAuth_NonBearerFormat_StillPasses(t *testing.T) {
	authSvc := newTestAuthService()

	router := gin.New()
	router.Use(OptionalJWTAuth(authSvc))
	router.GET("/optional", func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if exists {
			c.String(200, "authenticated")
		} else {
			c.String(200, "anonymous")
		}
	})

	req := httptest.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "anonymous" {
		t.Errorf("expected 'anonymous', got %q", w.Body.String())
	}
}
