---
phase: 06-inventory
plan: 01
subsystem: domain
tags: [go, ddd, hexagonal, uuid, ginkgo, gomega, mockery, domain-events, ports]

# Dependency graph
requires:
  - phase: 05-bootstrap
    provides: "go.work с inventory как полноправным членом; mockery v3.7.1 pinned + .mockery.yaml v3-каркас (example smoke); google/uuid в go.mod (indirect); Ginkgo/Gomega"
provides:
  - "5 типизированных ID-VO (HostID/ProjectID/DCID/ModuleID/RackID) struct-over-uuid.UUID с factory/Parse/String/IsZero"
  - "aggregateBase (record/PullEvents/Version) — встраиваемая база агрегатов"
  - "DomainEvent интерфейс + EventEnvelope + Actor — форма доменного события"
  - "доменные sentinel-ошибки + типизированные конфликты (ErrFQDNConflict/ErrProjectNotEmpty)"
  - "12 доменных портов (UnitOfWork/Outbox/repositories/query-ports/MatchAdvisor/Clock/IDGenerator)"
  - "12 mockery-моков портов в internal/domain/mocks"
  - "Wave-1 placeholder-агрегаты Host/Project/DC/Module/Rack (наполняются 06-02/06-03)"
affects: [06-02-host, 06-03-locations, 06-04-usecases, 06-05-crud-usecases, 06-06-events, phase-07-mongo, phase-08-protobuf]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Typed ID-VO как struct{ v uuid.UUID } (D-05) — переопределяет прецедент ExampleID type-string"
    - "aggregateBase: version++ в одной точке record() (Pitfall 3); PullEvents отдаёт-и-очищает (Pitfall 5)"
    - "DomainEvent — голый факт; envelope-мета заполняется на границе usecase, не в агрегате (D-14)"
    - "Порты в domain, ctx-first; реализации в repositories (D-02)"
    - "Типизированные конфликты-struct вместо сырого DB E11000 (Pitfall 7/D-11)"

key-files:
  created:
    - services/inventory/internal/domain/id.go
    - services/inventory/internal/domain/errors.go
    - services/inventory/internal/domain/aggregate.go
    - services/inventory/internal/domain/events.go
    - services/inventory/internal/domain/ports.go
    - services/inventory/internal/domain/aggregates_stub.go
    - services/inventory/internal/domain/domain_suite_test.go
    - services/inventory/internal/domain/id_test.go
    - services/inventory/internal/domain/aggregate_test.go
    - services/inventory/internal/domain/mocks/ (12 mocks)
  modified:
    - .mockery.yaml
    - services/inventory/go.mod

key-decisions:
  - "MatchAdvisor.Candidates(ctx, fqdn string) ([]HostID, error) — без HostHardware, чтобы ports.go компилировался в Wave 1 без forward-зависимости от 06-02 (hw добавится с advisory-движком, SEED-001)"
  - "Wave-1 placeholder-агрегаты в aggregates_stub.go (Rule 3) — минимальные struct{aggregateBase; id XID} + ID() геттер; 06-02/06-03 РАСШИРЯЮТ эти типы, а не переобъявляют (иначе redeclaration)"
  - "aggregate_test.go — white-box (package domain): aggregateBase/record/поля неэкспортируемы by design, тест встраивает базу напрямую"

patterns-established:
  - "Typed ID-VO struct-over-uuid: NewXID/ParseXID(%w ErrInvalidID)/String/IsZero на каждый из 5 ID"
  - "DescribeTable-драйвер по пяти ID-типам через idCase с функциями-замыканиями"
  - "Доменные порты + mockery-моки: .mockery.yaml packages-запись → make generate-mocks → internal/domain/mocks"

requirements-completed: [INV-03, EVT-01, EVT-02, INV-08, INV-10, SVC-01]

# Metrics
duration: ~5min
completed: 2026-06-30
---

# Phase 6 Plan 01: Доменное ядро Inventory Summary

