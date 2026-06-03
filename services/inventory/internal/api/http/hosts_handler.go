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

// HostsHandler — HTTP-хендлер для операций с хостами.
// Команды идут через CommandDispatcher (с транзакцией).
// Запросы идут через QueryDispatcher (без транзакции).
type HostsHandler struct {
	commands *application.CommandDispatcher
	queries  *application.QueryDispatcher
}

// NewHostsHandler создаёт хендлер хостов.
func NewHostsHandler(commands *application.CommandDispatcher, queries *application.QueryDispatcher) *HostsHandler {
	return &HostsHandler{commands: commands, queries: queries}
}

// RegisterHost обрабатывает POST /api/v1/hosts.
func (h *HostsHandler) RegisterHost(w http.ResponseWriter, r *http.Request) {
	var req RegisterHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "")
		return
	}

	result, err := application.SendCommand[commands.RegisterHostCommand, commands.RegisterHostResult](
		r.Context(), h.commands,
		commands.RegisterHostCommand{
			ProjectID: req.ProjectID,
			FQDN:      req.FQDN,
			Inv:       req.Inv,
			Kind:      req.Kind,
			Location: commands.RegisterHostLocationDTO{
				Country:  req.Location.Country,
				City:     req.Location.City,
				Building: req.Location.Building,
				Module:   req.Location.Module,
				Rack:     req.Location.Rack,
				Unit:     req.Location.Unit,
				Object:   req.Location.Object,
				RoomType: req.Location.RoomType,
			},
			Hardware: commands.RegisterHostHardwareDTO{
				Name:        req.Hardware.Name,
				Platform:    req.Hardware.Platform,
				IPMIMac:     req.Hardware.IPMIMac,
				Motherboard: req.Hardware.Motherboard,
				MACs:        req.Hardware.MACs,
			},
		},
	)
	if err != nil {
		writeHostError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, RegisterHostResponse{HostID: result.HostID})
}

// ListHosts обрабатывает GET /api/v1/hosts.
func (h *HostsHandler) ListHosts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	result, err := application.SendQuery[queries.ListHostsQuery, queries.ListHostsResult](
		r.Context(), h.queries,
		queries.ListHostsQuery{
			ProjectID: q.Get("project_id"),
			Kind:      q.Get("kind"),
		},
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list hosts", "")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetHost обрабатывает GET /api/v1/hosts/{hostID}.
func (h *HostsHandler) GetHost(w http.ResponseWriter, r *http.Request) {
	hostID := chi.URLParam(r, "hostID")
	if hostID == "" {
		writeError(w, http.StatusBadRequest, "host_id is required", "")
		return
	}

	view, err := application.SendQuery[queries.GetHostQuery, *queries.HostView](
		r.Context(), h.queries,
		queries.GetHostQuery{HostID: hostID},
	)
	if err != nil {
		if errors.Is(err, domain.ErrHostNotFound) {
			writeError(w, http.StatusNotFound, "host not found", "HOST_NOT_FOUND")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get host", "")
		return
	}

	writeJSON(w, http.StatusOK, view)
}

// writeHostError маппит доменные ошибки хостов в HTTP-статусы.
func writeHostError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrHostNotFound):
		writeError(w, http.StatusNotFound, "host not found", "HOST_NOT_FOUND")
	case errors.Is(err, domain.ErrHostAlreadyExists):
		writeError(w, http.StatusConflict, "host with this FQDN already exists", "HOST_ALREADY_EXISTS")
	case errors.Is(err, domain.ErrInvalidProjectForHost):
		writeError(w, http.StatusUnprocessableEntity, "project does not accept bare-metal hosts", "INVALID_PROJECT_KIND")
	case errors.Is(err, domain.ErrProjectNotFound):
		writeError(w, http.StatusNotFound, "project not found", "PROJECT_NOT_FOUND")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error", "")
	}
}
