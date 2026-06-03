package queries

import "context"

// GetProjectQuery — параметры запроса одного проекта.
type GetProjectQuery struct {
	ProjectID string
}

// ListProjectsQuery — параметры запроса списка проектов.
type ListProjectsQuery struct {
	Kind   string // "baremetal" | "vm"
	Status string // "active" | "archived"
	Page   int
	Limit  int
}

// ProjectView — Read Model проекта.
type ProjectView struct {
	ID          string
	Name        string
	Description string
	Kind        string
	Status      string
	Tags        []string
	CreatedAt   string
	UpdatedAt   string
}

// ListProjectsResult — результат запроса списка.
type ListProjectsResult struct {
	Projects   []ProjectView
	TotalCount int
	Page       int
	Limit      int
}

// ProjectReadModel — интерфейс чтения проектов.
type ProjectReadModel interface {
	GetProjectByID(ctx context.Context, id string) (*ProjectView, error)
	ListProjects(ctx context.Context, query ListProjectsQuery) (ListProjectsResult, error)
}

// GetProjectHandler — query handler для получения одного проекта.
type GetProjectHandler struct {
	readModel ProjectReadModel
}

func NewGetProjectHandler(readModel ProjectReadModel) *GetProjectHandler {
	return &GetProjectHandler{readModel: readModel}
}

func (h *GetProjectHandler) Handle(ctx context.Context, query GetProjectQuery) (*ProjectView, error) {
	return h.readModel.GetProjectByID(ctx, query.ProjectID)
}

// ListProjectsHandler — query handler для получения списка проектов.
type ListProjectsHandler struct {
	readModel ProjectReadModel
}

func NewListProjectsHandler(readModel ProjectReadModel) *ListProjectsHandler {
	return &ListProjectsHandler{readModel: readModel}
}

func (h *ListProjectsHandler) Handle(ctx context.Context, query ListProjectsQuery) (ListProjectsResult, error) {
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	return h.readModel.ListProjects(ctx, query)
}
