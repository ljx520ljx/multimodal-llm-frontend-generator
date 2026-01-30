package types

import (
	"errors"
	"testing"
)

func TestNewContentChunk(t *testing.T) {
	chunk := NewContentChunk("hello world")

	if chunk.Type != ChunkTypeContent {
		t.Errorf("expected type %s, got %s", ChunkTypeContent, chunk.Type)
	}
	if chunk.Content != "hello world" {
		t.Errorf("expected content 'hello world', got '%s'", chunk.Content)
	}
	if chunk.Error != nil {
		t.Error("expected no error")
	}
}

func TestNewErrorChunk(t *testing.T) {
	err := errors.New("test error")
	chunk := NewErrorChunk(err)

	if chunk.Type != ChunkTypeError {
		t.Errorf("expected type %s, got %s", ChunkTypeError, chunk.Type)
	}
	if chunk.Error != err {
		t.Error("expected error to be set")
	}
}

func TestNewDoneChunk(t *testing.T) {
	chunk := NewDoneChunk()

	if chunk.Type != ChunkTypeDone {
		t.Errorf("expected type %s, got %s", ChunkTypeDone, chunk.Type)
	}
}

func TestNewTextMessage(t *testing.T) {
	msg := NewTextMessage(RoleUser, "hello")

	if msg.Role != RoleUser {
		t.Errorf("expected role %s, got %s", RoleUser, msg.Role)
	}
	if len(msg.Content) != 1 {
		t.Fatalf("expected 1 content part, got %d", len(msg.Content))
	}
	if msg.Content[0].Type != ContentTypeText {
		t.Errorf("expected content type %s, got %s", ContentTypeText, msg.Content[0].Type)
	}
	if msg.Content[0].Text != "hello" {
		t.Errorf("expected text 'hello', got '%s'", msg.Content[0].Text)
	}
}

func TestNewImageMessage(t *testing.T) {
	msg := NewImageMessage(RoleUser, "describe this image", "data:image/png;base64,abc123")

	if msg.Role != RoleUser {
		t.Errorf("expected role %s, got %s", RoleUser, msg.Role)
	}
	if len(msg.Content) != 2 {
		t.Fatalf("expected 2 content parts, got %d", len(msg.Content))
	}
	if msg.Content[0].Type != ContentTypeText {
		t.Errorf("expected first content type %s, got %s", ContentTypeText, msg.Content[0].Type)
	}
	if msg.Content[1].Type != ContentTypeImageURL {
		t.Errorf("expected second content type %s, got %s", ContentTypeImageURL, msg.Content[1].Type)
	}
	if msg.Content[1].ImageURL == nil {
		t.Fatal("expected ImageURL to be set")
	}
	if msg.Content[1].ImageURL.URL != "data:image/png;base64,abc123" {
		t.Errorf("expected image URL to be set correctly")
	}
}

func TestGatewayError_Error(t *testing.T) {
	// Test with wrapped error
	innerErr := errors.New("inner error")
	err1 := NewGatewayError("openai", 500, "test message", true, innerErr)
	errStr1 := err1.Error()
	if errStr1 != "[openai] test message: inner error" {
		t.Errorf("unexpected error string: %s", errStr1)
	}

	// Test without wrapped error
	err2 := NewGatewayError("openai", 429, "rate limited", true, nil)
	errStr2 := err2.Error()
	if errStr2 != "[openai] rate limited (status: 429)" {
		t.Errorf("unexpected error string: %s", errStr2)
	}
}

func TestGatewayError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := NewGatewayError("openai", 500, "test", true, innerErr)

	unwrapped := err.Unwrap()
	if unwrapped != innerErr {
		t.Error("expected unwrapped error to match inner error")
	}
}

func TestGatewayError_IsRetryable(t *testing.T) {
	retryable := NewGatewayError("openai", 429, "rate limited", true, nil)
	if !retryable.IsRetryable() {
		t.Error("expected error to be retryable")
	}

	notRetryable := NewGatewayError("openai", 401, "unauthorized", false, nil)
	if notRetryable.IsRetryable() {
		t.Error("expected error to not be retryable")
	}
}

func TestIsRetryableStatusCode(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		result := IsRetryableStatusCode(tc.code)
		if result != tc.expected {
			t.Errorf("IsRetryableStatusCode(%d) = %v, expected %v", tc.code, result, tc.expected)
		}
	}
}

func TestIsRetryableError(t *testing.T) {
	// GatewayError that is retryable
	retryableErr := NewGatewayError("openai", 429, "rate limited", true, nil)
	if !IsRetryableError(retryableErr) {
		t.Error("expected GatewayError with Retryable=true to be retryable")
	}

	// GatewayError that is not retryable
	notRetryableErr := NewGatewayError("openai", 401, "unauthorized", false, nil)
	if IsRetryableError(notRetryableErr) {
		t.Error("expected GatewayError with Retryable=false to not be retryable")
	}

	// Regular error (not a GatewayError)
	regularErr := errors.New("some error")
	if IsRetryableError(regularErr) {
		t.Error("expected regular error to not be retryable")
	}
}

func TestNewGatewayError(t *testing.T) {
	innerErr := errors.New("test")
	err := NewGatewayError("gemini", 503, "service unavailable", true, innerErr)

	if err.Provider != "gemini" {
		t.Errorf("expected provider 'gemini', got '%s'", err.Provider)
	}
	if err.StatusCode != 503 {
		t.Errorf("expected status code 503, got %d", err.StatusCode)
	}
	if err.Message != "service unavailable" {
		t.Errorf("expected message 'service unavailable', got '%s'", err.Message)
	}
	if !err.Retryable {
		t.Error("expected Retryable to be true")
	}
	if err.Err != innerErr {
		t.Error("expected Err to match inner error")
	}
}
