package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"multimodal-llm-frontend-generator/internal/gateway/types"
	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockGateway implements gateway.LLMGateway for testing
type mockGateway struct {
	chunks []types.StreamChunk
}

func (m *mockGateway) ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan types.StreamChunk, error) {
	ch := make(chan types.StreamChunk)
	go func() {
		defer close(ch)
		for _, chunk := range m.chunks {
			ch <- chunk
		}
	}()
	return ch, nil
}

func (m *mockGateway) Provider() string {
	return "mock"
}

func createTestPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestUploadHandler_SingleImage(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	imageService := service.NewImageService(service.ImageServiceConfig{
		MaxSize:      10 * 1024 * 1024,
		Quality:      80,
		MaxDimension: 2048,
		AllowedTypes: []string{"image/png", "image/jpeg"},
	})

	handler := NewUploadHandler(store, imageService)

	// Create multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	pngData := createTestPNG()
	fw, _ := w.CreateFormFile("images[]", "test.png")
	fw.Write(pngData)
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/upload", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp UploadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.SessionID == "" {
		t.Error("expected session_id to be set")
	}
	if len(resp.Images) != 1 {
		t.Errorf("expected 1 image, got %d", len(resp.Images))
	}
}

func TestUploadHandler_MultipleImages(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	imageService := service.NewImageService(service.ImageServiceConfig{
		MaxSize:      10 * 1024 * 1024,
		Quality:      80,
		MaxDimension: 2048,
		AllowedTypes: []string{"image/png", "image/jpeg"},
	})

	handler := NewUploadHandler(store, imageService)

	// Create multipart form with 3 images
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	pngData := createTestPNG()
	for i := 0; i < 3; i++ {
		fw, _ := w.CreateFormFile("images[]", "test.png")
		fw.Write(pngData)
	}
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/upload", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp UploadResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Images) != 3 {
		t.Errorf("expected 3 images, got %d", len(resp.Images))
	}

	// Check order
	for i, img := range resp.Images {
		if img.Order != i {
			t.Errorf("expected order %d, got %d", i, img.Order)
		}
	}
}

func TestUploadHandler_NoImages(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	imageService := service.NewImageService(service.ImageServiceConfig{
		AllowedTypes: []string{"image/png"},
	})

	handler := NewUploadHandler(store, imageService)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/upload", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestGenerateHandler_Success(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := service.NewPromptService()

	mockGw := &mockGateway{
		chunks: []types.StreamChunk{
			{Type: types.ChunkTypeContent, Content: "<thinking>Analyzing...</thinking>\n\n"},
			{Type: types.ChunkTypeContent, Content: "```jsx\nexport default function App() { return <div>Hello</div> }\n```"},
			{Type: types.ChunkTypeDone},
		},
	}

	generateService := service.NewGenerateService(store, promptService, mockGw)
	handler := NewGenerateHandler(generateService)

	// Create session and add image
	ctx := context.Background()
	session, _ := store.Create(ctx)
	store.AddImage(ctx, session.ID, &service.ImageData{
		ID:       "img-1",
		Filename: "test.png",
		Base64:   "data:image/png;base64,test",
	})

	reqBody := GenerateRequest{
		SessionID: session.ID,
		ImageIDs:  []string{"img-1"},
		Framework: "react",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/generate", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Check SSE response
	respBody := rec.Body.String()
	if !strings.Contains(respBody, "event: message") {
		t.Error("expected SSE events")
	}
	if !strings.Contains(respBody, "thinking") {
		t.Error("expected thinking event")
	}
}

func TestGenerateHandler_SessionNotFound(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := service.NewPromptService()
	mockGw := &mockGateway{}
	generateService := service.NewGenerateService(store, promptService, mockGw)
	handler := NewGenerateHandler(generateService)

	reqBody := GenerateRequest{
		SessionID: "non-existent",
		ImageIDs:  []string{"img-1"},
		Framework: "react",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/generate", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestGenerateHandler_InvalidFramework(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := service.NewPromptService()
	mockGw := &mockGateway{}
	generateService := service.NewGenerateService(store, promptService, mockGw)
	handler := NewGenerateHandler(generateService)

	reqBody := `{"session_id": "test", "image_ids": ["img-1"], "framework": "angular"}`

	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/generate", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestChatHandler_Success(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := service.NewPromptService()

	mockGw := &mockGateway{
		chunks: []types.StreamChunk{
			{Type: types.ChunkTypeContent, Content: "```jsx\nexport default function App() { return <div className=\"text-blue-500\">Hello</div> }\n```"},
			{Type: types.ChunkTypeDone},
		},
	}

	generateService := service.NewGenerateService(store, promptService, mockGw)
	handler := NewChatHandler(generateService)

	// Create session with code
	ctx := context.Background()
	session, _ := store.Create(ctx)
	session.Code = "export default function App() { return <div>Hello</div> }"
	session.Framework = "react"
	store.Update(ctx, session)

	reqBody := ChatRequest{
		SessionID: session.ID,
		Message:   "把文字改成蓝色",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/chat", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestChatHandler_NoCodeGenerated(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := service.NewPromptService()
	mockGw := &mockGateway{}
	generateService := service.NewGenerateService(store, promptService, mockGw)
	handler := NewChatHandler(generateService)

	// Create session without code
	ctx := context.Background()
	session, _ := store.Create(ctx)

	reqBody := ChatRequest{
		SessionID: session.ID,
		Message:   "修改代码",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/chat", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code != ErrCodeNoCodeGenerated {
		t.Errorf("expected error code %s, got %s", ErrCodeNoCodeGenerated, resp.Code)
	}
}

// Helper to read SSE events
func readSSEEvents(r io.Reader) []service.SSEEvent {
	var events []service.SSEEvent
	body, _ := io.ReadAll(r)
	lines := strings.Split(string(body), "\n")

	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "data: ") {
			data := strings.TrimPrefix(lines[i], "data: ")
			var event service.SSEEvent
			if err := json.Unmarshal([]byte(data), &event); err == nil {
				events = append(events, event)
			}
		}
	}

	return events
}

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router := gin.New()
	router.GET("/health", Health)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp["status"])
	}
}

