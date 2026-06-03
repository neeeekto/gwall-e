package queries

import "context"

// ListNamespacesQuery — параметры запроса списка namespace.
type ListNamespacesQuery struct {
	ProjectID string // обязательный фильтр
	Page      int
	Limit     int
}

// GetNamespaceQuery — параметры запроса одного namespace.
type GetNamespaceQuery struct {
	NamespaceID string
}

// NamespaceView — Read Model namespace.
type NamespaceView struct {
	ID        string
	ProjectID string
	Name      string
	CreatedAt string
	UpdatedAt string
}

// ListNamespacesResult — результат запроса списка.
type ListNamespacesResult struct {
	Namespaces []NamespaceView
	TotalCount int
	Page       int
	Limit      int
}

// NamespaceReadModel — интерфейс чтения namespace.
type NamespaceReadModel interface {
	GetNamespaceByID(ctx context.Context, id string) (*NamespaceView, error)
	ListNamespaces(ctx context.Context, query ListNamespacesQuery) (ListNamespacesResult, error)
}

// ListNamespacesHandler — query handler для получения списка namespace.
type ListNamespacesHandler struct {
	readModel NamespaceReadModel
}

func NewListNamespacesHandler(readModel NamespaceReadModel) *ListNamespacesHandler {
	return &ListNamespacesHandler{readModel: readModel}
}

func (h *ListNamespacesHandler) Handle(ctx context.Context, query ListNamespacesQuery) (ListNamespacesResult, error) {
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	return h.readModel.ListNamespaces(ctx, query)
}

// GetNamespaceHandler — query handler для получения одного namespace.
type GetNamespaceHandler struct {
	readModel NamespaceReadModel
}

func NewGetNamespaceHandler(readModel NamespaceReadModel) *GetNamespaceHandler {
	return &GetNamespaceHandler{readModel: readModel}
}

func (h *GetNamespaceHandler) Handle(ctx context.Context, query GetNamespaceQuery) (*NamespaceView, error) {
	return h.readModel.GetNamespaceByID(ctx, query.NamespaceID)
}
