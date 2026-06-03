package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// CreateProjectCommand — входные данные команды создания проекта.
type CreateProjectCommand struct {
	Name        string
	Description string
	Kind        string // "baremetal" | "vm"
}

// CreateProjectResult — результат выполнения команды.
type CreateProjectResult struct {
	ProjectID string
}

// CreateProjectHandler — обработчик команды создания проекта.
type CreateProjectHandler struct {
	projects domain.ProjectRepository
}

func NewCreateProjectHandler(projects domain.ProjectRepository) *CreateProjectHandler {
	return &CreateProjectHandler{projects: projects}
}

// Handle оркестрирует создание проекта.
func (h *CreateProjectHandler) Handle(ctx context.Context, cmd CreateProjectCommand) (CreateProjectResult, error) {
	// 1. Проверяем уникальность имени
	exists, err := h.projects.ExistsByName(ctx, cmd.Name)
	if err != nil {
		return CreateProjectResult{}, fmt.Errorf("check project name uniqueness: %w", err)
	}
	if exists {
		return CreateProjectResult{}, domain.ErrProjectAlreadyExists
	}

	// 2. Парсим тип проекта
	kind, err := domain.ParseProjectKind(cmd.Kind)
	if err != nil {
		return CreateProjectResult{}, fmt.Errorf("parse project kind: %w", err)
	}

	// 3. Создаём агрегат через фабричный метод
	projectID := domain.ProjectID(uuid.New().String())
	project, err := domain.NewProject(projectID, cmd.Name, cmd.Description, kind)
	if err != nil {
		return CreateProjectResult{}, fmt.Errorf("create project aggregate: %w", err)
	}

	// 4. Сохраняем
	if err := h.projects.Save(ctx, project); err != nil {
		return CreateProjectResult{}, fmt.Errorf("save project: %w", err)
	}

	return CreateProjectResult{ProjectID: string(project.ID())}, nil
}
