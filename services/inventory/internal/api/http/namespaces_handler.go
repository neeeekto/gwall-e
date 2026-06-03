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

// NamespacesHandler — HTTP-хендлер для операций с namespace.
type NamespacesHandler struct {
	commands *application.CommandDispatcher
	queries  *application.QueryDispatcher
}

// NewNamespacesHandler создаёт хендлер namespace.
func NewNamespacesHandler(commands *application.CommandDispatcher, queries *application.QueryDispatcher) *NamespacesHandler {
	return &NamespacesHandler{commands: commands, queries: queries}
}

// CreateNamespace обрабатывает POST /api/v1/projects/{projectID}/namespaces.
func (h *NamespacesHandler) CreateNamespace(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "project_id is required", "")
		return
	}

	var req CreateNamespaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "")
		return
	}

	result, err := application.SendCommand[commands.CreateNamespaceCommand, commands.CreateNamespaceResult](
		r.Context(), h.commands,
		commands.CreateNamespaceCommand{
			ProjectID: projectID,
			Name:      req.Name,
		},
	)
	if err != nil {
		writeNamespaceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, CreateNamespaceResponse{NamespaceID: result.NamespaceID})
}

// ListNamespaces обрабатывает GET /api/v1/projects/{projectID}/namespaces.
func (h *NamespacesHandler) ListNamespaces(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		writeError(w, http.StatusBadRequest, "project_id is required", "")
		return
	}

	result, err := application.SendQuery[queries.ListNamespacesQuery, queries.ListNamespacesResult](
		r.Context(), h.queries,
		queries.ListNamespacesQuery{ProjectID: projectID},
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list namespaces", "")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// writeNamespaceError маппит доменные ошибки namespace в HTTP-статусы.
func writeNamespaceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNamespaceNotFound):
		writeError(w, http.StatusNotFound, "namespace not found", "NAMESPACE_NOT_FOUND")
	case errors.Is(err, domain.ErrNamespaceAlreadyExists):
		writeError(w, http.StatusConflict, "namespace with this name already exists", "NAMESPACE_ALREADY_EXISTS")
	case errors.Is(err, domain.ErrProjectNotFound):
		writeError(w, http.StatusNotFound, "project not found", "PROJECT_NOT_FOUND")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error", "")
	}
}
