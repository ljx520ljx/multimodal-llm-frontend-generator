package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AgentClient interface for communicating with Python Agent service
type AgentClient interface {
	// Echo sends an echo request and returns SSE events (for testing)
	Echo(ctx context.Context, req *EchoRequest) (<-chan SSEEvent, error)

	// Generate sends a generate request and returns SSE events
	Generate(ctx context.Context, req *AgentGenerateRequest) (<-chan SSEEvent, error)

	// Health checks if the agent service is healthy
	Health(ctx context.Context) error
}

// EchoRequest represents a request to the echo endpoint
type EchoRequest struct {
	Message string  `json:"message"`
	Count   int     `json:"count"`
	Delay   float64 `json:"delay,omitempty"`
}

// AgentGenerateRequest represents a request to the generate endpoint
type AgentGenerateRequest struct {
	SessionID string                   `json:"session_id"`
	Images    []map[string]interface{} `json:"images"`
	Options   map[string]interface{}   `json:"options,omitempty"`
}

// agentClient implements AgentClient
type agentClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAgentClient creates a new AgentClient
func NewAgentClient(baseURL string, timeout time.Duration) AgentClient {
	// For SSE streaming, we don't set a global timeout on the HTTP client
	// because it would cut off long-running streams.
	// Instead, we rely on context cancellation for request lifecycle management.
	// The timeout parameter is used for connection establishment only.
	transport := &http.Transport{
		ResponseHeaderTimeout: timeout, // Timeout for receiving response headers
	}

	return &agentClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Transport: transport,
			// No global Timeout - SSE streams can run indefinitely
			// Cancellation is handled via context
		},
	}
}

// Health checks if the agent service is healthy
func (c *agentClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to agent service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// Echo sends an echo request and returns SSE events
func (c *agentClient) Echo(ctx context.Context, req *EchoRequest) (<-chan SSEEvent, error) {
	// Default values
	if req.Count == 0 {
		req.Count = 5
	}
	if req.Delay == 0 {
		req.Delay = 0.5
	}

	// Create request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/echo", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent service: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("agent service error: status %d", resp.StatusCode)
	}

	// Create output channel
	outChan := make(chan SSEEvent)

	// Process SSE stream in background
	go func() {
		defer close(outChan)
		defer resp.Body.Close()
		c.processSSEStream(ctx, resp.Body, outChan)
	}()

	return outChan, nil
}

// processSSEStream parses SSE events from the response body
func (c *agentClient) processSSEStream(ctx context.Context, body io.Reader, outChan chan<- SSEEvent) {
	scanner := bufio.NewScanner(body)

	var eventType string
	var dataBuffer strings.Builder

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			outChan <- SSEEvent{Type: SSETypeError, Content: "request cancelled"}
			return
		default:
		}

		line := scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			if dataBuffer.Len() > 0 {
				event := c.parseSSEData(eventType, dataBuffer.String())
				if event.Type != "" {
					outChan <- event
				}

				// Check for done event
				if eventType == "done" {
					return
				}
			}
			eventType = ""
			dataBuffer.Reset()
			continue
		}

		// Parse SSE fields
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			dataBuffer.WriteString(data)
		}
	}

	if err := scanner.Err(); err != nil {
		outChan <- SSEEvent{Type: SSETypeError, Content: fmt.Sprintf("stream error: %v", err)}
	}
}

// parseSSEData parses the data field of an SSE event
func (c *agentClient) parseSSEData(eventType, data string) SSEEvent {
	// Map event types to our SSE types
	switch eventType {
	case "message":
		return SSEEvent{Type: SSETypeThinking, Content: data}
	case "done":
		return SSEEvent{Type: SSETypeDone, Content: ""}
	case "error":
		return SSEEvent{Type: SSETypeError, Content: data}
	case "thinking":
		return SSEEvent{Type: SSETypeThinking, Content: data}
	case "code":
		return SSEEvent{Type: SSETypeCode, Content: data}
	case "agent_start":
		return SSEEvent{Type: SSETypeThinking, Content: data}
	case "agent_result":
		return SSEEvent{Type: SSETypeThinking, Content: data}
	default:
		// For unknown event types, treat as thinking
		if eventType != "" {
			return SSEEvent{Type: SSETypeThinking, Content: data}
		}
		return SSEEvent{}
	}
}

// Generate sends a generate request and returns SSE events
func (c *agentClient) Generate(ctx context.Context, req *AgentGenerateRequest) (<-chan SSEEvent, error) {
	// Create request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent service: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("agent service error: status %d", resp.StatusCode)
	}

	// Create output channel
	outChan := make(chan SSEEvent)

	// Process SSE stream in background
	go func() {
		defer close(outChan)
		defer resp.Body.Close()
		c.processSSEStream(ctx, resp.Body, outChan)
	}()

	return outChan, nil
}
