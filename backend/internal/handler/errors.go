package handler

import (
	"net/http"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// handleError handles errors and sends appropriate HTTP responses
func handleError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *service.ErrSessionNotFound:
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    ErrCodeSessionNotFound,
			Message: e.Error(),
		})
	case *service.ErrImageNotFound:
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    ErrCodeImageNotFound,
			Message: e.Error(),
		})
	case *service.ErrInvalidImageFormat:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    ErrCodeInvalidImageFormat,
			Message: e.Error(),
		})
	case *service.ErrImageTooLarge:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    ErrCodeImageTooLarge,
			Message: e.Error(),
		})
	case *service.ErrNoCodeGenerated:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    ErrCodeNoCodeGenerated,
			Message: e.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    ErrCodeGenerationFailed,
			Message: "Internal server error",
		})
	}
}

// handleValidationError handles validation errors
func handleValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    ErrCodeInvalidRequest,
		Message: message,
	})
}
