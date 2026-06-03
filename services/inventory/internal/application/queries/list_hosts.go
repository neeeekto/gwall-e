package queries

import "context"

// ListHostsQuery — параметры запроса списка хостов.
type ListHostsQuery struct {
	ProjectID string
	Kind      string // "server" | "mac"
	Tags      []string
	Page      int
	Limit     int
}

// GetHostQuery — параметры запроса одного хоста.
type GetHostQuery struct {
	HostID string
}

// HostView — Read Model: плоская структура для отображения.
type HostView struct {
	ID          string
	FQDN        string
	Inv         int
	Kind        string
	Status      string
	ProjectID   string
	ProjectName string
	Tags        []string
	Location    HostLocationView
	Hardware    HostHardwareView
	CreatedAt   string
	UpdatedAt   string
}

type HostLocationView struct {
	Country  string
	City     string
	Building string
	Rack     string
	Unit     string
}

type HostHardwareView struct {
	Name     string
	Platform string
	CPUCount int
	RAMCount int
}

type ListHostsResult struct {
	Hosts      []HostView
	TotalCount int
	Page       int
	Limit      int
}

// HostReadModel — интерфейс только для чтения.
// Определён в application-слое: пагинация, фильтры — не доменные концепции.
type HostReadModel interface {
	ListHosts(ctx context.Context, query ListHostsQuery) (ListHostsResult, error)
	GetHostByID(ctx context.Context, id string) (*HostView, error)
}

// ListHostsHandler — query handler для получения списка хостов.
type ListHostsHandler struct {
	readModel HostReadModel
}

func NewListHostsHandler(readModel HostReadModel) *ListHostsHandler {
	return &ListHostsHandler{readModel: readModel}
}

func (h *ListHostsHandler) Handle(ctx context.Context, query ListHostsQuery) (ListHostsResult, error) {
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	return h.readModel.ListHosts(ctx, query)
}

// GetHostHandler — query handler для получения одного хоста.
type GetHostHandler struct {
	readModel HostReadModel
}

func NewGetHostHandler(readModel HostReadModel) *GetHostHandler {
	return &GetHostHandler{readModel: readModel}
}

func (h *GetHostHandler) Handle(ctx context.Context, query GetHostQuery) (*HostView, error) {
	return h.readModel.GetHostByID(ctx, query.HostID)
}
