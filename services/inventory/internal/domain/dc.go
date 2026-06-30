package domain

// DC — агрегат дата-центра: первоклассная локация-корень с CRUD (LOC-01/D-04), вершина
// иерархии DC→Module→Rack. Три локации — ТРИ независимых агрегата (D-04), НЕ один
// агрегат-дерево: маленькие агрегаты по канону DDD, ложатся на Kafka key=ID каждой
// сущности (Phase 8). Встраивает aggregateBase: каждая CRUD-операция эмитит ровно одно
// семантическое событие через record() (D-13/EVT-01). Ссылок на дочерние Module/Rack
// агрегат не держит — иерархия выражена ссылками снизу вверх по ID (D-06).
type DC struct {
	aggregateBase
	id       DCID   // внутренний постоянный идентификатор идентичности (INV-03)
	name     string // имя дата-центра (обязательное, LOC-01)
	location string // физическое расположение/адрес (опционально)
}

// NewDC — фабрика дата-центра (рецепт «NewX держит инварианты + record»): требует непустое
// name (LOC-01/V5), генерирует новый DCID (INV-03), эмитит ровно один DCCreated (EVT-01).
func NewDC(name, location string) (*DC, error) {
	if name == "" {
		return nil, ErrInvalidLocation // LOC-01: ДЦ без имени не существует
	}

	dc := &DC{
		id:       NewDCID(),
		name:     name,
		location: location,
	}
	dc.record(DCCreated{DCID: dc.id, Name: name, Location: location})
	return dc, nil
}

// ID возвращает идентификатор дата-центра.
func (dc *DC) ID() DCID { return dc.id }

// Name возвращает имя дата-центра.
func (dc *DC) Name() string { return dc.name }

// Location возвращает физическое расположение дата-центра.
func (dc *DC) Location() string { return dc.location }

// Update меняет редактируемые атрибуты дата-центра (CRUD-update, LOC-01) и эмитит ровно
// один DCUpdated. Пустое name отклоняется (инвариант обязательности сохраняется).
func (dc *DC) Update(name, location string) error {
	if name == "" {
		return ErrInvalidLocation // имя остаётся обязательным (LOC-01)
	}
	dc.name = name
	dc.location = location
	dc.record(DCUpdated{DCID: dc.id, Name: name, Location: location})
	return nil
}

// Delete — удаление дата-центра: эмитит ровно один DCDeleted. Инвариант непустоты
// (нет дочерних Module) enforced на уровне usecase, аналогично Project.Delete.
func (dc *DC) Delete() {
	dc.record(DCDeleted{DCID: dc.id, Name: dc.name})
}

// --- Доменные события DC (D-13: голый семантический факт). EntityID = строковый DCID. ---

// DCCreated — факт создания дата-центра (EVT-01).
type DCCreated struct {
	DCID     DCID
	Name     string
	Location string
}

func (e DCCreated) EventType() string { return "DCCreated" }
func (e DCCreated) EntityID() string  { return e.DCID.String() }

// DCUpdated — факт обновления атрибутов дата-центра (LOC-01).
type DCUpdated struct {
	DCID     DCID
	Name     string
	Location string
}

func (e DCUpdated) EventType() string { return "DCUpdated" }
func (e DCUpdated) EntityID() string  { return e.DCID.String() }

// DCDeleted — факт удаления дата-центра.
type DCDeleted struct {
	DCID DCID
	Name string
}

func (e DCDeleted) EventType() string { return "DCDeleted" }
func (e DCDeleted) EntityID() string  { return e.DCID.String() }