**Фундамент домена Inventory: 5 типизированных ID-VO над uuid.UUID, встраиваемый aggregateBase с накоплением событий, форма DomainEvent/EventEnvelope/Actor, доменные sentinel + типизированные конфликты, 12 портов и сгенерированные mockery-моки — go build/vet/test зелёные.**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-06-30T19:05:07Z
- **Completed:** 2026-06-30T19:10:16Z
- **Tasks:** 3
- **Files modified:** 23 (11 рукотворных + 12 сгенерированных моков)

## Accomplishments
- 5 типизированных ID-VO (`HostID`/`ProjectID`/`DCID`/`ModuleID`/`RackID`) как `struct{ v uuid.UUID }` (D-05) с фабрикой v4-random, `Parse` (обёртка `%w ErrInvalidID`), `String`, `IsZero`; компилятор различает типы (INV-03/T-06-01)
- `aggregateBase` встраиваемая база: `record()` инкрементит version в одной точке (Pitfall 3), `PullEvents()` отдаёт-и-очищает буфер (Pitfall 5), `Version()` accessor
- `DomainEvent` интерфейс (голый факт, D-14) + `EventEnvelope` (eventId/entityId/eventType/version/occurredAt/actor/payload) + `Actor` (id/source) — форма домена, мета не в агрегате (D-15/EVT-02)
- Доменные sentinel (`ErrInvalidID`/`ErrInvalidTransition`/`ErrAlreadyDecommissioned`) + типизированные конфликты `ErrFQDNConflict`/`ErrProjectNotEmpty` (D-11/Pitfall 7 — не сырой DB E11000)
- 12 доменных портов в `ports.go` (D-02) + 12 mockery-моков в `internal/domain/mocks` (mockery v3.7.1, DO NOT EDIT)
- Ginkgo-suite бутстрап + прямые unit-спеки ID-VO (DescribeTable по 5 типам) и aggregateBase (Pitfall 3/5)

## Task Commits

Each task was committed atomically:

1. **Task 1: Типизированные ID-VO + sentinel-ошибки** - `6c06a6e` (feat)
2. **Task 2: aggregateBase + DomainEvent/EventEnvelope/Actor** - `a2a8cb3` (feat)
3. **Task 3: Доменные порты + .mockery.yaml + генерация моков** - `b7d9db2` (feat)

_TDD-задачи 1 и 2: в Go тест-файлы не компилируются без типов, поэтому prod-код и спеки в одном коммите; RED-проверка выполнена через прогон `go test` после написания (зелёный = поведение покрыто)._

## Files Created/Modified
- `services/inventory/internal/domain/id.go` - 5 типизированных ID-VO struct-over-uuid с factory/parse/string/zero
- `services/inventory/internal/domain/errors.go` - доменные sentinel + ErrFQDNConflict/ErrProjectNotEmpty
- `services/inventory/internal/domain/aggregate.go` - aggregateBase (record/PullEvents/Version)
- `services/inventory/internal/domain/events.go` - DomainEvent + EventEnvelope + Actor
- `services/inventory/internal/domain/ports.go` - 12 доменных портов (ctx-first)
- `services/inventory/internal/domain/aggregates_stub.go` - Wave-1 placeholder-агрегаты (Rule 3)
- `services/inventory/internal/domain/domain_suite_test.go` - Ginkgo suite bootstrap
- `services/inventory/internal/domain/id_test.go` - ID-VO unit specs (DescribeTable x5)
- `services/inventory/internal/domain/aggregate_test.go` - aggregateBase white-box specs (Pitfall 3/5)
- `services/inventory/internal/domain/mocks/*.go` - 12 сгенерированных mockery-моков портов
- `.mockery.yaml` - packages-запись internal/domain с 12 портами
- `services/inventory/go.mod` - google/uuid promoted indirect→direct

