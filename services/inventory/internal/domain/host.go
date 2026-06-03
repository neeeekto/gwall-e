package domain

import (
	"fmt"
	"time"
)

// HostID — типизированный идентификатор хоста.
type HostID string

// ProjectID — ссылка на проект (только ID, не сам агрегат).
type ProjectID string

// HostStatus — жизненный цикл хоста.
type HostStatus int

const (
	HostStatusPending        HostStatus = iota // зарегистрирован, ещё не активен
	HostStatusActive                           // введён в эксплуатацию
	HostStatusDecommissioned                   // выведен из эксплуатации
)

func (s HostStatus) String() string {
	switch s {
	case HostStatusPending:
		return "pending"
	case HostStatusActive:
		return "active"
	case HostStatusDecommissioned:
		return "decommissioned"
	default:
		return "unknown"
	}
}

// Host — агрегат (Aggregate Root) инвентаря.
// Все поля приватные — агрегат контролирует доступ и инварианты.
// Встраивает AggregateRoot для накопления доменных событий.
type Host struct {
	AggregateRoot

	id        HostID
	projectID ProjectID
	fqdn      string
	inv       int
	kind      HostKind
	tags      []string
	location  Location
	hardware  HostHardware
	status    HostStatus
	createdAt time.Time
	updatedAt time.Time
}

// NewHost — фабричный метод для создания нового хоста.
func NewHost(
	id HostID,
	projectID ProjectID,
	fqdn string,
	inv int,
	kind HostKind,
	location Location,
	hardware HostHardware,
) (*Host, error) {
	if fqdn == "" {
		return nil, ErrInvalidFQDN
	}
	if inv <= 0 {
		return nil, ErrInvalidInv
	}
	if kind == HostKindUnknown {
		return nil, ErrInvalidHostKind
	}

	now := time.Now().UTC()
	h := &Host{
		id:        id,
		projectID: projectID,
		fqdn:      fqdn,
		inv:       inv,
		kind:      kind,
		location:  location,
		hardware:  hardware,
		status:    HostStatusPending,
		createdAt: now,
		updatedAt: now,
	}
	h.RecordEvent(NewHostRegistered(id, projectID, fqdn, kind))
	return h, nil
}

// RestoreHostParams — параметры для восстановления агрегата из хранилища.
type RestoreHostParams struct {
	ID        HostID
	ProjectID ProjectID
	FQDN      string
	Inv       int
	Kind      HostKind
	Tags      []string
	Location  Location
	Hardware  HostHardware
	Status    HostStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RestoreHost — конструктор для восстановления агрегата из персистентного хранилища.
func RestoreHost(p RestoreHostParams) *Host {
	return &Host{
		id:        p.ID,
		projectID: p.ProjectID,
		fqdn:      p.FQDN,
		inv:       p.Inv,
		kind:      p.Kind,
		tags:      p.Tags,
		location:  p.Location,
		hardware:  p.Hardware,
		status:    p.Status,
		createdAt: p.CreatedAt,
		updatedAt: p.UpdatedAt,
	}
}

// --- Геттеры ---

func (h *Host) ID() HostID             { return h.id }
func (h *Host) ProjectID() ProjectID   { return h.projectID }
func (h *Host) FQDN() string           { return h.fqdn }
func (h *Host) Inv() int               { return h.inv }
func (h *Host) Kind() HostKind         { return h.kind }
func (h *Host) Status() HostStatus     { return h.status }
func (h *Host) Location() Location     { return h.location }
func (h *Host) Hardware() HostHardware { return h.hardware }
func (h *Host) CreatedAt() time.Time   { return h.createdAt }
func (h *Host) UpdatedAt() time.Time   { return h.updatedAt }

func (h *Host) Tags() []string {
	if h.tags == nil {
		return nil
	}
	result := make([]string, len(h.tags))
	copy(result, h.tags)
	return result
}

// --- Доменные методы ---

// Activate переводит хост в активное состояние.
func (h *Host) Activate() error {
	if h.status != HostStatusPending {
		return fmt.Errorf("%w: cannot activate host with status %s", ErrInvalidStatus, h.status)
	}
	h.status = HostStatusActive
	h.updatedAt = time.Now().UTC()
	h.RecordEvent(NewHostActivated(h.id, h.projectID))
	return nil
}

// Decommission выводит хост из эксплуатации.
func (h *Host) Decommission() error {
	if h.status != HostStatusActive {
		return fmt.Errorf("%w: cannot decommission host with status %s", ErrInvalidStatus, h.status)
	}
	h.status = HostStatusDecommissioned
	h.updatedAt = time.Now().UTC()
	h.RecordEvent(NewHostDecommissioned(h.id, h.projectID))
	return nil
}

// AddTag добавляет тег идемпотентно.
func (h *Host) AddTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}
	for _, t := range h.tags {
		if t == tag {
			return nil
		}
	}
	h.tags = append(h.tags, tag)
	h.updatedAt = time.Now().UTC()
	return nil
}

// RemoveTag удаляет тег идемпотентно.
func (h *Host) RemoveTag(tag string) {
	filtered := h.tags[:0]
	for _, t := range h.tags {
		if t != tag {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) != len(h.tags) {
		h.tags = filtered
		h.updatedAt = time.Now().UTC()
	}
}

// UpdateHardware обновляет аппаратную конфигурацию хоста.
func (h *Host) UpdateHardware(hardware HostHardware) error {
	if h.status == HostStatusDecommissioned {
		return fmt.Errorf("%w: cannot update hardware of decommissioned host", ErrInvalidStatus)
	}
	h.hardware = hardware
	h.updatedAt = time.Now().UTC()
	return nil
}
