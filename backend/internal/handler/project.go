package handler

import (
	"errors"
	"net/http"

	"multimodal-llm-frontend-generator/internal/service"

	"github.com/gin-gonic/gin"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	projectService *service.ProjectService
}

// NewProjectHandler creates a new ProjectHandler
func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

type createProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type updateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// Create handles project creation
func (h *ProjectHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "Not authenticated",
		})
		return
	}

	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrCodeInvalidRequest,
			"message": err.Error(),
		})
		return
	}

	project, err := h.projectService.Create(c.Request.Context(), userID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "PROJECT_CREATE_FAILED",
			"message": "Failed to create project",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"project": project})
}

// List handles listing user's projects
func (h *ProjectHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "Not authenticated",
		})
		return
	}

	projects, err := h.projectService.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "PROJECT_LIST_FAILED",
			"message": "Failed to list projects",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

// Get handles getting a single project
func (h *ProjectHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("id")

	project, err := h.projectService.Get(c.Request.Context(), userID, projectID)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "PROJECT_NOT_FOUND",
				"message": "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "PROJECT_GET_FAILED",
			"message": "Failed to get project",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project": project})
}

// Update handles project update
func (h *ProjectHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("id")

	var req updateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ErrCodeInvalidRequest,
			"message": err.Error(),
		})
		return
	}

	project, err := h.projectService.Update(c.Request.Context(), userID, projectID, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "PROJECT_NOT_FOUND",
				"message": "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "PROJECT_UPDATE_FAILED",
			"message": "Failed to update project",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project": project})
}

// Delete handles project deletion
func (h *ProjectHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("id")

	err := h.projectService.Delete(c.Request.Context(), userID, projectID)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "PROJECT_NOT_FOUND",
				"message": "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "PROJECT_DELETE_FAILED",
			"message": "Failed to delete project",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
}
