package domain

import "time"

// ModuleID — типизированный идентификатор модуля (зала) в ДЦ.
type ModuleID string

// Module — агрегат физической иерархии.
// Представляет зал/модуль внутри дата-центра.
type Module struct {
	id           ModuleID
	dataCenterID DataCenterID
	name         string
	createdAt    time.Time
	updatedAt    time.Time
}

// NewModule — фабричный метод с валидацией.
func NewModule(id ModuleID, dataCenterID DataCenterID, name string) (*Module, error) {
	if name == "" {
		return nil, ErrInvalidModuleName
	}
	now := time.Now().UTC()
	return &Module{
		id:           id,
		dataCenterID: dataCenterID,
		name:         name,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// RestoreModuleParams — параметры восстановления из хранилища.
type RestoreModuleParams struct {
	ID           ModuleID
	DataCenterID DataCenterID
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RestoreModule — конструктор для восстановления из персистентного хранилища.
func RestoreModule(p RestoreModuleParams) *Module {
	return &Module{
		id:           p.ID,
		dataCenterID: p.DataCenterID,
		name:         p.Name,
		createdAt:    p.CreatedAt,
		updatedAt:    p.UpdatedAt,
	}
}

// --- Геттеры ---

func (m *Module) ID() ModuleID               { return m.id }
func (m *Module) DataCenterID() DataCenterID { return m.dataCenterID }
func (m *Module) Name() string               { return m.name }
func (m *Module) CreatedAt() time.Time       { return m.createdAt }
func (m *Module) UpdatedAt() time.Time       { return m.updatedAt }
