package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/auth"
	"task-management-backend/internal/domain"
	"task-management-backend/internal/dto"
	"task-management-backend/internal/repository"
	"task-management-backend/internal/service"
)

type ProjectHandler struct {
	projectService service.ProjectService
}

func NewProjectHandler(projectService service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleProjectServiceError(c, service.ErrUnauthorized)
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), userID, req.Name, req.Description)
	if err != nil {
		handleProjectServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusCreated, "Project created successfully", toProjectResponse(project))
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleProjectServiceError(c, service.ErrUnauthorized)
		return
	}

	projects, err := h.projectService.ListProjectsByUserID(c.Request.Context(), userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list projects", err)
		return
	}

	writeSuccess(c, http.StatusOK, "Projects retrieved successfully", toProjectResponses(projects))
}

func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleProjectServiceError(c, service.ErrUnauthorized)
		return
	}

	id, ok := parseInt64Param(c, "id", "Invalid project ID")
	if !ok {
		return
	}

	project, err := h.projectService.GetProjectByID(c.Request.Context(), userID, id)
	if err != nil {
		handleProjectServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Project retrieved successfully", toProjectResponse(project))
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleProjectServiceError(c, service.ErrUnauthorized)
		return
	}

	id, ok := parseInt64Param(c, "id", "Invalid project ID")
	if !ok {
		return
	}

	var req dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	project, err := h.projectService.UpdateProject(c.Request.Context(), userID, id, req.Name, req.Description)
	if err != nil {
		handleProjectServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Project updated successfully", toProjectResponse(project))
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleProjectServiceError(c, service.ErrUnauthorized)
		return
	}

	id, ok := parseInt64Param(c, "id", "Invalid project ID")
	if !ok {
		return
	}

	if err := h.projectService.DeleteProject(c.Request.Context(), userID, id); err != nil {
		handleProjectServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Project deleted successfully", nil)
}

func (h *ProjectHandler) AddProjectMember(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleProjectServiceError(c, service.ErrUnauthorized)
		return
	}

	projectID, ok := parseInt64Param(c, "id", "Invalid project ID")
	if !ok {
		return
	}

	var req dto.AddProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	member, err := h.projectService.AddProjectMember(c.Request.Context(), userID, projectID, req.Email)
	if err != nil {
		handleProjectServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusCreated, "Project member added successfully", member)
}

func (h *ProjectHandler) ListProjectMembers(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c.Request.Context())
	if !ok {
		handleProjectServiceError(c, service.ErrUnauthorized)
		return
	}

	projectID, ok := parseInt64Param(c, "id", "Invalid project ID")
	if !ok {
		return
	}

	members, err := h.projectService.ListProjectMembers(c.Request.Context(), userID, projectID)
	if err != nil {
		handleProjectServiceError(c, err)
		return
	}

	writeSuccess(c, http.StatusOK, "Project members retrieved successfully", members)
}

func toProjectResponse(project *domain.Project) dto.ProjectResponse {
	return dto.ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func toProjectResponses(projects []*domain.Project) []dto.ProjectResponse {
	responses := make([]dto.ProjectResponse, 0, len(projects))
	for _, project := range projects {
		responses = append(responses, toProjectResponse(project))
	}

	return responses
}

func handleProjectServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUnauthorized):
		writeError(c, http.StatusUnauthorized, "Unauthorized", err)
	case errors.Is(err, service.ErrForbidden):
		writeError(c, http.StatusForbidden, "Forbidden", err)
	case errors.Is(err, service.ErrProjectNameRequired):
		writeError(c, http.StatusBadRequest, "Project name is required", err)
	case errors.Is(err, service.ErrProjectOwnerRequired):
		writeError(c, http.StatusBadRequest, "Project owner is required", err)
	case errors.Is(err, service.ErrProjectMemberEmailRequired):
		writeError(c, http.StatusBadRequest, "Project member email is required", err)
	case errors.Is(err, service.ErrProjectMemberExists):
		writeError(c, http.StatusConflict, "Project member already exists", err)
	case errors.Is(err, repository.ErrProjectNotFound):
		writeError(c, http.StatusNotFound, "Project not found", err)
	case errors.Is(err, repository.ErrUserNotFound):
		writeError(c, http.StatusNotFound, "User not found", err)
	default:
		writeError(c, http.StatusInternalServerError, "Internal server error", err)
	}
}
