package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// Типизированные ID-VO агрегатов (D-05): каждый — обёртка над uuid.UUID, а не «голая»
// строка (переопределяет прецедент ExampleID `type X string`). Компилятор различает
// HostID/ProjectID/DCID/ModuleID/RackID и ловит перепутанную ссылку (T-06-01). ID —
// внутренний постоянный идентификатор идентичности; v4-random, непереиспользуемый (INV-03).

// HostID — идентификатор хоста.
type HostID struct{ v uuid.UUID }

// NewHostID порождает новый случайный (v4) идентификатор хоста (INV-03).
func NewHostID() HostID { return HostID{v: uuid.New()} }

// ParseHostID восстанавливает HostID из строки; битый вход оборачивает ErrInvalidID (%w).
func ParseHostID(s string) (HostID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return HostID{}, fmt.Errorf("parse host id %q: %w", s, ErrInvalidID)
	}
	return HostID{v: u}, nil
}

// String отдаёт каноническое строковое представление (будущий Kafka-ключ Phase 8, логи).
func (id HostID) String() string { return id.v.String() }

// IsZero сообщает, является ли ID zero-value (guard в инвариантах фабрик).
func (id HostID) IsZero() bool { return id.v == uuid.Nil }

// ProjectID — идентификатор проекта.
type ProjectID struct{ v uuid.UUID }

// NewProjectID порождает новый случайный (v4) идентификатор проекта (INV-03).
func NewProjectID() ProjectID { return ProjectID{v: uuid.New()} }

// ParseProjectID восстанавливает ProjectID из строки; битый вход оборачивает ErrInvalidID.
func ParseProjectID(s string) (ProjectID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ProjectID{}, fmt.Errorf("parse project id %q: %w", s, ErrInvalidID)
	}
	return ProjectID{v: u}, nil
}

// String отдаёт каноническое строковое представление.
func (id ProjectID) String() string { return id.v.String() }

// IsZero сообщает, является ли ID zero-value.
func (id ProjectID) IsZero() bool { return id.v == uuid.Nil }

// DCID — идентификатор дата-центра.
type DCID struct{ v uuid.UUID }

// NewDCID порождает новый случайный (v4) идентификатор дата-центра (INV-03).
func NewDCID() DCID { return DCID{v: uuid.New()} }

// ParseDCID восстанавливает DCID из строки; битый вход оборачивает ErrInvalidID.
func ParseDCID(s string) (DCID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return DCID{}, fmt.Errorf("parse dc id %q: %w", s, ErrInvalidID)
	}
	return DCID{v: u}, nil
}

// String отдаёт каноническое строковое представление.
func (id DCID) String() string { return id.v.String() }

// IsZero сообщает, является ли ID zero-value.
func (id DCID) IsZero() bool { return id.v == uuid.Nil }

// ModuleID — идентификатор модуля (зала) внутри дата-центра.
type ModuleID struct{ v uuid.UUID }

// NewModuleID порождает новый случайный (v4) идентификатор модуля (INV-03).
func NewModuleID() ModuleID { return ModuleID{v: uuid.New()} }

// ParseModuleID восстанавливает ModuleID из строки; битый вход оборачивает ErrInvalidID.
func ParseModuleID(s string) (ModuleID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ModuleID{}, fmt.Errorf("parse module id %q: %w", s, ErrInvalidID)
	}
	return ModuleID{v: u}, nil
}

// String отдаёт каноническое строковое представление.
func (id ModuleID) String() string { return id.v.String() }

// IsZero сообщает, является ли ID zero-value.
func (id ModuleID) IsZero() bool { return id.v == uuid.Nil }

// RackID — идентификатор стойки внутри модуля.
type RackID struct{ v uuid.UUID }

// NewRackID порождает новый случайный (v4) идентификатор стойки (INV-03).
func NewRackID() RackID { return RackID{v: uuid.New()} }

// ParseRackID восстанавливает RackID из строки; битый вход оборачивает ErrInvalidID.
func ParseRackID(s string) (RackID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return RackID{}, fmt.Errorf("parse rack id %q: %w", s, ErrInvalidID)
	}
	return RackID{v: u}, nil
}

// String отдаёт каноническое строковое представление.
func (id RackID) String() string { return id.v.String() }

// IsZero сообщает, является ли ID zero-value.
func (id RackID) IsZero() bool { return id.v == uuid.Nil }
