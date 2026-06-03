package domain

import "time"

// RackID — типизированный идентификатор стойки.
type RackID string

// Rack — агрегат физической иерархии.
// Представляет серверную стойку внутри модуля дата-центра.
type Rack struct {
	id        RackID
	moduleID  ModuleID
	name      string
	capacity  int // количество юнитов (U)
	createdAt time.Time
	updatedAt time.Time
}

// NewRack — фабричный метод с валидацией.
func NewRack(id RackID, moduleID ModuleID, name string, capacity int) (*Rack, error) {
	if name == "" {
		return nil, ErrInvalidRackName
	}
	if capacity <= 0 {
		return nil, ErrInvalidRackCapacity
	}
	now := time.Now().UTC()
	return &Rack{
		id:        id,
		moduleID:  moduleID,
		name:      name,
		capacity:  capacity,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// RestoreRackParams — параметры восстановления из хранилища.
type RestoreRackParams struct {
	ID        RackID
	ModuleID  ModuleID
	Name      string
	Capacity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RestoreRack — конструктор для восстановления из персистентного хранилища.
func RestoreRack(p RestoreRackParams) *Rack {
	return &Rack{
		id:        p.ID,
		moduleID:  p.ModuleID,
		name:      p.Name,
		capacity:  p.Capacity,
		createdAt: p.CreatedAt,
		updatedAt: p.UpdatedAt,
	}
}

// --- Геттеры ---

func (r *Rack) ID() RackID           { return r.id }
func (r *Rack) ModuleID() ModuleID   { return r.moduleID }
func (r *Rack) Name() string         { return r.name }
func (r *Rack) Capacity() int        { return r.capacity }
func (r *Rack) CreatedAt() time.Time { return r.createdAt }
func (r *Rack) UpdatedAt() time.Time { return r.updatedAt }
