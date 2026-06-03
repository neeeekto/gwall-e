package domain

import (
	"context"
	"time"
)

// DomainEvent — базовый интерфейс для всех доменных событий.
// Событие — факт, который уже произошёл. Имя в прошедшем времени.
type DomainEvent interface {
	// EventName возвращает уникальное имя события для роутинга.
	EventName() string
	// OccurredAt возвращает время возникновения события.
	OccurredAt() time.Time
}

// EventPublisher — порт для публикации доменных событий.
// Определён в домене, реализован в infra (Kafka, NATS, in-memory и т.д.).
// Handler вызывает Publish ПОСЛЕ успешного сохранения агрегата.
type EventPublisher interface {
	Publish(ctx context.Context, events ...DomainEvent) error
}

// HostRegistered — событие: хост зарегистрирован в инвентаре.
type HostRegistered struct {
	HostID     HostID
	ProjectID  ProjectID
	FQDN       string
	Kind       HostKind
	occurredAt time.Time
}

func NewHostRegistered(hostID HostID, projectID ProjectID, fqdn string, kind HostKind) HostRegistered {
	return HostRegistered{
		HostID:     hostID,
		ProjectID:  projectID,
		FQDN:       fqdn,
		Kind:       kind,
		occurredAt: time.Now().UTC(),
	}
}

func (e HostRegistered) EventName() string     { return "inventory.host.registered" }
func (e HostRegistered) OccurredAt() time.Time { return e.occurredAt }

// HostActivated — событие: хост введён в эксплуатацию.
type HostActivated struct {
	HostID     HostID
	ProjectID  ProjectID
	occurredAt time.Time
}

func NewHostActivated(hostID HostID, projectID ProjectID) HostActivated {
	return HostActivated{
		HostID:     hostID,
		ProjectID:  projectID,
		occurredAt: time.Now().UTC(),
	}
}

func (e HostActivated) EventName() string     { return "inventory.host.activated" }
func (e HostActivated) OccurredAt() time.Time { return e.occurredAt }

// HostDecommissioned — событие: хост выведен из эксплуатации.
type HostDecommissioned struct {
	HostID     HostID
	ProjectID  ProjectID
	occurredAt time.Time
}

func NewHostDecommissioned(hostID HostID, projectID ProjectID) HostDecommissioned {
	return HostDecommissioned{
		HostID:     hostID,
		ProjectID:  projectID,
		occurredAt: time.Now().UTC(),
	}
}

func (e HostDecommissioned) EventName() string     { return "inventory.host.decommissioned" }
func (e HostDecommissioned) OccurredAt() time.Time { return e.occurredAt }

// ShadowHostDiscovered — событие: хост обнаружен в bot-инвентори и синхронизирован.
type ShadowHostDiscovered struct {
	ShadowHostID ShadowHostID
	ExternalID   string
	Inv          int
	occurredAt   time.Time
}

func NewShadowHostDiscovered(id ShadowHostID, externalID string, inv int) ShadowHostDiscovered {
	return ShadowHostDiscovered{
		ShadowHostID: id,
		ExternalID:   externalID,
		Inv:          inv,
		occurredAt:   time.Now().UTC(),
	}
}

func (e ShadowHostDiscovered) EventName() string     { return "inventory.shadow_host.discovered" }
func (e ShadowHostDiscovered) OccurredAt() time.Time { return e.occurredAt }

// ShadowHostMounted — событие: хост физически смонтирован в стойку.
type ShadowHostMounted struct {
	ShadowHostID ShadowHostID
	ExternalID   string
	Location     Location
	occurredAt   time.Time
}

func NewShadowHostMounted(id ShadowHostID, externalID string, location Location) ShadowHostMounted {
	return ShadowHostMounted{
		ShadowHostID: id,
		ExternalID:   externalID,
		Location:     location,
		occurredAt:   time.Now().UTC(),
	}
}

func (e ShadowHostMounted) EventName() string     { return "inventory.shadow_host.mounted" }
func (e ShadowHostMounted) OccurredAt() time.Time { return e.occurredAt }

// ShadowHostProvisioned — событие: shadow-хост добавлен в проект, создан Host.
type ShadowHostProvisioned struct {
	ShadowHostID ShadowHostID
	ExternalID   string
	HostID       HostID
	occurredAt   time.Time
}

func NewShadowHostProvisioned(id ShadowHostID, externalID string, hostID HostID) ShadowHostProvisioned {
	return ShadowHostProvisioned{
		ShadowHostID: id,
		ExternalID:   externalID,
		HostID:       hostID,
		occurredAt:   time.Now().UTC(),
	}
}

func (e ShadowHostProvisioned) EventName() string     { return "inventory.shadow_host.provisioned" }
func (e ShadowHostProvisioned) OccurredAt() time.Time { return e.occurredAt }
