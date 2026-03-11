package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newTestRateLimiter(ipRate float64, ipBurst, maxConcurrent int) *RateLimiter {
	return NewRateLimiter(RateLimitConfig{
		IPRate:        ipRate,
		IPBurst:       ipBurst,
		IPCleanupTTL:  time.Minute,
		MaxConcurrent: maxConcurrent,
	})
}

func TestRateLimiter_AllowsNormalRequests(t *testing.T) {
	rl := newTestRateLimiter(10, 10, 10)
	defer rl.Close()

	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRateLimiter_BlocksExcessiveRequests(t *testing.T) {
	// Allow only 1 request with burst of 1
	rl := newTestRateLimiter(1, 1, 100)
	defer rl.Close()

	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// First request should succeed
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("first request: expected status 200, got %d", w.Code)
	}

	// Second request should be rate limited
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected status 429, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "RATE_LIMITED" {
		t.Errorf("expected error code 'RATE_LIMITED', got '%s'", resp["code"])
	}
}

func TestRateLimiter_ConcurrencyLimit(t *testing.T) {
	// Allow high rate but only 1 concurrent request
	rl := newTestRateLimiter(1000, 1000, 1)
	defer rl.Close()

	router := gin.New()
	router.Use(rl.Middleware())

	blocked := make(chan struct{})
	proceed := make(chan struct{})

	router.GET("/slow", func(c *gin.Context) {
		close(blocked) // Signal that we're inside the handler
		<-proceed      // Wait until told to proceed
		c.String(200, "ok")
	})

	// Start first request that will block inside the handler
	go func() {
		req := httptest.NewRequest("GET", "/slow", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}()

	// Wait for the first request to enter the handler
	<-blocked

	// Second request should get 503 because concurrency is full
	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "SERVICE_BUSY" {
		t.Errorf("expected error code 'SERVICE_BUSY', got '%s'", resp["code"])
	}

	// Unblock the first request
	close(proceed)
}

func TestRateLimiter_DifferentIPsHaveSeparateLimits(t *testing.T) {
	// Allow 1 request per IP with burst of 1
	rl := newTestRateLimiter(1, 1, 100)
	defer rl.Close()

	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// First IP, first request - should succeed
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("IP1 first request: expected status 200, got %d", w.Code)
	}

	// Second IP, first request - should also succeed (different IP)
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "5.6.7.8:1234"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("IP2 first request: expected status 200, got %d", w.Code)
	}

	// First IP, second request - should be rate limited
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 second request: expected status 429, got %d", w.Code)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{
		IPRate:        10,
		IPBurst:       10,
		IPCleanupTTL:  10 * time.Millisecond,
		MaxConcurrent: 10,
	})
	defer rl.Close()

	// Add an entry
	rl.getLimiter("1.2.3.4")

	// Verify entry exists
	if _, ok := rl.ips.Load("1.2.3.4"); !ok {
		t.Fatal("expected IP entry to exist")
	}

	// Wait for TTL to expire
	time.Sleep(20 * time.Millisecond)

	// Trigger cleanup manually
	rl.cleanup()

	// Verify entry was removed
	if _, ok := rl.ips.Load("1.2.3.4"); ok {
		t.Error("expected IP entry to be cleaned up")
	}
}

func TestRateLimiter_Close(t *testing.T) {
	rl := newTestRateLimiter(10, 10, 10)

	// Close should not panic
	rl.Close()

	// Double close should not panic
	rl.Close()
}
