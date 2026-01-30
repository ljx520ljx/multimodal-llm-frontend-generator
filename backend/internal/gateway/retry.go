package gateway

import (
	"context"
	"math"
	"math/rand"
	"time"

	"multimodal-llm-frontend-generator/internal/gateway/types"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries     int           // Maximum number of retries, default 3
	InitialBackoff time.Duration // Initial backoff duration, default 1s
	MaxBackoff     time.Duration // Maximum backoff duration, default 30s
	Multiplier     float64       // Backoff multiplier, default 2.0
	Jitter         float64       // Jitter ratio (0-1), default 0.1
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
		Jitter:         0.1,
	}
}

// Retryer handles retry logic with exponential backoff
type Retryer struct {
	config RetryConfig
}

// NewRetryer creates a new Retryer with the given configuration
func NewRetryer(config RetryConfig) *Retryer {
	return &Retryer{config: config}
}

// Do executes the given function with retry logic
func (r *Retryer) Do(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !types.IsRetryableError(err) {
			return err
		}

		// Check if we've exhausted retries
		if attempt >= r.config.MaxRetries {
			break
		}

		// Calculate backoff duration
		backoff := r.calculateBackoff(attempt)

		// Wait for backoff duration or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	return lastErr
}

// calculateBackoff calculates the backoff duration for a given attempt
func (r *Retryer) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff
	backoff := float64(r.config.InitialBackoff) * math.Pow(r.config.Multiplier, float64(attempt))

	// Apply maximum cap
	if backoff > float64(r.config.MaxBackoff) {
		backoff = float64(r.config.MaxBackoff)
	}

	// Apply jitter
	if r.config.Jitter > 0 {
		jitter := backoff * r.config.Jitter
		backoff = backoff - jitter + (rand.Float64() * 2 * jitter)
	}

	return time.Duration(backoff)
}

// RetryableGateway wraps an LLMGateway with retry logic
type RetryableGateway struct {
	gateway LLMGateway
	retryer *Retryer
}

// NewRetryableGateway creates a gateway with retry capabilities
func NewRetryableGateway(gateway LLMGateway, config RetryConfig) *RetryableGateway {
	return &RetryableGateway{
		gateway: gateway,
		retryer: NewRetryer(config),
	}
}

// Provider returns the underlying provider identifier
func (g *RetryableGateway) Provider() string {
	return g.gateway.Provider()
}

// ChatStream sends a request with retry logic
// Note: Retry is only applied to the initial connection, not to stream errors
func (g *RetryableGateway) ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan types.StreamChunk, error) {
	var resultCh <-chan types.StreamChunk

	err := g.retryer.Do(ctx, func() error {
		ch, err := g.gateway.ChatStream(ctx, req)
		if err != nil {
			return err
		}
		resultCh = ch
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resultCh, nil
}
