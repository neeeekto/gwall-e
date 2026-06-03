package domain

import "time"

// DataCenterID — типизированный идентификатор дата-центра.
type DataCenterID string

// DataCenter — агрегат физической иерархии.
// Представляет дата-центр как физическую единицу размещения оборудования.
type DataCenter struct {
	id        DataCenterID
	name      string
	location  string // город/адрес ДЦ
	createdAt time.Time
	updatedAt time.Time
}

// NewDataCenter — фабричный метод с валидацией.
func NewDataCenter(id DataCenterID, name string, location string) (*DataCenter, error) {
	if name == "" {
		return nil, ErrInvalidDataCenterName
	}
	now := time.Now().UTC()
	return &DataCenter{
		id:        id,
		name:      name,
		location:  location,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// RestoreDataCenterParams — параметры восстановления из хранилища.
type RestoreDataCenterParams struct {
	ID        DataCenterID
	Name      string
	Location  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RestoreDataCenter — конструктор для восстановления из персистентного хранилища.
func RestoreDataCenter(p RestoreDataCenterParams) *DataCenter {
	return &DataCenter{
		id:        p.ID,
		name:      p.Name,
		location:  p.Location,
		createdAt: p.CreatedAt,
		updatedAt: p.UpdatedAt,
	}
}

// --- Геттеры ---

func (dc *DataCenter) ID() DataCenterID     { return dc.id }
func (dc *DataCenter) Name() string         { return dc.name }
func (dc *DataCenter) Location() string     { return dc.location }
func (dc *DataCenter) CreatedAt() time.Time { return dc.createdAt }
func (dc *DataCenter) UpdatedAt() time.Time { return dc.updatedAt }
