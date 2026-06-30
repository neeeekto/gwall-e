package domain

// Project — агрегат проекта: обязательный родитель Host (привязка INV-02) и носитель
// идентичности проекта. Встраивает aggregateBase: каждая операция-переход записывает РОВНО
// одно семантическое событие через record() (D-13/EVT-01), version растёт в одной точке
// (Pitfall 3). Owner — непрозрачный внешний string-ID группы-владельца (INV-09/D-05): домен
// НЕ резолвит и НЕ доверяет его как identity — храним как факт (T-06-10). Ссылок на другие
// агрегаты Project не держит: список своих хостов живёт вне агрегата (D-06).
type Project struct {
	aggregateBase
	id          ProjectID // внутренний постоянный идентификатор идентичности (INV-03)
	name        string    // человекочитаемое имя проекта (обязательное, INV-01)
	description string    // свободное описание (опционально)
	owner       string    // непрозрачный внешний ID группы-владельца (INV-09/D-05) — raw string
}

// NewProject — фабрика проекта, держащая инварианты идентичности (рецепт «NewX держит
// инварианты + record»). Требует непустое name (INV-01/V5), хранит owner как raw string без
// резолва (INV-09), генерирует новый ProjectID (INV-03), эмитит ровно один ProjectCreated
// (EVT-01). description и owner необязательны: владелец может назначаться позже (ChangeOwner),
// описание — свободный текст.
func NewProject(name, description, owner string) (*Project, error) {
	if name == "" {
		return nil, ErrInvalidProject // INV-01: проект без имени не существует
	}

	p := &Project{
		id:          NewProjectID(), // система генерирует идентичность, не вводят снаружи (INV-03)
		name:        name,
		description: description,
		owner:       owner,
	}
	p.record(ProjectCreated{
		ProjectID:   p.id,
		Name:        name,
		Description: description,
		Owner:       owner,
	})
	return p, nil
}

// ID возвращает идентификатор проекта.
func (p *Project) ID() ProjectID { return p.id }

// Name возвращает имя проекта.
func (p *Project) Name() string { return p.name }

// Description возвращает описание проекта.
func (p *Project) Description() string { return p.description }

// Owner возвращает непрозрачный внешний ID группы-владельца (INV-09) — raw string, домен
// не резолвит его как identity.
func (p *Project) Owner() string { return p.owner }

// Rename меняет имя проекта (INV-01) и эмитит ровно один ProjectRenamed. Пустое новое имя
// отклоняется (инвариант обязательности имени сохраняется).
func (p *Project) Rename(name string) error {
	if name == "" {
		return ErrInvalidProject // имя остаётся обязательным (INV-01)
	}
	p.name = name
	p.record(ProjectRenamed{ProjectID: p.id, Name: name})
	return nil
}

// ChangeOwner меняет владельца проекта (INV-09): owner — непрозрачный внешний string-ID,
// домен принимает его как факт без резолва (T-06-10). Эмитит ровно один ProjectOwnerChanged.
func (p *Project) ChangeOwner(owner string) error {
	from := p.owner
	p.owner = owner
	p.record(ProjectOwnerChanged{ProjectID: p.id, FromOwner: from, ToOwner: owner})
	return nil
}

// Delete — удаление проекта: эмитит ровно один ProjectDeleted. Инвариант delete-only-if-empty
// (нет привязанных хостов, INV-10) enforced на уровне usecase через query-порт HostsInProject
// (06-05) — агрегатный Delete лишь эмитит факт, аналогично Host.Delete (D-09).
func (p *Project) Delete() {
	p.record(ProjectDeleted{ProjectID: p.id, Name: p.name, Owner: p.owner})
}

// --- Доменные события Project (D-13: голый семантический факт, минимально-достаточный
// payload, не ProjectUpdated-дамп). Каждое реализует DomainEvent (EventType/EntityID).
// EntityID = строковый ProjectID — будущий Kafka-ключ (Phase 8). ---

// ProjectCreated — факт создания проекта (EVT-01): идентичность, имя, описание, владелец.
type ProjectCreated struct {
	ProjectID   ProjectID
	Name        string
	Description string
	Owner       string
}

func (e ProjectCreated) EventType() string { return "ProjectCreated" }
func (e ProjectCreated) EntityID() string  { return e.ProjectID.String() }

// ProjectRenamed — факт переименования проекта (INV-01).
type ProjectRenamed struct {
	ProjectID ProjectID
	Name      string
}

func (e ProjectRenamed) EventType() string { return "ProjectRenamed" }
func (e ProjectRenamed) EntityID() string  { return e.ProjectID.String() }

// ProjectOwnerChanged — факт смены владельца проекта (INV-09): несёт прежнего и нового
// owner как непрозрачные string-ID.
type ProjectOwnerChanged struct {
	ProjectID ProjectID
	FromOwner string
	ToOwner   string
}

func (e ProjectOwnerChanged) EventType() string { return "ProjectOwnerChanged" }
func (e ProjectOwnerChanged) EntityID() string  { return e.ProjectID.String() }

// ProjectDeleted — факт удаления проекта (INV-10): снапшот имени/владельца для аудит-следа.
type ProjectDeleted struct {
	ProjectID ProjectID
	Name      string
	Owner     string
}

func (e ProjectDeleted) EventType() string { return "ProjectDeleted" }
func (e ProjectDeleted) EntityID() string  { return e.ProjectID.String() }
