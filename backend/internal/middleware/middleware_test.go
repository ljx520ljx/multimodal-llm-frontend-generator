package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORS_AllowedOrigin(t *testing.T) {
	router := gin.New()
	router.Use(CORS([]string{"http://localhost:3000", "http://example.com"}))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin 'http://localhost:3000', got '%s'",
			w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	router := gin.New()
	router.Use(CORS([]string{"http://localhost:3000"}))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://malicious.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected empty Access-Control-Allow-Origin, got '%s'",
			w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_WildcardOrigin(t *testing.T) {
	router := gin.New()
	router.Use(CORS([]string{"*"}))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://any-origin.com" {
		t.Errorf("expected Access-Control-Allow-Origin 'http://any-origin.com', got '%s'",
			w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_Preflight(t *testing.T) {
	router := gin.New()
	router.Use(CORS([]string{"http://localhost:3000"}))
	router.OPTIONS("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}

	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("expected Access-Control-Allow-Headers header")
	}
}

func TestCORS_Headers(t *testing.T) {
	router := gin.New()
	router.Use(CORS([]string{"http://localhost:3000"}))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials 'true', got '%s'",
			w.Header().Get("Access-Control-Allow-Credentials"))
	}

	if w.Header().Get("Access-Control-Max-Age") != "86400" {
		t.Errorf("expected Access-Control-Max-Age '86400', got '%s'",
			w.Header().Get("Access-Control-Max-Age"))
	}
}

func TestLogger(t *testing.T) {
	router := gin.New()
	router.Use(Logger())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestLogger_POST(t *testing.T) {
	router := gin.New()
	router.Use(Logger())
	router.POST("/api/test", func(c *gin.Context) {
		c.JSON(201, gin.H{"created": true})
	})

	req := httptest.NewRequest("POST", "/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 201 {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestRecovery_NoPanic(t *testing.T) {
	router := gin.New()
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRecovery_WithPanic(t *testing.T) {
	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("expected error response body")
	}
}

func TestRecovery_WithPanicError(t *testing.T) {
	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		var s []int
		_ = s[10] // This will cause an index out of range panic
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}
