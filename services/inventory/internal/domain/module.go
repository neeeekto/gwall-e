package domain

// Module — агрегат модуля (зала) внутри дата-центра: первоклассная локация-корень с CRUD
// (LOC-01/D-04), средний уровень иерархии DC→Module→Rack. Независимый агрегат (D-04), НЕ
// часть дерева-DC. Несёт ссылку на родительский дата-центр ТОЛЬКО по внутреннему DCID
// (LOC-02/D-06) — не вложенный DC-объект, не back-reference. Встраивает aggregateBase:
// каждая CRUD-операция эмитит ровно одно семантическое событие (D-13/EVT-01).
type Module struct {
	aggregateBase
	id   ModuleID // внутренний постоянный идентификатор идентичности (INV-03)
	dcID DCID     // привязка к родительскому дата-центру по внутреннему ID (LOC-02/D-06)
	name string   // имя модуля/зала (обязательное, LOC-01)
}

// NewModule — фабрика модуля (рецепт «NewX держит инварианты + record»): guard на zero DCID
// (привязка к ДЦ обязательна — нет «висячего» модуля без родителя, LOC-02/D-06/T-06-09),
// требует непустое name (LOC-01/V5), генерирует новый ModuleID (INV-03), эмитит ровно один
// ModuleCreated (EVT-01).
func NewModule(dcID DCID, name string) (*Module, error) {
	if dcID.IsZero() {
		return nil, ErrInvalidLocation // нет модуля без родительского ДЦ (LOC-02/D-06)
	}
	if name == "" {
		return nil, ErrInvalidLocation // LOC-01: модуль без имени не существует
	}

	m := &Module{
		id:   NewModuleID(),
		dcID: dcID,
		name: name,
	}
	m.record(ModuleCreated{ModuleID: m.id, DCID: dcID, Name: name})
	return m, nil
}

// ID возвращает идентификатор модуля.
func (m *Module) ID() ModuleID { return m.id }

// DCID возвращает привязку к родительскому дата-центру (иерархия по ID, LOC-02/D-06).
func (m *Module) DCID() DCID { return m.dcID }

// Name возвращает имя модуля.
func (m *Module) Name() string { return m.name }

// Update меняет редактируемые атрибуты модуля (CRUD-update, LOC-01) и эмитит ровно один
// ModuleUpdated. Родительский DCID неизменяем (перенос модуля между ДЦ вне scope CRUD).
func (m *Module) Update(name string) error {
	if name == "" {
		return ErrInvalidLocation // имя остаётся обязательным (LOC-01)
	}
	m.name = name
	m.record(ModuleUpdated{ModuleID: m.id, Name: name})
	return nil
}

// Delete — удаление модуля: эмитит ровно один ModuleDeleted. Инвариант непустоты
// (нет дочерних Rack) enforced на уровне usecase.
func (m *Module) Delete() {
	m.record(ModuleDeleted{ModuleID: m.id, DCID: m.dcID, Name: m.name})
}

// --- Доменные события Module (D-13: голый семантический факт). EntityID = строковый
// ModuleID. ---

// ModuleCreated — факт создания модуля (EVT-01): несёт привязку к родительскому ДЦ.
type ModuleCreated struct {
	ModuleID ModuleID
	DCID     DCID
	Name     string
}

func (e ModuleCreated) EventType() string { return "ModuleCreated" }
func (e ModuleCreated) EntityID() string  { return e.ModuleID.String() }

// ModuleUpdated — факт обновления атрибутов модуля (LOC-01).
type ModuleUpdated struct {
	ModuleID ModuleID
	Name     string
}

func (e ModuleUpdated) EventType() string { return "ModuleUpdated" }
func (e ModuleUpdated) EntityID() string  { return e.ModuleID.String() }

// ModuleDeleted — факт удаления модуля: несёт привязку к ДЦ для аудит-следа.
type ModuleDeleted struct {
	ModuleID ModuleID
	DCID     DCID
	Name     string
}

func (e ModuleDeleted) EventType() string { return "ModuleDeleted" }
func (e ModuleDeleted) EntityID() string  { return e.ModuleID.String() }
