# Phase 6: Доменная модель Inventory - Pattern Map

**Mapped:** 2026-06-30
**Files analyzed:** ~24 new files (domain, usecase, mocks, fakes, tests, glossary)
**Analogs found:** 24 / 24 (all mapped — knowledge/ canon recipes + example-пакет + style/testing canon; no file left without a source)

> **Greenfield note.** `services/inventory/internal/{domain,usecase,api,cron,app}` are genuinely
> empty directories; `query` and `repositories` do not exist yet. The only compiled Go in
> `internal/` is the throwaway `internal/example/` package (mockery v3 + Ginkgo smoke). Therefore
> the **primary pattern source** for new domain/usecase code is the canon recipe set in
> `knowledge/` (architecture.md / patterns.md / style.md / testing.md), and the **primary form
> source** for tests + mocks is `internal/example/` (a real, compiling mockery-v3 + Ginkgo
> reference). Every new file below maps to a concrete excerpt from one of these. There are no
> "no analog" files — see the table at the bottom for the one structural caveat.

> **Path correction for the planner.** The skeleton dir is `services/inventory/internal/usecase`
> (singular), not `usecases` as RESEARCH.md §"Recommended Project Structure" wrote. Mirror the
> existing dir name unless the planner deliberately renames it. `query`/`repositories` dirs must
> be created if used.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/domain/id.go` | model (value-object) | transform (pure) | `internal/example/provisioner.go` (`ExampleID` typed-id) + style.md §"Типизированные ID" | role-match (form differs: struct-over-uuid vs `type X string`) |
| `internal/domain/aggregate.go` | model (aggregate base) | event-driven | patterns.md Рецепт 3 (`PullEvents`) + architecture.md §"Доменные события" | recipe-exact |
| `internal/domain/events.go` | model (domain events + envelope) | event-driven | architecture.md §"Доменные события" + RESEARCH Pattern 4 (envelope) | recipe-exact |
| `internal/domain/errors.go` | model (sentinel errors) | transform | `internal/example/provisioner.go` (`ErrExampleProvisionFailed`) + style.md §"Sentinel vs обёрнутые" | exact |
| `internal/domain/host.go` | model (aggregate root) | event-driven (lifecycle SM) | patterns.md Рецепт 3 (factory + record) + RESEARCH §"Lifecycle state-machine" | recipe-exact |
| `internal/domain/hardware.go` | model (immutable VO) | transform | RESEARCH Pattern 3 (HostHardware) — D-07 (no repo analog; greenfield VO) | recipe-only |
| `internal/domain/project.go` | model (aggregate root) | event-driven | patterns.md Рецепт 3 (factory + PullEvents) | recipe-exact |
| `internal/domain/dc.go` / `module.go` / `rack.go` | model (aggregate roots) | CRUD / event-driven | patterns.md Рецепт 3 (per-aggregate factory) — mirrors host.go form | recipe-exact |
| `internal/domain/ports.go` | port (interfaces) | request-response | architecture.md §"UnitOfWork" + Рецепт 1/4 + `example.ExampleProvisioner` (port shape) | recipe-exact |
| `internal/usecase/register_host.go` (+ decommission/delete/reassign/relocate/change_hardware) | usecase (interactor) | request-response (write) | patterns.md Рецепт 1 (`Execute`) + architecture.md §"Write-side" + RESEARCH Pattern 5 | recipe-exact |
| `internal/usecase/create_project.go` / `delete_project.go` / loc CRUD | usecase (interactor) | request-response (write) | patterns.md Рецепт 1 + RESEARCH §"Project delete-only-if-empty" | recipe-exact |
| `internal/usecase/envelope.go` | utility (enrichment helper) | transform | RESEARCH Pattern 4 (enrich) — D-14/D-15 boundary | recipe-only |
| `internal/usecase/fakes/*.go` (fakeUoW/fakeOutbox/fakeClock/fakeIDGen) | test-double (hand fakes) | event-driven | RESEARCH Pattern 6 (trivial fakes) — D-02 | recipe-only |
| `internal/{domain,usecase}/mocks/*.go` | test-double (generated) | request-response | `internal/example/mocks/ExampleProvisioner.go` (mockery v3 output) | exact |
| `internal/domain/*_test.go` | test (direct unit) | n/a | `internal/example/provisioner_test.go` (Ginkgo suite) + testing.md §"Структура спеков" | exact (suite shape) |
| `internal/usecase/*_test.go` | test (mock + fake) | n/a | `internal/example/provisioner_test.go` (mockery expecter) + testing.md §"Мокинг портов" | exact (mock shape) |
| `.mockery.yaml` (modify) | config | n/a | existing `.mockery.yaml` (example entry) | exact |
| `knowledge/glossary.md` | doc (glossary) | n/a | authoring.md (MUST/SHOULD/WON'T) + RESEARCH §"DOC-07 structure" | recipe-only |

## Pattern Assignments

### `internal/domain/id.go` (typed ID-VO, transform)

**Analog:** `services/inventory/internal/example/provisioner.go` (typed-id form) + `knowledge/style.md` §"Типизированные ID".

**Existing typed-id form** (`internal/example/provisioner.go:11-13`) — the repo's only real precedent; note it uses `type X string`, NOT struct-over-uuid:
```go
// ExampleID — типизированный идентификатор примера-ресурса (style.md: типизированные ID
// вместо «голой» строки). Не доменный тип; служит лишь для smoke-демонстрации.
type ExampleID string
```

**Canon rule** (`knowledge/style.md:51-55`) — MUST use a typed ID, compiler-checked:
```go
// хорошо: типизированный ID — компилятор ловит подмену
type OrderID string
func GetOrder(id OrderID) (*Order, error)
```

**D-05 deviation (load-bearing):** CONTEXT D-05 explicitly requires "обёртка над `uuid.UUID` с
фабрикой/парсингом" → planner overrides the `type X string` precedent with `struct{ v uuid.UUID }`
(RESEARCH Pattern 1: `NewHostID()`/`ParseHostID()`/`String()`/`IsZero()`). Use `github.com/google/uuid`
(`uuid.New()` v4), already in go.mod (indirect → direct after first import + `go mod tidy`). Wrap
`uuid.Parse` errors with `%w` into a domain sentinel (`ErrInvalidID`). Comments Russian, identifiers
English (style.md:31-36). One ID-VO per aggregate: `HostID`/`ProjectID`/`DCID`/`ModuleID`/`RackID`.

---

### `internal/domain/aggregate.go` (aggregateBase, event-driven)

**Analog:** `knowledge/patterns.md` Рецепт 3 + `knowledge/architecture.md` §"Доменные события".

**Canon: PullEvents + factory rule** (`knowledge/patterns.md:128-143`):
```go
// 1) фабрика держит инварианты
func NewOrder(sku string, qty int) (*Order, error) {
    if qty <= 0 {
        return nil, ErrInvalidQty // sentinel — style.md §ошибки
    }
    o := &Order{id: newOrderID(), sku: sku, qty: qty}
    o.record(OrderRegistered{ID: o.id}) // 2) событие копится в агрегате
    return o, nil
}
// 3) репозиторий перед записью сливает события: outbox.Append(order.PullEvents())
func (o *Order) PullEvents() []DomainEvent { /* отдаёт и очищает буфер */ }
```

**Canon: PullEvents → outbox in same tx** (`knowledge/architecture.md:125-132`):
```go
events := order.PullEvents()          // агрегат отдаёт накопленные доменные события
if err := outbox.Append(ctx, events); err != nil { // тот же ctx = та же tx
    return err
}
```

**Form to build** (RESEARCH Pattern 2, embeds into all 5 aggregates): `aggregateBase{ version int; events []DomainEvent }` with `record(e)` (version++ AND append — single point, Pitfall 3) and `PullEvents()` (return-and-clear `events=nil`, Pitfall 5) and `Version() int`. NOT in `pkg/` — inventory-specific (RESEARCH A3 / MEMORY shared-code-in-pkg applies only to truly-generic code).

---

### `internal/domain/events.go` (domain events + envelope, event-driven)

**Analog:** `knowledge/architecture.md` §"Доменные события" + RESEARCH Pattern 4.

**Bare semantic facts (D-14):** `DomainEvent` interface = `EventType() string` + `EntityID() string`. One semantic event per business op (D-13, anti-pattern: `HostUpdated`-dump — Pitfall 5). Names = ubiquitous language fixed in glossary: `HostRegistered`, `HostHardwareChanged`, `HostReassigned`, `HostRelocated`, `HostDecommissioned`, `HostDeleted`; `ProjectCreated`/…; `DCCreated`/`ModuleCreated`/`RackCreated`/… (RESEARCH:564-572).

**EventEnvelope (D-15)** carries `EventID`, `EntityID`, `EventType`, `Version`, `OccurredAt`, `Actor{ID, Source}`, `Payload DomainEvent` — from day one. `Actor.Source ∈ {human|api|integration|system}`. Envelope-meta is NOT set in the aggregate (D-14) — only the struct shape lives in domain; enrichment happens in usecase (see envelope.go).

---

### `internal/domain/errors.go` (sentinel errors, transform)

**Analog:** `internal/example/provisioner.go:18` + `knowledge/style.md:66-84`.

**Existing sentinel form** (`internal/example/provisioner.go:16-18`):
```go
// ErrExampleProvisionFailed — sentinel-ошибка примера (style.md: предсказуемые ошибки —
// sentinel, сравниваются через errors.Is).
var ErrExampleProvisionFailed = errors.New("example provision failed")
```

**Canon** (`knowledge/style.md:77-84`) — sentinel + `%w`, never `%v`:
```go
var ErrOrderNotFound = errors.New("order not found")
// хорошо: %w сохраняет sentinel в цепочке — errors.Is работает
return fmt.Errorf("get order %s: %w", id, ErrOrderNotFound)
```

**Domain sentinels/typed conflicts to declare:** `ErrInvalidID`, `ErrInvalidTransition`/`ErrAlreadyDecommissioned`, and **typed conflicts** `ErrFQDNConflict{FQDN, ExistingID, Candidates}` + `ErrProjectNotEmpty{ProjectID, HostCount}` (D-11/Pitfall 7 — NOT raw DB E11000; comparable via `errors.As`). Error **strings English**, comments Russian (style.md:38-39).

---

### `internal/domain/host.go` (aggregate root, lifecycle state-machine)

**Analog:** `knowledge/patterns.md` Рецепт 3 (factory + record) + RESEARCH §"Lifecycle state-machine".

**Factory-holds-invariants** — same shape as patterns.md:128-143 above. `NewHost(projectID, fqdn, hw, lifecycle)` requires non-zero `ProjectID` (INV-02), generates `HostID` (INV-03), records `HostRegistered`.

**State-machine + Delete form** (RESEARCH:519-543 — `[CITED: CONTEXT.md D-08/D-10]`):
```go
type lifecycleState int
const (
    stateShadow lifecycleState = iota
    stateRegistered
    stateDecommissioned // терминально (D-10: нет воскрешения)
)
func (h *Host) Decommission(reason string) error {
    if h.state == stateDecommissioned { return ErrAlreadyDecommissioned }
    h.state = stateDecommissioned
    h.record(HostDecommissioned{ID: h.id, Reason: reason}) // version++ внутри record
    return nil
}
func (h *Host) Delete() { // hard-delete из ЛЮБОГО состояния (D-09): эмитит факт, usecase зовёт repo.Delete
    h.record(HostDeleted{ID: h.id, Snapshot: h.snapshot()})
}
```

**Critical (Pitfall 2 / D-09):** `deleted` is NOT a 4th enum member — exactly 3 lifecycle states; `Delete()` is a separate method + physical `repo.Delete`. Methods: `Reassign(newProjectID)`→`HostReassigned`, `Relocate(rackID,pos)`→`HostRelocated`, `ChangeHardware(newHW)`→`HostHardwareChanged`. Inter-aggregate refs by internal ID only (D-06).

---

### `internal/domain/hardware.go` (immutable HostHardware VO, transform)

**Analog:** RESEARCH Pattern 3 (D-07). **No codebase analog — greenfield VO**; build from D-07/HW-01…06.

**Form** (RESEARCH:300-326): single immutable `HostHardware` with private fields, all sub-components structured: `Motherboard`, `[]RAMModule`, `[]CPU`, `[]Drive`, `[]NIC` (structured: model/speedGbE/`macs []string`, NOT flat `MACs[]` — HW-03), `[]PSU`, `[]StorageController`, `[]GPU`, `Chassis`, `ipmiMAC string`. Constructor `NewHostHardware(spec) (HostHardware, error)`.

**Critical (Pitfall 2):** Go slices are reference types — defensive-copy slices in constructor AND any slice getter (`append([]NIC(nil), in...)`), else the "immutable" VO leaks. Change = build a new VO whole → one `HostHardwareChanged` event (NOT per-component mutation/events). All external IDs (serial/inv/MAC/model) = raw `string` (HW-06).

---

### `internal/domain/project.go` + `dc.go`/`module.go`/`rack.go` (aggregate roots)

**Analog:** `knowledge/patterns.md` Рецепт 3 — mirror `host.go` form (factory + record + PullEvents via aggregateBase).

Project: `NewProject(name, desc, owner)` (`Owner` raw string, INV-09) → `ProjectCreated`; ops `Rename`/`ChangeOwner`/`Delete` (set — Claude's Discretion, D-13). Locations are **3 independent aggregates** (D-04), hierarchy by ID ref: `Module` holds `DCID`, `Rack` holds `ModuleID`; `Rack` carries topology attrs (LOC-04). NOT one Location tree.

---

### `internal/domain/ports.go` (port interfaces, request-response)

**Analog:** `knowledge/architecture.md` §"UnitOfWork" + Рецепты 1/4 + `example.ExampleProvisioner` (port shape: `ctx` first, interface in domain).

**Port shape precedent** (`internal/example/provisioner.go:24-28`) — interface declared with ctx-first method, real implementations elsewhere:
```go
type ExampleProvisioner interface {
    Provision(ctx context.Context, id ExampleID, name string) error
}
```

**Canon (UnitOfWork)** (`knowledge/architecture.md:94-100`): MUST draw tx boundary with port `UnitOfWork.Do(ctx, fn)` declared in `domain`, impl in `repositories`; aggregate + outbox in one tx.

**Ports to declare** (RESEARCH Pattern 6 / §"Architecture Diagram"): `UnitOfWork{ Do(ctx, fn) }`, `Outbox{ Append(ctx, []EventEnvelope) }`, `HostRepository{ Save; Delete; (Load) }`, `ProjectRepository`, `*LocationRepository`, query-port `ActiveHostByFQDN(ctx, fqdn) (HostID, bool, error)` (D-11), `MatchAdvisor{ Candidates(...) }` no-op hook (D-12/INV-08), `Clock{ Now() }` + `IDGenerator{ NewID() }` for deterministic envelope tests (Pitfall 6). These ports become the **first real mockery targets** (D-03) — see `.mockery.yaml`.

---

### `internal/usecase/register_host.go` (+ all write interactors) (interactor, request-response)

**Analog:** `knowledge/patterns.md` Рецепт 1 + `knowledge/architecture.md` §"Write-side" + RESEARCH Pattern 5.

**Canon interactor** (`knowledge/patterns.md:46-64`) — 1 use case = 1 struct + `Execute`, deps = ports only, write inside `uow.Do`:
```go
type RegisterOrderUseCase struct {
    orders OrderRepository // порт
    uow    UnitOfWork      // порт транзакционной границы
}
func (uc *RegisterOrderUseCase) Execute(ctx context.Context, in RegisterOrderInput) (RegisterOrderOutput, error) {
    order, err := NewOrder(in.SKU, in.Qty) // фабрика держит инварианты
    if err != nil { return RegisterOrderOutput{}, err }
    if err := uc.uow.Do(ctx, func(ctx context.Context) error {
        return uc.orders.Save(ctx, order)
    }); err != nil {
        return RegisterOrderOutput{}, fmt.Errorf("register order: %w", err)
    }
    return RegisterOrderOutput{ID: order.ID()}, nil
}
```

**Inventory extension** (RESEARCH:381-404) — inside `uow.Do(fn)`: FQDN-check via query-port (typed `ErrFQDNConflict`, Pitfall 7) → `repo.Save` → `enrich(host.PullEvents(), actor, clock, idgen)` → `outbox.Append`. **Critical (Pitfall 8 / D-02):** keep `outbox.Append` INSIDE `uow.Do(fn)` even on fakes — Phase 7 = pure swap of port impls, `Execute` untouched. `actor` arrives as `Execute` param (never enters aggregate, D-14).

---

### `internal/usecase/delete_project.go` (interactor, delete-if-empty invariant)

**Analog:** RESEARCH:546-561 (INV-10) + Рецепт 1.
```go
func (uc *DeleteProjectUseCase) Execute(ctx context.Context, id domain.ProjectID, actor Actor) error {
    return uc.uow.Do(ctx, func(ctx context.Context) error {
        n, err := uc.hostCount.HostsInProject(ctx, id) // query-порт
        if err != nil { return err }
        if n > 0 { return domain.ErrProjectNotEmpty{ProjectID: id, HostCount: n} }
        proj, err := uc.projects.Load(ctx, id)
        if err != nil { return err }
        proj.Delete()
        if err := uc.projects.Delete(ctx, id); err != nil { return err }
        return uc.outbox.Append(ctx, enrich(proj.PullEvents(), actor, uc.clock, uc.idgen))
    })
}
```

---

### `internal/usecase/envelope.go` (enrichment helper, transform)

**Analog:** RESEARCH Pattern 4 (D-14/D-15). No codebase analog — boundary helper.

`enrich(events []domain.DomainEvent, actor Actor, clock domain.Clock, idgen domain.IDGenerator) []domain.EventEnvelope` — loops bare facts, sets `EventID` (idgen port), `OccurredAt` (clock port), `Actor`, copies `EventType`/`EntityID`/version. Lives at the usecase boundary BETWEEN `PullEvents()` and `outbox.Append` (RESEARCH:344-359). Ports give deterministic tests (Pitfall 6).

---

### `internal/usecase/fakes/*.go` (hand fakes, test-double)

**Analog:** RESEARCH Pattern 6 (D-02). Trivial hand fakes (allowed exception to "no hand fakes" since behaviour is trivial/deterministic — RESEARCH A4):
```go
type fakeUoW struct{}
func (fakeUoW) Do(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }
type fakeOutbox struct{ appended []EventEnvelope }
func (o *fakeOutbox) Append(_ context.Context, evs []EventEnvelope) error { o.appended = append(o.appended, evs...); return nil }
```
Also `fakeClock` (fixed time) + `fakeIDGen` (sequence) for deterministic envelope assertions. **Split (D-02 vs D-03):** uow/outbox/clock/idgen = hand fakes (need to run `fn` and collect events for outcome assertions); repo/uniq/advisor = mockery mocks (expectation/call-order checks).

---

### `internal/{domain,usecase}/mocks/*.go` (generated mocks, test-double)

**Analog:** `services/inventory/internal/example/mocks/ExampleProvisioner.go` — the real mockery-v3 output shape (DO NOT EDIT header, `NewMockX(t)` auto-Cleanup, expecter `EXPECT()`, testify import). Generated by `make generate-mocks`; never hand-written. New targets: the domain ports added to `.mockery.yaml`.

---

### `internal/domain/*_test.go` (direct unit specs)

**Analog:** `services/inventory/internal/example/provisioner_test.go` + `knowledge/testing.md`.

**Suite bootstrap** (`internal/example/provisioner_test.go:18-21` — real, compiles):
```go
func TestExampleSuite(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Example Package Suite")
}
```

**Imports** (`provisioner_test.go:13-15`): dot-import ginkgo/gomega, regular import for testify:
```go
. "github.com/onsi/ginkgo/v2"
. "github.com/onsi/gomega"
"github.com/stretchr/testify/mock" // regular qualified import (NOT dot)
```

Domain/aggregate/VO/ID/lifecycle tested by **direct** unit specs, no mocks (D-03). Use `DescribeTable` for lifecycle transitions + hardware structure (testing.md:56-58). Comments English (style.md:38). Suite name per file. Assert via Gomega `MatchError`/`Equal`/`BeAssignableToTypeOf`.

---

### `internal/usecase/*_test.go` (mock + fake specs)

**Analog:** `internal/example/provisioner_test.go:23-57` (mockery expecter usage) + `knowledge/testing.md` §"Мокинг портов".

**Mock expecter pattern** (`provisioner_test.go:32-55`):
```go
m = mocks.NewMockExampleProvisioner(GinkgoT()) // auto-Cleanup asserts expectations
m.EXPECT().Provision(mock.Anything, mock.Anything, mock.Anything).Return(nil)
// ...
Expect(err).To(MatchError(example.ErrExampleProvisionFailed))
```

Usecase specs: mockery mocks for repo/uniq/advisor (`GinkgoT()` + `EXPECT()`), hand fakes for uow/outbox/clock/idgen, assert on `fakeOutbox.appended` (exactly one semantic event per op — EVT-01) and on typed conflicts via `MatchError`/`errors.As`.

---

### `.mockery.yaml` (modify, config)

**Analog:** existing `.mockery.yaml` (repo root) — extend the `packages:` map with the new domain ports. Keep v3 invariants (header comment lines 1-4): `template: testify`, `dir: "{{.InterfaceDir}}/mocks"`, `.SrcPackageName` (NOT v2 `.PackageName`). Add entries for `github.com/gwall-e/services/inventory/internal/domain` listing `HostRepository`/`ProjectRepository`/`Outbox`/`UnitOfWork`/`ActiveHostByFQDN`/`MatchAdvisor`/location repos. Example entry to mirror (`.mockery.yaml:14-17`):
```yaml
github.com/gwall-e/services/inventory/internal/example:
    interfaces:
      ExampleProvisioner: {}
```
The throwaway `example` package may be removed once real ports are mockery targets (RESEARCH A6, planner's call).

---

### `knowledge/glossary.md` (doc, DOC-07)

**Analog:** `knowledge/authoring.md` (MUST/SHOULD/WON'T standard) + RESEARCH:574-590 structure. No prior glossary file — greenfield doc. Fixes ubiquitous language: Project/Host/Owner/Module/Rack/DC/Connection (forward-term), identity (INV-03), **`decommission ≠ delete`**, the boundary "факт существования (Inventory owns) vs динамическое состояние (State/Health)", and semantic event names (D-13). Glossary is the one allowed exception to no-phantom (may name forward terms). Index it in `knowledge/README.md` (status "exists").

## Shared Patterns

### Error handling (sentinel + `%w`, typed conflicts)
**Source:** `knowledge/style.md:66-84`; precedent `internal/example/provisioner.go:16-18`.
**Apply to:** all domain + usecase files. Predictable errors as `var Err… = errors.New(...)`; wrap with `%w`; typed conflicts (`ErrFQDNConflict`, `ErrProjectNotEmpty`) comparable via `errors.As` — never leak raw DB errors (Pitfall 7). Error strings English, comments Russian.

### Transactional orchestration (uow.Do + PullEvents → outbox)
**Source:** `knowledge/architecture.md:94-132` + `knowledge/patterns.md:46-64,145-181`.
**Apply to:** every write usecase. Aggregate + outbox in one `uow.Do(fn)`; `outbox.Append` stays inside `fn` (Pitfall 8); Phase 7 swaps port impls only (D-02). No CQRS bus / TxManager / `pkg/mediatr` (architecture.md:145-155 — depguard `no-cqrs-bus` bites today).

### Ginkgo suite + mockery v3
**Source:** `internal/example/provisioner_test.go` (real) + `internal/example/mocks/ExampleProvisioner.go` (real) + `knowledge/testing.md:13-143`.
**Apply to:** all `*_test.go` and all generated mocks. `RegisterFailHandler(Fail)`+`RunSpecs`; dot-import ginkgo/gomega, regular import testify; `NewMockX(GinkgoT())` auto-Cleanup; `Describe`→`Context`→`It`; `DescribeTable` for tables; assert via Gomega.

### Typed identity (compiler-checked IDs)
**Source:** `knowledge/style.md:51-64` + `internal/example/provisioner.go:11-13`; D-05 override to struct-over-uuid.
**Apply to:** all 5 aggregate IDs. One typed ID-VO per aggregate; external IDs (Owner/INV/serial/MAC) raw `string` (HW-06).

### Comment/identifier language
**Source:** `knowledge/style.md:25-48`.
**Apply to:** all Go files. Russian domain comments, English identifiers; English comments in `*_test.go`.

## No Analog Found

No file is left without a mapped source. The structural caveat below is for planner awareness, not a gap:

| File / concern | Role | Data Flow | Note |
|------|------|-----------|------|
| `internal/domain/hardware.go` | immutable VO | transform | No codebase VO precedent (greenfield) — mapped to RESEARCH Pattern 3 / D-07; defensive-copy is the load-bearing detail (Pitfall 2). |
| `internal/usecase/envelope.go` | helper | transform | No precedent — boundary enrichment per D-14/D-15 (RESEARCH Pattern 4). |
| `query` / `repositories` dirs | — | — | Do not exist on disk yet; Phase 6 keeps them as skeleton/fakes (SVC-01). Query-port *interface* lives in `domain/ports.go`; real Mongo impl is Phase 7. |

## Metadata

**Analog search scope:** `services/inventory/internal/` (example, kafka, skeleton dirs), `services/audit/`, `pkg/` (http, kafka, mongoconn), `knowledge/*.md`, `.mockery.yaml`, `Makefile`, `services/inventory/go.mod`.
**Files scanned:** ~20 (all Go under services + pkg + 5 knowledge canon docs + mockery/Makefile/go.mod).
**Key finding:** inventory `internal/{domain,usecase,api,cron,app}` are empty; `query`/`repositories` absent. Pattern source of truth = knowledge/ canon recipes (form) + `internal/example/` (test+mock shape) + `google/uuid` (already in go.mod). `pkg/` holds nothing reusable for domain aggregates (http/kafka/mongoconn are infra-level). `services/audit/` is a stub `main.go` only — no domain code to mirror.
**Pattern extraction date:** 2026-06-30
```