func TestHandleError_ImageNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	handleError(c, &service.ErrImageNotFound{ID: "img-123"})

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code != ErrCodeImageNotFound {
		t.Errorf("expected error code %s, got %s", ErrCodeImageNotFound, resp.Code)
	}
}

func TestHandleError_InvalidImageFormat(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	handleError(c, &service.ErrInvalidImageFormat{MimeType: "text/plain"})

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code != ErrCodeInvalidImageFormat {
		t.Errorf("expected error code %s, got %s", ErrCodeInvalidImageFormat, resp.Code)
	}
}

func TestHandleError_ImageTooLarge(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	handleError(c, &service.ErrImageTooLarge{Size: 20 * 1024 * 1024, MaxSize: 10 * 1024 * 1024})

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code != ErrCodeImageTooLarge {
		t.Errorf("expected error code %s, got %s", ErrCodeImageTooLarge, resp.Code)
	}
}

func TestHandleError_UnknownError(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	handleError(c, io.EOF) // Unknown error type

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code != ErrCodeGenerationFailed {
		t.Errorf("expected error code %s, got %s", ErrCodeGenerationFailed, resp.Code)
	}
}

func TestChatHandler_EmptyMessage(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := service.NewPromptService()
	mockGw := &mockGateway{}
	generateService := service.NewGenerateService(store, promptService, mockGw)
	handler := NewChatHandler(generateService)

	ctx := context.Background()
	session, _ := store.Create(ctx)
	session.Code = "export default function App() { return <div>Hello</div> }"
	store.Update(ctx, session)

	reqBody := ChatRequest{
		SessionID: session.ID,
		Message:   "", // Empty message
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/chat", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestChatHandler_InvalidJSON(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	promptService := service.NewPromptService()
	mockGw := &mockGateway{}
	generateService := service.NewGenerateService(store, promptService, mockGw)
	handler := NewChatHandler(generateService)

	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/chat", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUploadHandler_InvalidMultipartForm(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	imageService := service.NewImageService(service.ImageServiceConfig{
		AllowedTypes: []string{"image/png"},
	})

	handler := NewUploadHandler(store, imageService)

	// Send non-multipart request
	req := httptest.NewRequest(http.MethodPost, "/api/upload", strings.NewReader("not multipart"))
	req.Header.Set("Content-Type", "text/plain")

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/upload", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUploadHandler_WithExistingSession(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	imageService := service.NewImageService(service.ImageServiceConfig{
		MaxSize:      10 * 1024 * 1024,
		Quality:      80,
		MaxDimension: 2048,
		AllowedTypes: []string{"image/png", "image/jpeg"},
	})

	handler := NewUploadHandler(store, imageService)

	// Create existing session
	ctx := context.Background()
	session, _ := store.Create(ctx)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	pngData := createTestPNG()
	fw, _ := w.CreateFormFile("images[]", "test.png")
	fw.Write(pngData)

	// Add session_id field
	w.WriteField("session_id", session.ID)
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/upload", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp UploadResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.SessionID != session.ID {
		t.Errorf("expected session ID %s, got %s", session.ID, resp.SessionID)
	}
}

func TestUploadHandler_AlternativeFieldName(t *testing.T) {
	store := service.NewMemoryStore(30 * time.Minute)
	defer store.Close()

	imageService := service.NewImageService(service.ImageServiceConfig{
		MaxSize:      10 * 1024 * 1024,
		Quality:      80,
		MaxDimension: 2048,
		AllowedTypes: []string{"image/png", "image/jpeg"},
	})

	handler := NewUploadHandler(store, imageService)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	pngData := createTestPNG()
	// Use "images" instead of "images[]"
	fw, _ := w.CreateFormFile("images", "test.png")
	fw.Write(pngData)
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	rec := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/upload", handler.Handle)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
