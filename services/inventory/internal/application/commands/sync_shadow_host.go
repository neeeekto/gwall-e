package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// SyncShadowHostsCommand — команда запуска синхронизации из bot-инвентори.
type SyncShadowHostsCommand struct {
	// Since — unix timestamp для инкрементальной синхронизации.
	// Если 0 — полная синхронизация всех хостов.
	Since int64
}

// SyncShadowHostsResult — результат синхронизации.
type SyncShadowHostsResult struct {
	Created int
	Updated int
	Mounted int
}

// SyncShadowHostsHandler — обработчик синхронизации из bot-инвентори.
type SyncShadowHostsHandler struct {
	shadowHosts domain.ShadowHostRepository
	botClient   domain.BotInventoryClient
	publisher   domain.EventPublisher
}

func NewSyncShadowHostsHandler(
	shadowHosts domain.ShadowHostRepository,
	botClient domain.BotInventoryClient,
	publisher domain.EventPublisher,
) *SyncShadowHostsHandler {
	return &SyncShadowHostsHandler{
		shadowHosts: shadowHosts,
		botClient:   botClient,
		publisher:   publisher,
	}
}

// Handle выполняет синхронизацию shadow-хостов из bot-инвентори.
func (h *SyncShadowHostsHandler) Handle(ctx context.Context, cmd SyncShadowHostsCommand) (SyncShadowHostsResult, error) {
	var records []domain.BotHostRecord
	var err error

	if cmd.Since > 0 {
		records, err = h.botClient.FetchUpdatedSince(ctx, cmd.Since)
	} else {
		records, err = h.botClient.FetchAll(ctx)
	}
	if err != nil {
		return SyncShadowHostsResult{}, fmt.Errorf("fetch from bot inventory: %w", err)
	}

	result := SyncShadowHostsResult{}

	for _, record := range records {
		if err := h.syncRecord(ctx, record, &result); err != nil {
			fmt.Printf("sync shadow host %s: %v\n", record.ExternalID, err)
		}
	}

	return result, nil
}

func (h *SyncShadowHostsHandler) syncRecord(ctx context.Context, record domain.BotHostRecord, result *SyncShadowHostsResult) error {
	kind, err := domain.ParseHostKind(record.Kind)
	if err != nil {
		return fmt.Errorf("parse host kind %q: %w", record.Kind, err)
	}

	location := domain.Location{
		Country:  record.Location.Country,
		City:     record.Location.City,
		Building: record.Location.Building,
		Module:   record.Location.Module,
		Rack:     record.Location.Rack,
		Unit:     record.Location.Unit,
	}

	hardware := domain.HostHardware{
		Name:        record.Hardware.Name,
		Platform:    record.Hardware.Platform,
		IPMIMac:     record.Hardware.IPMIMac,
		Motherboard: record.Hardware.Motherboard,
		MACs:        record.Hardware.MACs,
	}

	existing, err := h.shadowHosts.FindByExternalID(ctx, record.ExternalID)
	if err != nil && err != domain.ErrHostNotFound {
		return fmt.Errorf("find shadow host by external id: %w", err)
	}

	if existing == nil {
		return h.createShadowHost(ctx, record, kind, location, hardware, result)
	}

	return h.updateShadowHost(ctx, existing, record, location, hardware, result)
}

func (h *SyncShadowHostsHandler) createShadowHost(
	ctx context.Context,
	record domain.BotHostRecord,
	kind domain.HostKind,
	location domain.Location,
	hardware domain.HostHardware,
	result *SyncShadowHostsResult,
) error {
	id := domain.ShadowHostID(uuid.New().String())
	sh, err := domain.NewShadowHost(id, record.ExternalID, record.Inv, kind, location, hardware)
	if err != nil {
		return fmt.Errorf("create shadow host: %w", err)
	}

	if record.IsMounted {
		if err := sh.MarkMounted(location); err != nil {
			return fmt.Errorf("mark mounted: %w", err)
		}
		result.Mounted++
	}

	if err := h.shadowHosts.Save(ctx, sh); err != nil {
		return fmt.Errorf("save shadow host: %w", err)
	}

	events := sh.PullEvents()
	if len(events) > 0 {
		_ = h.publisher.Publish(ctx, events...)
	}

	result.Created++
	return nil
}

func (h *SyncShadowHostsHandler) updateShadowHost(
	ctx context.Context,
	sh *domain.ShadowHost,
	record domain.BotHostRecord,
	location domain.Location,
	hardware domain.HostHardware,
	result *SyncShadowHostsResult,
) error {
	if err := sh.SyncHardware(hardware, location); err != nil {
		return fmt.Errorf("sync hardware: %w", err)
	}

	if record.IsMounted && sh.Status() == domain.ShadowHostStatusDiscovered {
		if err := sh.MarkMounted(location); err != nil {
			return fmt.Errorf("mark mounted: %w", err)
		}
		result.Mounted++
	}

	if err := h.shadowHosts.Update(ctx, sh); err != nil {
		return fmt.Errorf("update shadow host: %w", err)
	}

	events := sh.PullEvents()
	if len(events) > 0 {
		_ = h.publisher.Publish(ctx, events...)
	}

	result.Updated++
	return nil
}
