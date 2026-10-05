package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/yannick2009/yily/internal/domain/model"
	"github.com/yannick2009/yily/internal/repository"
)

// ProjectHandler handles HTTP requests for project operations.
type ProjectHandler struct {
	repo repository.ProjectRepository
}

// NewProjectHandler creates a new instance of ProjectHandler.
func NewProjectHandler(repo repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: repo}
}

// ProjectPayload represents the payload for creating or updating a project.
type ProjectPayload struct {
	Name string `json:"name"`
}

// Create creates a new project
func (h *ProjectHandler) Create(c *echo.Context) error {
	var req ProjectPayload
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "project name is required"})
	}

	project := &model.Project{Name: name}
	if err := h.repo.Create(c.Request().Context(), project); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create project"})
	}

	return c.JSON(http.StatusCreated, project)

}

// GetAll retrieves all projects.
func (h *ProjectHandler) GetAll(c *echo.Context) error {
	projects, err := h.repo.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve projects"})
	}

	if projects == nil {
		projects = []model.Project{}
	}

	return c.JSON(http.StatusOK, projects)
}

// GetByID retrieves a project by its UUID.
func (h *ProjectHandler) GetByID(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid project ID"})
	}

	project, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "project not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve project"})
	}

	return c.JSON(http.StatusOK, project)
}

// SetName updates the name of an existing project.
func (h *ProjectHandler) SetName(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid project ID"})
	}

	var req ProjectPayload
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "project name is required"})
	}

	if err := h.repo.SetName(c.Request().Context(), id, name); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update project name"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "project updated"})
}

// Delete deletes a project by its UUID.
func (h *ProjectHandler) Delete(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid project ID"})
	}

	if err := h.repo.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete project"})
	}

	return c.NoContent(http.StatusNoContent)
}
