package domain

// PowerTopology — маленький value-object топологических атрибутов стойки (LOC-04): источник
// питания и резервный генератор, к которым подключена стойка. Это предметная топология
// энергоснабжения (а не состав железа хостов): на ней строится анализ влияния отказа
// питания/генератора на парк. Поля — непрозрачные внешние идентификаторы линий питания
// (raw string, как Owner — INV-09/HW-06): домен хранит их как факт, не резолвит.
type PowerTopology struct {
	PowerSource string // идентификатор линии/ввода питания, к которой подключена стойка
	Generator   string // идентификатор резервного дизель-генератора (опционально)
}

// Rack — агрегат стойки внутри модуля: первоклассная локация-корень с CRUD (LOC-01/D-04),
// нижний уровень иерархии DC→Module→Rack. Независимый агрегат (D-04). Несёт ссылку на
// родительский модуль ТОЛЬКО по внутреннему ModuleID (LOC-02/D-06) — не вложенный
// Module-объект. Несёт топологические атрибуты питания как VO (LOC-04). Привязка Host→Rack
// (RackID + позиция) живёт на стороне Host (06-02), НЕ здесь — back-reference запрещён (D-06).
// Встраивает aggregateBase: каждая CRUD-операция эмитит ровно одно событие (D-13/EVT-01).
type Rack struct {
	aggregateBase
	id       RackID        // внутренний постоянный идентификатор идентичности (INV-03)
	moduleID ModuleID      // привязка к родительскому модулю по внутреннему ID (LOC-02/D-06)
	name     string        // имя/метка стойки (обязательное, LOC-01)
	power    PowerTopology // топологические атрибуты питания (LOC-04)
}

// NewRack — фабрика стойки (рецепт «NewX держит инварианты + record»): guard на zero
// ModuleID (привязка к модулю обязательна — нет «висячей» стойки без родителя,
// LOC-02/D-06/T-06-09), требует непустое name (LOC-01/V5), генерирует новый RackID
// (INV-03), эмитит ровно один RackCreated (EVT-01). Топология питания опциональна.
func NewRack(moduleID ModuleID, name string, power PowerTopology) (*Rack, error) {
	if moduleID.IsZero() {
		return nil, ErrInvalidLocation // нет стойки без родительского модуля (LOC-02/D-06)
	}
	if name == "" {
		return nil, ErrInvalidLocation // LOC-01: стойка без имени не существует
	}

	r := &Rack{
		id:       NewRackID(),
		moduleID: moduleID,
		name:     name,
		power:    power,
	}
	r.record(RackCreated{RackID: r.id, ModuleID: moduleID, Name: name, Power: power})
	return r, nil
}

// ID возвращает идентификатор стойки.
func (r *Rack) ID() RackID { return r.id }

// ModuleID возвращает привязку к родительскому модулю (иерархия по ID, LOC-02/D-06).
func (r *Rack) ModuleID() ModuleID { return r.moduleID }

// Name возвращает имя/метку стойки.
func (r *Rack) Name() string { return r.name }

// Power возвращает топологические атрибуты питания стойки (LOC-04).
func (r *Rack) Power() PowerTopology { return r.power }

// Update меняет редактируемые атрибуты стойки, включая топологию питания (CRUD-update,
// LOC-01/04) и эмитит ровно один RackUpdated. Родительский ModuleID неизменяем.
func (r *Rack) Update(name string, power PowerTopology) error {
	if name == "" {
		return ErrInvalidLocation // имя остаётся обязательным (LOC-01)
	}
	r.name = name
	r.power = power
	r.record(RackUpdated{RackID: r.id, Name: name, Power: power})
	return nil
}

// Delete — удаление стойки: эмитит ровно один RackDeleted. Инвариант непустоты
// (нет размещённых хостов) enforced на уровне usecase.
func (r *Rack) Delete() {
	r.record(RackDeleted{RackID: r.id, ModuleID: r.moduleID, Name: r.name})
}

// --- Доменные события Rack (D-13: голый семантический факт). EntityID = строковый RackID. ---

// RackCreated — факт создания стойки (EVT-01): несёт привязку к модулю и топологию питания.
type RackCreated struct {
	RackID   RackID
	ModuleID ModuleID
	Name     string
	Power    PowerTopology
}

func (e RackCreated) EventType() string { return "RackCreated" }
func (e RackCreated) EntityID() string  { return e.RackID.String() }

// RackUpdated — факт обновления атрибутов/топологии стойки (LOC-01/04).
type RackUpdated struct {
	RackID RackID
	Name   string
	Power  PowerTopology
}

func (e RackUpdated) EventType() string { return "RackUpdated" }
func (e RackUpdated) EntityID() string  { return e.RackID.String() }

// RackDeleted — факт удаления стойки: несёт привязку к модулю для аудит-следа.
type RackDeleted struct {
	RackID   RackID
	ModuleID ModuleID
	Name     string
}

func (e RackDeleted) EventType() string { return "RackDeleted" }
func (e RackDeleted) EntityID() string  { return e.RackID.String() }
