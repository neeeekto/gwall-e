package domain

// Host — центральный агрегат парка (носитель идентичности хоста, источник большинства
// семантических событий backbone). Встраивает aggregateBase: каждая операция-переход
// записывает РОВНО одно семантическое событие через record() (D-13/EVT-01), version
// растёт в одной точке (Pitfall 3). Ссылки на другие агрегаты — только по внутреннему ID
// (D-06). Состав железа — отдельный immutable VO HostHardware (D-07).
type Host struct {
	aggregateBase
	id        HostID         // внутренний постоянный идентификатор идентичности (INV-03)
	projectID ProjectID      // обязательная привязка к проекту (INV-02)
	fqdn      string         // полностью квалифицированное доменное имя
	hardware  HostHardware   // immutable VO состава железа (HW-01/D-07)
	rackID    RackID         // стойка размещения (LOC-03)
	position  string         // позиция (юнит) в стойке (LOC-03)
	state     lifecycleState // статус жизненного цикла (D-08)
}

// lifecycleState — статус жизненного цикла хоста (D-08). РОВНО три члена: deleted НЕ член
// enum (Pitfall 1/D-09) — удаление это hard-delete через Delete()+repo.Delete, не строка
// state=deleted в живой коллекции.
type lifecycleState int

const (
	stateShadow         lifecycleState = iota // заготовка/обнаружен (ещё не зарегистрирован вручную)
	stateRegistered                           // полная ручная регистрация
	stateDecommissioned                       // списан — терминально (D-10: воскрешения нет)
)

// NewHost — фабрика хоста, держащая инварианты идентичности и привязки (рецепт «NewX держит
// инварианты + record»). Требует непустой projectID (INV-02), генерирует новый HostID
// (INV-03), несёт стойку+позицию (LOC-03), эмитит ровно один HostRegistered (EVT-01).
// initial допускает гибкий вход: хост создаётся в shadow или сразу в registered (D-10);
// любое иное стартовое состояние (в т.ч. decommissioned) отклоняется как недопустимый старт.
func NewHost(projectID ProjectID, fqdn string, hw HostHardware, rackID RackID, position string, initial lifecycleState) (*Host, error) {
	if projectID.IsZero() {
		return nil, ErrMissingProject // INV-02: хост не существует без проекта
	}
	if initial != stateShadow && initial != stateRegistered {
		return nil, ErrInvalidTransition // старт допустим только в shadow/registered (D-10)
	}

	h := &Host{
		id:        NewHostID(), // система генерирует идентичность, не вводят снаружи (INV-03)
		projectID: projectID,
		fqdn:      fqdn,
		hardware:  hw,
		rackID:    rackID,
		position:  position,
		state:     initial,
	}
	h.record(HostRegistered{
		HostID:    h.id,
		ProjectID: projectID,
		FQDN:      fqdn,
		RackID:    rackID,
		Position:  position,
	})
	return h, nil
}

// ID возвращает идентификатор хоста.
func (h *Host) ID() HostID { return h.id }

// ProjectID возвращает текущую привязку к проекту.
func (h *Host) ProjectID() ProjectID { return h.projectID }

// FQDN возвращает доменное имя хоста.
func (h *Host) FQDN() string { return h.fqdn }

// Hardware возвращает immutable VO состава железа (value-копия VO; внутренние слайсы
// защищены defensive-copy в самом VO).
func (h *Host) Hardware() HostHardware { return h.hardware }

// RackID возвращает стойку размещения (LOC-03).
func (h *Host) RackID() RackID { return h.rackID }

// Position возвращает позицию в стойке (LOC-03).
func (h *Host) Position() string { return h.position }

// isDecommissioned сообщает, списан ли хост (терминальное состояние, D-10).
func (h *Host) isDecommissioned() bool { return h.state == stateDecommissioned }

// Reassign переназначает хост в другой проект (INV-05/D-06): меняет привязку и эмитит
// ровно один HostReassigned. Списанный хост не переназначается (терминально).
func (h *Host) Reassign(newProjectID ProjectID) error {
	if h.isDecommissioned() {
		return ErrAlreadyDecommissioned
	}
	if newProjectID.IsZero() {
		return ErrMissingProject // привязка остаётся обязательной (INV-02)
	}
	from := h.projectID
	h.projectID = newProjectID
	h.record(HostReassigned{HostID: h.id, FromProjectID: from, ToProjectID: newProjectID})
	return nil
}

