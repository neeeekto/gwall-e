package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// RegisterHostCommand — входные данные команды (DTO).
type RegisterHostCommand struct {
	ProjectID string
	FQDN      string
	Inv       int
	Kind      string // "server" | "mac"
	Location  RegisterHostLocationDTO
	Hardware  RegisterHostHardwareDTO
}

type RegisterHostLocationDTO struct {
	Country  string
	City     string
	Building string
	Module   string
	Rack     string
	Unit     string
	Object   string
	RoomType string
}

type RegisterHostHardwareDTO struct {
	Name        string
	Platform    string
	IPMIMac     string
	Motherboard string
	MACs        []string
}

// RegisterHostResult — результат выполнения команды.
type RegisterHostResult struct {
	HostID string
}

// RegisterHostHandler — обработчик команды регистрации хоста.
// Использует ProjectRepository напрямую — нет нужды в отдельном ProjectChecker порте.
type RegisterHostHandler struct {
	hosts     domain.HostRepository
	projects  domain.ProjectRepository
	publisher domain.EventPublisher
}

func NewRegisterHostHandler(
	hosts domain.HostRepository,
	projects domain.ProjectRepository,
	publisher domain.EventPublisher,
) *RegisterHostHandler {
	return &RegisterHostHandler{
		hosts:     hosts,
		projects:  projects,
		publisher: publisher,
	}
}

// Handle оркестрирует выполнение команды.
func (h *RegisterHostHandler) Handle(ctx context.Context, cmd RegisterHostCommand) (RegisterHostResult, error) {
	projectID := domain.ProjectID(cmd.ProjectID)

	// 1. Загружаем проект и проверяем его тип
	project, err := h.projects.FindByID(ctx, projectID)
	if err != nil {
		return RegisterHostResult{}, fmt.Errorf("find project: %w", err)
	}

	// 2. Доменное правило: физический хост только в bare-metal проекте
	if !project.IsBareMetal() {
		return RegisterHostResult{}, domain.ErrInvalidProjectForHost
	}

	// 3. Проверяем уникальность FQDN
	exists, err := h.hosts.ExistsByFQDN(ctx, cmd.FQDN)
	if err != nil {
		return RegisterHostResult{}, fmt.Errorf("check fqdn uniqueness: %w", err)
	}
	if exists {
		return RegisterHostResult{}, domain.ErrHostAlreadyExists
	}

	// 4. Маппинг DTO → доменные типы
	kind, err := domain.ParseHostKind(cmd.Kind)
	if err != nil {
		return RegisterHostResult{}, fmt.Errorf("parse host kind: %w", err)
	}

	location := domain.Location{
		Country:  cmd.Location.Country,
		City:     cmd.Location.City,
		Building: cmd.Location.Building,
		Module:   cmd.Location.Module,
		Rack:     cmd.Location.Rack,
		Unit:     cmd.Location.Unit,
		Object:   cmd.Location.Object,
		RoomType: cmd.Location.RoomType,
	}

	hardware := domain.HostHardware{
		Name:        cmd.Hardware.Name,
		Platform:    cmd.Hardware.Platform,
		IPMIMac:     cmd.Hardware.IPMIMac,
		Motherboard: cmd.Hardware.Motherboard,
		MACs:        cmd.Hardware.MACs,
	}

	// 5. Создаём агрегат через фабричный метод
	hostID := domain.HostID(uuid.New().String())
	host, err := domain.NewHost(hostID, projectID, cmd.FQDN, cmd.Inv, kind, location, hardware)
	if err != nil {
		return RegisterHostResult{}, fmt.Errorf("create host aggregate: %w", err)
	}

	// 6. Сохраняем через репозиторий
	if err := h.hosts.Save(ctx, host); err != nil {
		return RegisterHostResult{}, fmt.Errorf("save host: %w", err)
	}

	// 7. Публикуем накопленные события после успешного Save
	events := host.PullEvents()
	if len(events) > 0 {
		if err := h.publisher.Publish(ctx, events...); err != nil {
			return RegisterHostResult{HostID: string(host.ID())},
				fmt.Errorf("host saved but event publish failed: %w", err)
		}
	}

	return RegisterHostResult{HostID: string(host.ID())}, nil
}
