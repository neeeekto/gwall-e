package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gwall-e/services/inventory/internal/application"
	"github.com/gwall-e/services/inventory/internal/application/commands"
	"github.com/gwall-e/services/inventory/internal/application/queries"
	"github.com/gwall-e/services/inventory/internal/domain"
)

// ProjectsHandler — HTTP-хендлер для операций с проектами.
type ProjectsHandler struct {
	commands *application.CommandDispatcher
	queries  *application.QueryDispatcher
}

// NewProjectsHandler создаёт хендлер проектов.
func NewProjectsHandler(commands *application.CommandDispatcher, queries *application.QueryDispatcher) *ProjectsHandler {
	return &ProjectsHandler{commands: commands, queries: queries}
}

// CreateProject обрабатывает POST /api/v1/projects.
func (h *ProjectsHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "")
		return
	}

	result, err := application.SendCommand[commands.CreateProjectCommand, commands.CreateProjectResult](
		r.Context(), h.commands,
		commands.CreateProjectCommand{
			Name:        req.Name,
			Description: req.Description,
			Kind:        req.Kind,
		},
	)
	if err != nil {
		writeProjectError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, CreateProjectResponse{ProjectID: result.ProjectID})
}

// GetProject обрабатывает GET /api/v1/projects/{projectID}.
func (h *ProjectsHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "project_id is required", "")
		return
	}

	view, err := application.SendQuery[queries.GetProjectQuery, *queries.ProjectView](
		r.Context(), h.queries,
		queries.GetProjectQuery{ProjectID: projectID},
	)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "project not found", "PROJECT_NOT_FOUND")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get project", "")
		return
	}

	writeJSON(w, http.StatusOK, view)
}

// ListProjects обрабатывает GET /api/v1/projects.
func (h *ProjectsHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	result, err := application.SendQuery[queries.ListProjectsQuery, queries.ListProjectsResult](
		r.Context(), h.queries,
		queries.ListProjectsQuery{
			Kind:   q.Get("kind"),
			Status: q.Get("status"),
		},
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list projects", "")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// writeProjectError маппит доменные ошибки проектов в HTTP-статусы.
func writeProjectError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		writeError(w, http.StatusNotFound, "project not found", "PROJECT_NOT_FOUND")
	case errors.Is(err, domain.ErrProjectAlreadyExists):
		writeError(w, http.StatusConflict, "project with this name already exists", "PROJECT_ALREADY_EXISTS")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error", "")
	}
}
