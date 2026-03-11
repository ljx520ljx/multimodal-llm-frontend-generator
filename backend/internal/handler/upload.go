package handler

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// UploadHandler handles image upload requests
type UploadHandler struct {
	sessionStore service.SessionStore
	imageService service.ImageService
}

// NewUploadHandler creates a new UploadHandler
func NewUploadHandler(sessionStore service.SessionStore, imageService service.ImageService) *UploadHandler {
	return &UploadHandler{
		sessionStore: sessionStore,
		imageService: imageService,
	}
}

// Handle handles POST /api/upload
func (h *UploadHandler) Handle(c *gin.Context) {
	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		handleValidationError(c, "Failed to parse multipart form")
		return
	}

	files := form.File["images[]"]
	if len(files) == 0 {
		// Try alternative field name
		files = form.File["images"]
	}
	if len(files) == 0 {
		handleValidationError(c, "No images provided")
		return
	}

	ctx := c.Request.Context()

	// Get or create session
	sessionID := c.PostForm("session_id")
	var session *service.Session

	if sessionID != "" {
		session, err = h.sessionStore.Get(ctx, sessionID)
		if err != nil {
			// Session ID was provided but not found — return 404
			handleError(c, err)
			return
		}
	} else {
		session, err = h.sessionStore.Create(ctx)
		if err != nil {
			handleError(c, err)
			return
		}
	}

	// Process each image
	var imageInfos []ImageInfo
	for i, fileHeader := range files {
		info, err := h.processOneImage(ctx, session.ID, fileHeader, i)
		if err != nil {
			handleError(c, err)
			return
		}
		imageInfos = append(imageInfos, *info)
	}

	c.JSON(http.StatusOK, UploadResponse{
		SessionID: session.ID,
		Images:    imageInfos,
	})
}

// processOneImage opens, processes, and stores a single uploaded image.
// The file is closed before this function returns, avoiding defer-in-loop leaks.
func (h *UploadHandler) processOneImage(ctx context.Context, sessionID string, fileHeader *multipart.FileHeader, order int) (*ImageInfo, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	imageData, err := h.imageService.Process(ctx, file, fileHeader)
	if err != nil {
		return nil, err
	}

	imageData.Order = order

	if err := h.sessionStore.AddImage(ctx, sessionID, imageData); err != nil {
		return nil, err
	}

	return &ImageInfo{
		ID:       imageData.ID,
		Filename: imageData.Filename,
		Order:    imageData.Order,
	}, nil
}
