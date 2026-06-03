package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// ProvisionHostFromShadowCommand — команда добавления shadow-хоста в проект.
type ProvisionHostFromShadowCommand struct {
	ShadowHostID string
	ProjectID    string
	FQDN         string
}

// ProvisionHostFromShadowResult — результат выполнения команды.
type ProvisionHostFromShadowResult struct {
	HostID string
}

// ProvisionHostFromShadowHandler — обработчик команды добавления в проект.
// Использует ProjectRepository напрямую вместо ProjectChecker порта.
type ProvisionHostFromShadowHandler struct {
	hosts       domain.HostRepository
	shadowHosts domain.ShadowHostRepository
	projects    domain.ProjectRepository
	publisher   domain.EventPublisher
}

func NewProvisionHostFromShadowHandler(
	hosts domain.HostRepository,
	shadowHosts domain.ShadowHostRepository,
	projects domain.ProjectRepository,
	publisher domain.EventPublisher,
) *ProvisionHostFromShadowHandler {
	return &ProvisionHostFromShadowHandler{
		hosts:       hosts,
		shadowHosts: shadowHosts,
		projects:    projects,
		publisher:   publisher,
	}
}

// Handle оркестрирует добавление shadow-хоста в проект.
func (h *ProvisionHostFromShadowHandler) Handle(ctx context.Context, cmd ProvisionHostFromShadowCommand) (ProvisionHostFromShadowResult, error) {
	// 1. Загружаем shadow-хост
	shadowID := domain.ShadowHostID(cmd.ShadowHostID)
	sh, err := h.shadowHosts.FindByID(ctx, shadowID)
	if err != nil {
		return ProvisionHostFromShadowResult{}, fmt.Errorf("find shadow host: %w", err)
	}

	// 2. Доменная проверка: хост должен быть смонтирован
	if !sh.IsReadyForProvisioning() {
		return ProvisionHostFromShadowResult{},
			fmt.Errorf("shadow host %s is not ready for provisioning (status: %s)",
				cmd.ShadowHostID, sh.Status())
	}

	// 3. Загружаем проект и проверяем его тип
	projectID := domain.ProjectID(cmd.ProjectID)
	project, err := h.projects.FindByID(ctx, projectID)
	if err != nil {
		return ProvisionHostFromShadowResult{}, fmt.Errorf("find project: %w", err)
	}
	if !project.IsBareMetal() {
		return ProvisionHostFromShadowResult{}, domain.ErrInvalidProjectForHost
	}

	// 4. Проверяем уникальность FQDN
	exists, err := h.hosts.ExistsByFQDN(ctx, cmd.FQDN)
	if err != nil {
		return ProvisionHostFromShadowResult{}, fmt.Errorf("check fqdn uniqueness: %w", err)
	}
	if exists {
		return ProvisionHostFromShadowResult{}, domain.ErrHostAlreadyExists
	}

	// 5. Создаём Host-агрегат из данных ShadowHost
	hostID := domain.HostID(uuid.New().String())
	host, err := domain.NewHost(
		hostID,
		projectID,
		cmd.FQDN,
		sh.Inv(),
		sh.Kind(),
		sh.Location(),
		sh.Hardware(),
	)
	if err != nil {
		return ProvisionHostFromShadowResult{}, fmt.Errorf("create host aggregate: %w", err)
	}

	// 6. Сохраняем Host
	if err := h.hosts.Save(ctx, host); err != nil {
		return ProvisionHostFromShadowResult{}, fmt.Errorf("save host: %w", err)
	}

	// 7. Помечаем ShadowHost как Provisioned
	if err := sh.MarkProvisioned(hostID); err != nil {
		return ProvisionHostFromShadowResult{HostID: string(hostID)},
			fmt.Errorf("host created but shadow host mark failed: %w", err)
	}

	if err := h.shadowHosts.Update(ctx, sh); err != nil {
		return ProvisionHostFromShadowResult{HostID: string(hostID)},
			fmt.Errorf("host created but shadow host update failed: %w", err)
	}

	// 8. Публикуем события обоих агрегатов
	var allEvents []domain.DomainEvent
	allEvents = append(allEvents, host.PullEvents()...)
	allEvents = append(allEvents, sh.PullEvents()...)

	if len(allEvents) > 0 {
		if err := h.publisher.Publish(ctx, allEvents...); err != nil {
			return ProvisionHostFromShadowResult{HostID: string(hostID)},
				fmt.Errorf("provisioned but event publish failed: %w", err)
		}
	}

	return ProvisionHostFromShadowResult{HostID: string(hostID)}, nil
}
