package middleware

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// IP-based token bucket
	IPRate       float64       // Requests per second per IP
	IPBurst      int           // Max burst size per IP
	IPCleanupTTL time.Duration // How long to keep idle IP entries

	// Endpoint concurrency semaphore
	MaxConcurrent int // Max concurrent requests for the endpoint
}

// ipLimiter stores a rate limiter and the last time it was used
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64 // unix nano timestamp
}

// RateLimiter manages per-IP rate limiting and endpoint concurrency
type RateLimiter struct {
	ips          sync.Map
	ipRate       rate.Limit
	ipBurst      int
	cleanupTTL   time.Duration
	semaphore    chan struct{}
	done         chan struct{}
	closeOnce    sync.Once
}

// NewRateLimiter creates a new RateLimiter with background cleanup
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		ipRate:     rate.Limit(cfg.IPRate),
		ipBurst:    cfg.IPBurst,
		cleanupTTL: cfg.IPCleanupTTL,
		semaphore:  make(chan struct{}, cfg.MaxConcurrent),
		done:       make(chan struct{}),
	}

	go rl.cleanupLoop()

	return rl
}

// cleanupLoop periodically removes expired IP entries
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.done:
			return
		}
	}
}

// cleanup removes IP entries that haven't been seen recently
func (rl *RateLimiter) cleanup() {
	now := time.Now().UnixNano()
	ttlNano := rl.cleanupTTL.Nanoseconds()
	rl.ips.Range(func(key, value any) bool {
		entry := value.(*ipLimiter)
		if now-entry.lastSeen.Load() > ttlNano {
			rl.ips.Delete(key)
		}
		return true
	})
}

// getLimiter returns the rate limiter for a given IP, creating one if needed
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	now := time.Now().UnixNano()

	if v, ok := rl.ips.Load(ip); ok {
		entry := v.(*ipLimiter)
		entry.lastSeen.Store(now)
		return entry.limiter
	}

	limiter := rate.NewLimiter(rl.ipRate, rl.ipBurst)
	entry := &ipLimiter{limiter: limiter}
	entry.lastSeen.Store(now)
	actual, _ := rl.ips.LoadOrStore(ip, entry)
	return actual.(*ipLimiter).limiter
}

// Close stops the background cleanup goroutine
func (rl *RateLimiter) Close() {
	rl.closeOnce.Do(func() {
		close(rl.done)
	})
}

// Middleware returns a Gin middleware that enforces both IP rate limiting
// and endpoint concurrency limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Layer 1: IP-based token bucket rate limiting
		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			log.Printf("[RATELIMIT] IP %s exceeded rate limit on %s", ip, c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    "RATE_LIMITED",
				"message": "Too many requests, please try again later",
			})
			return
		}

		// Layer 2: Endpoint concurrency semaphore
		select {
		case rl.semaphore <- struct{}{}:
			defer func() { <-rl.semaphore }()
		default:
			log.Printf("[RATELIMIT] Max concurrent requests reached on %s", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    "SERVICE_BUSY",
				"message": "Server is busy, please try again later",
			})
			return
		}

		c.Next()
	}
}
