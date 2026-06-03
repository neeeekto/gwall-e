package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// ProvisionVMCommand — входные данные команды создания VM.
type ProvisionVMCommand struct {
	ProjectID   string
	NamespaceID string
	Name        string
	Spec        ProvisionVMSpecDTO
}

type ProvisionVMSpecDTO struct {
	CPUCores uint32
	RAMmb    uint32
	DiskGB   uint32
	DiskType string // "ssd" | "hdd" | "nvme"
}

// ProvisionVMResult — результат выполнения команды.
type ProvisionVMResult struct {
	VMID string
}

// ProvisionVMHandler — обработчик команды создания VM.
// Использует ProjectRepository напрямую вместо ProjectChecker порта.
type ProvisionVMHandler struct {
	vms      domain.VMRepository
	projects domain.ProjectRepository
}

func NewProvisionVMHandler(
	vms domain.VMRepository,
	projects domain.ProjectRepository,
) *ProvisionVMHandler {
	return &ProvisionVMHandler{
		vms:      vms,
		projects: projects,
	}
}

// Handle оркестрирует создание виртуальной машины.
func (h *ProvisionVMHandler) Handle(ctx context.Context, cmd ProvisionVMCommand) (ProvisionVMResult, error) {
	projectID := domain.ProjectID(cmd.ProjectID)

	// 1. Загружаем проект и проверяем его тип
	project, err := h.projects.FindByID(ctx, projectID)
	if err != nil {
		return ProvisionVMResult{}, fmt.Errorf("find project: %w", err)
	}

	// 2. Доменное правило: VM только в VM-проекте
	if !project.IsVM() {
		return ProvisionVMResult{}, domain.ErrInvalidProjectForVM
	}

	// 3. Проверяем уникальность имени в рамках проекта
	exists, err := h.vms.ExistsByName(ctx, projectID, cmd.Name)
	if err != nil {
		return ProvisionVMResult{}, fmt.Errorf("check vm name uniqueness: %w", err)
	}
	if exists {
		return ProvisionVMResult{}, domain.ErrVMAlreadyExists
	}

	// 4. Маппинг DTO → доменные типы
	spec := domain.VMSpec{
		CPUCores: cmd.Spec.CPUCores,
		RAMmb:    cmd.Spec.RAMmb,
		DiskGB:   cmd.Spec.DiskGB,
		DiskType: cmd.Spec.DiskType,
	}

	// 5. Создаём агрегат через фабричный метод
	vmID := domain.VMID(uuid.New().String())
	vm, err := domain.NewVM(
		vmID,
		projectID,
		domain.NamespaceID(cmd.NamespaceID),
		cmd.Name,
		spec,
	)
	if err != nil {
		return ProvisionVMResult{}, fmt.Errorf("create vm aggregate: %w", err)
	}

	// 6. Сохраняем через репозиторий
	if err := h.vms.Save(ctx, vm); err != nil {
		return ProvisionVMResult{}, fmt.Errorf("save vm: %w", err)
	}

	return ProvisionVMResult{VMID: string(vm.ID())}, nil
}