## Decisions Made
- **MatchAdvisor без HostHardware:** сигнатура `Candidates(ctx, fqdn string) ([]HostID, error)` — план рекомендовал этот вариант, чтобы `ports.go` компилировался в Wave 1 без forward-зависимости от типа `HostHardware` (он рождается в 06-02, Wave 2). hw-аргумент добавится с advisory-движком (SEED-001).
- **aggregate_test.go — white-box:** `aggregateBase`, `record()` и поля `version`/`events` неэкспортируемы by design (события рождаются только внутри доменных переходов). Тест помещён в `package domain` (не `domain_test`), чтобы встроить базу и дёрнуть `record()`. Suite-бутстрап и ID-тесты остаются black-box (`package domain_test`) — оба пакета сосуществуют в каталоге.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Wave-1 placeholder-агрегаты для компиляции ports.go**
- **Found during:** Task 3 (доменные порты)
- **Issue:** `ports.go` ссылается на агрегаты `Host`/`Project`/`DC`/`Module`/`Rack` в сигнатурах репозиториев (`Save(ctx, *Host)` и т.д.), но сами агрегаты рождаются позже: Host — в 06-02, остальные — в 06-03. `go build ./internal/domain/...` падал с 10+ `undefined: Host/Project/...`. План требует `go build ./...` exit 0 уже в Wave 1 (acceptance criterion Task 3), но flagged forward-зависимость только для `HostHardware`/MatchAdvisor, упустив сами struct-агрегаты.
- **Fix:** Создан `aggregates_stub.go` с минимальными placeholder-агрегатами: каждый `struct{ aggregateBase; id XID }` + геттер `ID() XID`. Файл явно документирует, что 06-02/06-03 РАСШИРЯЮТ эти типы (добавляют поля/методы/фабрики/события в host.go/project.go/...), а НЕ переобъявляют `type Host struct` (иначе redeclaration).
- **Files modified:** services/inventory/internal/domain/aggregates_stub.go
- **Verification:** `go build ./...` exit 0; `go vet ./internal/domain/...` чистый; `go test ./internal/domain/...` зелёный; mockery сгенерировал моки репозиториев по этим типам.
- **Committed in:** `b7d9db2` (Task 3 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Заглушки необходимы для компиляции Wave 1, как требует план. Влияние на 06-02/06-03: те планы должны РАСШИРИТЬ существующие типы `Host`/`Project`/`DC`/`Module`/`Rack` (добавить поля в struct + методы в свои файлы), а не переобъявлять `type X struct` — иначе redeclaration error. Это контракт для следующих волн (см. ниже Next Phase Readiness).

## Issues Encountered
None — планируемая работа прошла без проблем сверх задокументированной выше forward-зависимости.

## User Setup Required
None - внешних сервисов/конфигурации не требуется (чистый домен на фейках).

## Known Stubs

| Stub | File | Reason |
|------|------|--------|
| Wave-1 placeholder-агрегаты Host/Project/DC/Module/Rack | services/inventory/internal/domain/aggregates_stub.go | Минимальные struct{aggregateBase; id XID}+ID(); порты должны компилироваться в Wave 1 до рождения агрегатов. Резолвятся: Host — 06-02 (lifecycle/state-machine/события), Project/DC/Module/Rack — 06-03 (CRUD-агрегаты). Эти планы расширяют типы, не переобъявляют. |

## Next Phase Readiness
- **Готово для 06-02 (Host):** aggregateBase для встраивания, DomainEvent/EventEnvelope-форма, ID-VO, sentinel-ошибки, HostRepository/ActiveHostByFQDN-порты + моки. **Контракт:** наполнять `type Host struct` (поля/lifecycle/события) в host.go, расширяя существующую заглушку из aggregates_stub.go — НЕ переобъявлять.
- **Готово для 06-03 (Project/DC/Module/Rack):** аналогично — расширять placeholder-типы.
- **Готово для 06-04 (usecases):** все порты + Clock/IDGenerator/MatchAdvisor для no-op заглушки и детерминизма envelope-тестов.
- **Концерн:** MatchAdvisor.Candidates пока без HostHardware-аргумента; при появлении advisory-движка (SEED-001) сигнатуру расширят — потребует правки .mockery.yaml + regen.

## Self-Check: PASSED

Все заявленные файлы существуют на диске (id/errors/aggregate/events/ports/aggregates_stub.go, 12 моков, SUMMARY); коммиты задач 6c06a6e/a2a8cb3/b7d9db2 присутствуют в git log.

---
*Phase: 06-inventory*
*Completed: 2026-06-30*