// Relocate перемещает хост в другую стойку/позицию (LOC-03): эмитит ровно один HostRelocated.
// Списанный хост не перемещается (терминально).
func (h *Host) Relocate(rackID RackID, position string) error {
	if h.isDecommissioned() {
		return ErrAlreadyDecommissioned
	}
	h.rackID = rackID
	h.position = position
	h.record(HostRelocated{HostID: h.id, RackID: rackID, Position: position})
	return nil
}

// ChangeHardware заменяет состав железа ЦЕЛИКОМ новым immutable VO (D-07) и эмитит ровно
// одно HostHardwareChanged — НЕ per-компонент события. Списанный хост не меняет железо.
func (h *Host) ChangeHardware(newHW HostHardware) error {
	if h.isDecommissioned() {
		return ErrAlreadyDecommissioned
	}
	h.hardware = newHW
	h.record(HostHardwareChanged{HostID: h.id, HardwareName: newHW.Name()})
	return nil
}

// Decommission списывает хост (INV-06): из shadow или registered. Терминально (D-10) —
// повторный вызов возвращает ErrAlreadyDecommissioned; воскрешения нет. Эмитит ровно один
// HostDecommissioned.
func (h *Host) Decommission(reason string) error {
	if h.isDecommissioned() {
		return ErrAlreadyDecommissioned
	}
	h.state = stateDecommissioned
	h.record(HostDecommissioned{HostID: h.id, Reason: reason})
	return nil
}

// Delete — hard-удаление записи хоста из ЛЮБОГО состояния (INV-07/D-09): эмитит HostDeleted
// с полным snapshot-payload для аудита (история живёт на событиях, не на soft-delete-флаге).
// lifecycleState НЕ получает значение «deleted» (Pitfall 1) — usecase зовёт repo.Delete.
func (h *Host) Delete() {
	h.record(HostDeleted{
		HostID:    h.id,
		ProjectID: h.projectID,
		FQDN:      h.fqdn,
		RackID:    h.rackID,
		Position:  h.position,
	})
}

// --- Доменные события Host (D-13: голый семантический факт, минимально-достаточный payload,
// не HostUpdated-дамп). Каждое реализует DomainEvent (EventType/EntityID). EntityID = строковый
// HostID — будущий Kafka-ключ (Phase 8). ---

// HostRegistered — факт регистрации хоста (EVT-01): идентичность, привязка, локация.
type HostRegistered struct {
	HostID    HostID
	ProjectID ProjectID
	FQDN      string
	RackID    RackID
	Position  string
}

func (e HostRegistered) EventType() string { return "HostRegistered" }
func (e HostRegistered) EntityID() string  { return e.HostID.String() }

// HostReassigned — факт переназначения хоста в другой проект (INV-05/D-06).
type HostReassigned struct {
	HostID        HostID
	FromProjectID ProjectID
	ToProjectID   ProjectID
}

func (e HostReassigned) EventType() string { return "HostReassigned" }
func (e HostReassigned) EntityID() string  { return e.HostID.String() }

// HostRelocated — факт перемещения хоста в другую стойку/позицию (LOC-03).
type HostRelocated struct {
	HostID   HostID
	RackID   RackID
	Position string
}

func (e HostRelocated) EventType() string { return "HostRelocated" }
func (e HostRelocated) EntityID() string  { return e.HostID.String() }

// HostHardwareChanged — факт замены состава железа целиком (D-07): несёт имя нового VO,
// не дамп всех компонентов (D-13).
type HostHardwareChanged struct {
	HostID       HostID
	HardwareName string
}

func (e HostHardwareChanged) EventType() string { return "HostHardwareChanged" }
func (e HostHardwareChanged) EntityID() string  { return e.HostID.String() }

// HostDecommissioned — факт списания хоста (INV-06): терминальный переход + причина.
type HostDecommissioned struct {
	HostID HostID
	Reason string
}

func (e HostDecommissioned) EventType() string { return "HostDecommissioned" }
func (e HostDecommissioned) EntityID() string  { return e.HostID.String() }

// HostDeleted — факт hard-удаления записи хоста (INV-07/D-09): полный snapshot-payload для
// аудит-следа (FQDN освобождается, история сохраняется на событии).
type HostDeleted struct {
	HostID    HostID
	ProjectID ProjectID
	FQDN      string
	RackID    RackID
	Position  string
}

func (e HostDeleted) EventType() string { return "HostDeleted" }
func (e HostDeleted) EntityID() string  { return e.HostID.String() }
