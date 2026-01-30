package handler

import (
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
			// Create new session if not found
			session, err = h.sessionStore.Create(ctx)
			if err != nil {
				handleError(c, err)
				return
			}
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
		file, err := fileHeader.Open()
		if err != nil {
			handleValidationError(c, "Failed to open uploaded file")
			return
		}
		defer file.Close()

		// Process image
		imageData, err := h.imageService.Process(ctx, file, fileHeader)
		if err != nil {
			handleError(c, err)
			return
		}

		// Set order
		imageData.Order = i

		// Store image
		err = h.sessionStore.AddImage(ctx, session.ID, imageData)
		if err != nil {
			handleError(c, err)
			return
		}

		imageInfos = append(imageInfos, ImageInfo{
			ID:       imageData.ID,
			Filename: imageData.Filename,
			Order:    imageData.Order,
		})
	}

	c.JSON(http.StatusOK, UploadResponse{
		SessionID: session.ID,
		Images:    imageInfos,
	})
}
