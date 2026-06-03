package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// CreateNamespaceCommand — входные данные команды создания namespace.
type CreateNamespaceCommand struct {
	ProjectID string
	Name      string
}

// CreateNamespaceResult — результат выполнения команды.
type CreateNamespaceResult struct {
	NamespaceID string
}

// CreateNamespaceHandler — обработчик команды создания namespace.
type CreateNamespaceHandler struct {
	namespaces domain.NamespaceRepository
}

func NewCreateNamespaceHandler(namespaces domain.NamespaceRepository) *CreateNamespaceHandler {
	return &CreateNamespaceHandler{namespaces: namespaces}
}

// Handle оркестрирует создание namespace.
func (h *CreateNamespaceHandler) Handle(ctx context.Context, cmd CreateNamespaceCommand) (CreateNamespaceResult, error) {
	projectID := domain.ProjectID(cmd.ProjectID)

	// 1. Проверяем уникальность имени в рамках проекта
	exists, err := h.namespaces.ExistsByName(ctx, projectID, cmd.Name)
	if err != nil {
		return CreateNamespaceResult{}, fmt.Errorf("check namespace name uniqueness: %w", err)
	}
	if exists {
		return CreateNamespaceResult{}, domain.ErrNamespaceAlreadyExists
	}

	// 2. Создаём агрегат через фабричный метод
	nsID := domain.NamespaceID(uuid.New().String())
	ns, err := domain.NewNamespace(nsID, projectID, cmd.Name)
	if err != nil {
		return CreateNamespaceResult{}, fmt.Errorf("create namespace aggregate: %w", err)
	}

	// 3. Сохраняем через репозиторий
	if err := h.namespaces.Save(ctx, ns); err != nil {
		return CreateNamespaceResult{}, fmt.Errorf("save namespace: %w", err)
	}

	return CreateNamespaceResult{NamespaceID: string(ns.ID())}, nil
}
