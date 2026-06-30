package domain

import (
	"context"
	"time"
)

// Доменные порты Inventory (D-02): интерфейсы, объявленные в domain; реальные реализации
// живут в repositories (Mongo) и infra. Usecase зависит ТОЛЬКО от этих портов — Phase 7
// свапает Mongo-impl без правок usecase. Везде ctx — первый аргумент (architecture.md).

// UnitOfWork очерчивает транзакционную границу (architecture.md §UnitOfWork): Do кладёт
// транзакцию в ctx и выполняет fn; агрегат и его outbox-события пишутся в одной tx
// (запрет dual-write). Фейк в тестах просто зовёт fn.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// Outbox — transactional outbox: Append пишет envelope-события в той же tx, что и агрегат
// (architecture.md §события); отдельный relay публикует их позже (at-least-once).
type Outbox interface {
	Append(ctx context.Context, events []EventEnvelope) error
}

// HostRepository — порт хранилища агрегата Host. Delete — физическое удаление (INV-07/D-09),
// НЕ state=deleted (история живёт на событиях, не на soft-delete-флаге).
type HostRepository interface {
	Save(ctx context.Context, h *Host) error
	Load(ctx context.Context, id HostID) (*Host, error)
	Delete(ctx context.Context, id HostID) error
}

// ProjectRepository — порт хранилища агрегата Project.
type ProjectRepository interface {
	Save(ctx context.Context, p *Project) error
	Load(ctx context.Context, id ProjectID) (*Project, error)
	Delete(ctx context.Context, id ProjectID) error
}

// DCRepository — порт хранилища агрегата DC (дата-центр, LOC-01 CRUD).
type DCRepository interface {
	Save(ctx context.Context, dc *DC) error
	Load(ctx context.Context, id DCID) (*DC, error)
	Delete(ctx context.Context, id DCID) error
}

// ModuleRepository — порт хранилища агрегата Module (зал, LOC-01 CRUD).
type ModuleRepository interface {
	Save(ctx context.Context, m *Module) error
	Load(ctx context.Context, id ModuleID) (*Module, error)
	Delete(ctx context.Context, id ModuleID) error
}

// RackRepository — порт хранилища агрегата Rack (стойка, LOC-01 CRUD).
type RackRepository interface {
	Save(ctx context.Context, r *Rack) error
	Load(ctx context.Context, id RackID) (*Rack, error)
	Delete(ctx context.Context, id RackID) error
}

// ActiveHostByFQDN — query-порт уникальности FQDN среди active-хостов (D-11/INV-10):
// возвращает занявший FQDN хост и флаг found. Реализация Phase 7 — partial unique index.
type ActiveHostByFQDN interface {
	ActiveHostByFQDN(ctx context.Context, fqdn string) (existing HostID, found bool, err error)
}

// HostsInProject — query-порт количества хостов в проекте: для delete-only-if-empty (INV-10).
type HostsInProject interface {
	HostsInProject(ctx context.Context, id ProjectID) (int, error)
}

// MatchAdvisor — порт-хук советочного матчинга при ре-идентификации (D-12/INV-08): по FQDN
// предлагает кандидатов-хостов. В 06-04 — no-op заглушка (пустой слайс). Сигнатура без
// HostHardware, чтобы ports.go компилировался в Wave 1 без forward-зависимости от 06-02;
// hw-аргумент добавится с advisory-движком (SEED-001).
type MatchAdvisor interface {
	Candidates(ctx context.Context, fqdn string) ([]HostID, error)
}

// Clock — порт времени для детерминизма envelope-тестов (Pitfall 6); фейк в 06-04.
type Clock interface {
	Now() time.Time
}

// IDGenerator — порт генерации id событий для детерминизма envelope-тестов (Pitfall 6);
// фейк в 06-04. NewID отдаёт строковый eventId для EventEnvelope.
type IDGenerator interface {
	NewID() string
}
