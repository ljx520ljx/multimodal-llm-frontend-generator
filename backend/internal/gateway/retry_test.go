package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"multimodal-llm-frontend-generator/internal/gateway/types"
)

func TestRetryer_CalculateBackoff(t *testing.T) {
	config := RetryConfig{
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
		Jitter:         0, // Disable jitter for predictable tests
	}
	retryer := NewRetryer(config)

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
		{5, 30 * time.Second}, // Capped at max
		{6, 30 * time.Second}, // Stays at max
	}

	for _, tt := range tests {
		backoff := retryer.calculateBackoff(tt.attempt)
		if backoff != tt.expected {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, backoff)
		}
	}
}

func TestRetryer_Do_Success(t *testing.T) {
	config := DefaultRetryConfig()
	retryer := NewRetryer(config)

	callCount := 0
	err := retryer.Do(context.Background(), func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestRetryer_Do_NonRetryableError(t *testing.T) {
	config := DefaultRetryConfig()
	retryer := NewRetryer(config)

	nonRetryableErr := types.NewGatewayError("test", 400, "bad request", false, nil)

	callCount := 0
	err := retryer.Do(context.Background(), func() error {
		callCount++
		return nonRetryableErr
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (no retry), got %d", callCount)
	}
}

func TestRetryer_Do_RetryableError(t *testing.T) {
	config := RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0,
	}
	retryer := NewRetryer(config)

	retryableErr := types.NewGatewayError("test", 429, "rate limited", true, nil)

	callCount := 0
	err := retryer.Do(context.Background(), func() error {
		callCount++
		return retryableErr
	})

	if err == nil {
		t.Error("expected error after retries exhausted")
	}
	if callCount != 3 { // 1 initial + 2 retries
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryer_Do_EventualSuccess(t *testing.T) {
	config := RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         0,
	}
	retryer := NewRetryer(config)

	retryableErr := types.NewGatewayError("test", 500, "server error", true, nil)

	callCount := 0
	err := retryer.Do(context.Background(), func() error {
		callCount++
		if callCount < 3 {
			return retryableErr
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected success after retries, got %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryer_Do_ContextCanceled(t *testing.T) {
	config := RetryConfig{
		MaxRetries:     10,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
		Jitter:         0,
	}
	retryer := NewRetryer(config)

	ctx, cancel := context.WithCancel(context.Background())
	retryableErr := types.NewGatewayError("test", 500, "server error", true, nil)

	callCount := 0
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := retryer.Do(ctx, func() error {
		callCount++
		return retryableErr
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestIsRetryableStatusCode(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{504, true},
	}

	for _, tt := range tests {
		result := types.IsRetryableStatusCode(tt.code)
		if result != tt.expected {
			t.Errorf("status %d: expected %v, got %v", tt.code, tt.expected, result)
		}
	}
}
