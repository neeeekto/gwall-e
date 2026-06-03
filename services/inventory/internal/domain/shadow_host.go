package domain

import (
	"fmt"
	"time"
)

// ShadowHostID — типизированный идентификатор shadow-хоста.
type ShadowHostID string

// ShadowHostStatus — жизненный цикл shadow-хоста.
//
// Переходы:
//
//	Discovered → Mounted → Provisioned
type ShadowHostStatus int

const (
	ShadowHostStatusDiscovered ShadowHostStatus = iota
	ShadowHostStatusMounted
	ShadowHostStatusProvisioned
)

func (s ShadowHostStatus) String() string {
	switch s {
	case ShadowHostStatusDiscovered:
		return "discovered"
	case ShadowHostStatusMounted:
		return "mounted"
	case ShadowHostStatusProvisioned:
		return "provisioned"
	default:
		return "unknown"
	}
}

// ShadowHost — агрегат, представляющий хост из внешней инвентори (bot).
type ShadowHost struct {
	AggregateRoot

	id              ShadowHostID
	externalID      string
	inv             int
	kind            HostKind
	location        Location
	hardware        HostHardware
	status          ShadowHostStatus
	provisionedAsID HostID
	lastSyncedAt    time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

// NewShadowHost — фабричный метод. Вызывается при первой синхронизации из bot.
func NewShadowHost(
	id ShadowHostID,
	externalID string,
	inv int,
	kind HostKind,
	location Location,
	hardware HostHardware,
) (*ShadowHost, error) {
	if externalID == "" {
		return nil, fmt.Errorf("external id cannot be empty")
	}
	if inv <= 0 {
		return nil, ErrInvalidInv
	}

	now := time.Now().UTC()
	sh := &ShadowHost{
		id:           id,
		externalID:   externalID,
		inv:          inv,
		kind:         kind,
		location:     location,
		hardware:     hardware,
		status:       ShadowHostStatusDiscovered,
		lastSyncedAt: now,
		createdAt:    now,
		updatedAt:    now,
	}
	sh.RecordEvent(NewShadowHostDiscovered(id, externalID, inv))
	return sh, nil
}

// RestoreShadowHostParams — параметры восстановления из хранилища.
type RestoreShadowHostParams struct {
	ID              ShadowHostID
	ExternalID      string
	Inv             int
	Kind            HostKind
	Location        Location
	Hardware        HostHardware
	Status          ShadowHostStatus
	ProvisionedAsID HostID
	LastSyncedAt    time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// RestoreShadowHost — конструктор для восстановления из хранилища.
func RestoreShadowHost(p RestoreShadowHostParams) *ShadowHost {
	return &ShadowHost{
		id:              p.ID,
		externalID:      p.ExternalID,
		inv:             p.Inv,
		kind:            p.Kind,
		location:        p.Location,
		hardware:        p.Hardware,
		status:          p.Status,
		provisionedAsID: p.ProvisionedAsID,
		lastSyncedAt:    p.LastSyncedAt,
		createdAt:       p.CreatedAt,
		updatedAt:       p.UpdatedAt,
	}
}

// --- Геттеры ---

func (sh *ShadowHost) ID() ShadowHostID         { return sh.id }
func (sh *ShadowHost) ExternalID() string       { return sh.externalID }
func (sh *ShadowHost) Inv() int                 { return sh.inv }
func (sh *ShadowHost) Kind() HostKind           { return sh.kind }
func (sh *ShadowHost) Location() Location       { return sh.location }
func (sh *ShadowHost) Hardware() HostHardware   { return sh.hardware }
func (sh *ShadowHost) Status() ShadowHostStatus { return sh.status }
func (sh *ShadowHost) ProvisionedAsID() HostID  { return sh.provisionedAsID }
func (sh *ShadowHost) LastSyncedAt() time.Time  { return sh.lastSyncedAt }
func (sh *ShadowHost) CreatedAt() time.Time     { return sh.createdAt }
func (sh *ShadowHost) UpdatedAt() time.Time     { return sh.updatedAt }

// IsReadyForProvisioning возвращает true если хост можно добавить в проект.
func (sh *ShadowHost) IsReadyForProvisioning() bool {
	return sh.status == ShadowHostStatusMounted
}

// --- Доменные методы ---

// MarkMounted переводит хост в статус "смонтирован в стойку".
func (sh *ShadowHost) MarkMounted(location Location) error {
	if sh.status != ShadowHostStatusDiscovered {
		return fmt.Errorf("%w: cannot mount shadow host with status %s", ErrInvalidStatus, sh.status)
	}
	sh.location = location
	sh.status = ShadowHostStatusMounted
	sh.updatedAt = time.Now().UTC()
	sh.RecordEvent(NewShadowHostMounted(sh.id, sh.externalID, location))
	return nil
}

// SyncHardware обновляет аппаратную конфигурацию из bot-инвентори.
func (sh *ShadowHost) SyncHardware(hardware HostHardware, location Location) error {
	if sh.status == ShadowHostStatusProvisioned {
		return fmt.Errorf("%w: cannot sync hardware of provisioned shadow host", ErrInvalidStatus)
	}
	sh.hardware = hardware
	sh.location = location
	sh.lastSyncedAt = time.Now().UTC()
	sh.updatedAt = time.Now().UTC()
	return nil
}

// MarkProvisioned помечает shadow-хост как добавленный в проект.
func (sh *ShadowHost) MarkProvisioned(hostID HostID) error {
	if sh.status != ShadowHostStatusMounted {
		return fmt.Errorf("%w: cannot provision shadow host with status %s", ErrInvalidStatus, sh.status)
	}
	sh.status = ShadowHostStatusProvisioned
	sh.provisionedAsID = hostID
	sh.updatedAt = time.Now().UTC()
	sh.RecordEvent(NewShadowHostProvisioned(sh.id, sh.externalID, hostID))
	return nil
}
