---
phase: 06-inventory
plan: 03
subsystem: domain
tags: [go, ddd, hexagonal, aggregate, domain-events, locations, ginkgo, gomega]

# Dependency graph
requires:
  - phase: 06-inventory
    plan: 01
    provides: "ID-VO (ProjectID/DCID/ModuleID/RackID), aggregateBase (record/PullEvents/Version), DomainEvent-форма, доменные sentinel, ProjectRepository/DCRepository/ModuleRepository/RackRepository-порты; placeholder-стабы Project/DC/Module/Rack в aggregates_stub.go"
provides:
  - "Project-агрегат: NewProject (инвариант non-empty name INV-01, Owner raw string INV-09), операции Rename/ChangeOwner/Delete по одному событию (D-13)"
  - "DC/Module/Rack — три независимых локационных агрегата (D-04) с CRUD; иерархия по внутреннему ID (Module.DCID, Rack.ModuleID — LOC-02/D-06)"
  - "PowerTopology VO — топологические атрибуты питания стойки (LOC-04)"
  - "Семантические события: ProjectCreated/Renamed/OwnerChanged/Deleted; DCCreated/Updated/Deleted; ModuleCreated/Updated/Deleted; RackCreated/Updated/Deleted (EVT-01/D-13)"
  - "ErrInvalidProject/ErrInvalidLocation доменные sentinel"
  - "aggregates_stub.go удалён — все 5 агрегатов теперь реальные типы (Host в 06-02, остальные здесь)"
affects: [06-04-usecases, 06-05-crud-usecases, 06-06-events, phase-07-mongo, phase-08-protobuf]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "NewX держит инварианты + record(событие) — рецепт patterns.md Рецепт 3 (зеркалит host.go из 06-02)"
    - "Три независимых локационных агрегата (D-04), НЕ дерево; иерархия выражена ссылкой снизу вверх по внутреннему ID (D-06)"
    - "Guard на zero parent-ID в фабрике (NewModule/NewRack) — нет висячей локации без родителя (T-06-09)"
    - "Внешние строковые идентификаторы (Owner, линии питания) — raw string без резолва (INV-09/HW-06)"
    - "Одно семантическое событие на CRUD-операцию (D-13), не Updated-дамп всех полей"

key-files:
  created:
    - services/inventory/internal/domain/project.go
    - services/inventory/internal/domain/project_test.go
    - services/inventory/internal/domain/dc.go
    - services/inventory/internal/domain/module.go
    - services/inventory/internal/domain/rack.go
    - services/inventory/internal/domain/location_test.go
  modified:
    - services/inventory/internal/domain/errors.go
  deleted:
    - services/inventory/internal/domain/aggregates_stub.go

key-decisions:
  - "Owner/description в NewProject опциональны (только name обязательно, INV-01): владелец назначается позже через ChangeOwner — A7 Claude's Discretion в рамках D-13"
  - "PowerTopology как маленький VO {PowerSource, Generator} (raw string идентификаторы линий) — топология энергоснабжения стойки, отдельно от HostHardware (LOC-04)"
  - "ErrInvalidProject/ErrInvalidLocation — две новые доменные sentinel вместо переиспользования общего ErrInvalidID (битый uuid ≠ пустое обязательное поле); расширяет errors.go из 06-01"
  - "Родительский parent-ID (Module.DCID, Rack.ModuleID) неизменяем в Update — перенос локации между родителями вне scope CRUD; CRUD-update меняет только редактируемые атрибуты"

requirements-completed: [INV-01, INV-03, INV-09, LOC-01, LOC-02, LOC-04, EVT-01]

# Metrics
duration: ~4min
completed: 2026-06-30
---

# Phase 6 Plan 03: Project и локационные агрегаты Inventory Summary

**Project-агрегат (идентичность + Owner raw string + операции по одному событию) и три независимых локационных агрегата DC/Module/Rack с CRUD, иерархией по внутреннему ID и топологией питания Rack как VO — каждая операция эмитит ровно одно семантическое событие; cross-plan стаб-файл aggregates_stub.go удалён, домен компилируется и тесты зелёные.**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-06-30T19:24:10Z
- **Completed:** 2026-06-30T19:28:15Z
- **Tasks:** 2
- **Files modified:** 8 (6 created + 1 modified + 1 deleted)

## Accomplishments
- **Project-агрегат** (`project.go`): `NewProject(name, description, owner)` держит инвариант non-empty name (INV-01), генерирует `ProjectID` (INV-03), хранит `owner` как непрозрачный raw string без резолва (INV-09/T-06-10); операции `Rename`/`ChangeOwner`/`Delete` эмитят по одному семантическому событию (`ProjectRenamed`/`ProjectOwnerChanged`/`ProjectDeleted` — D-13); геттеры ID/Name/Description/Owner
- **DC/Module/Rack** — три независимых агрегата (D-04, НЕ дерево-Location): каждый встраивает `aggregateBase`, свой корень с CRUD (`New*`/`Update`/`Delete`)
- **Иерархия по внутреннему ID** (LOC-02/D-06): `Module` несёт `DCID`, `Rack` несёт `ModuleID` — не вложенные объекты, не back-references; `NewModule`/`NewRack` guard на zero parent-ID → `ErrInvalidLocation` (нет висячей локации, T-06-09)
- **Rack топология** (LOC-04): `PowerTopology{PowerSource, Generator}` VO — идентификаторы линии питания/дизель-генератора как raw string
- **Семантические события** (EVT-01/D-13): 4 для Project + по 3 на каждую локацию (Created/Updated/Deleted), все реализуют `DomainEvent` (EntityID = строковый ID — будущий Kafka-ключ)
- **Cross-plan контракт исполнен:** удалены стабы Project/DC/Module/Rack из `aggregates_stub.go`; файл опустел → `git rm` (Host стаб уже снят в 06-02). Все 5 агрегатов теперь реальные типы
- **Тесты** (`project_test.go`/`location_test.go`): прямые black-box unit-спеки Ginkgo/Gomega без моков (D-03) — инварианты фабрик, guard на zero parent-ID, одно-событие-на-операцию, version-трекинг

## Task Commits

Each task was committed atomically:

1. **Task 1: Project-агрегат + операции/события** - `8d6b273` (feat)
2. **Task 2: Локационные агрегаты DC/Module/Rack (иерархия по ID)** - `b4bf9cb` (feat)

_TDD-задачи: в Go тест-файлы не компилируются без типов, поэтому prod-код и спеки в одном коммите; RED-проверка — прогон `go test` после написания (зелёный = поведение покрыто). Тот же подход, что в 06-01/06-02._

## Files Created/Modified
- `services/inventory/internal/domain/project.go` - Project-агрегат: фабрика + Rename/ChangeOwner/Delete + 4 события
- `services/inventory/internal/domain/project_test.go` - прямые unit-спеки Project (инварианты, события, version)
- `services/inventory/internal/domain/dc.go` - DC-агрегат: NewDC/Update/Delete + DCCreated/Updated/Deleted
- `services/inventory/internal/domain/module.go` - Module-агрегат (несёт DCID) + ModuleCreated/Updated/Deleted
- `services/inventory/internal/domain/rack.go` - Rack-агрегат (несёт ModuleID + PowerTopology VO) + RackCreated/Updated/Deleted
- `services/inventory/internal/domain/location_test.go` - прямые unit-спеки на три локационных агрегата (CRUD, guard на zero parent-ID, иерархия по ID, топология)
- `services/inventory/internal/domain/errors.go` - добавлены ErrInvalidProject/ErrInvalidLocation
- `services/inventory/internal/domain/aggregates_stub.go` - **УДАЛЁН** (все стабы стали реальными типами)

## Decisions Made
- **Owner/description опциональны:** план разрешил executor'у решить обязательность (A7 в рамках INV-01). Обязателен только `name`; `owner`/`description` могут быть пустыми — владелец назначается позже (`ChangeOwner`), это ложится на «Owner — внешний string, домен не доверяет» (INV-09).
- **PowerTopology как VO:** топологические атрибуты питания (LOC-04) вынесены в маленький value-object `{PowerSource, Generator}` (raw string идентификаторы линий), а не плоские поля — даёт точку расширения под анализ влияния отказа питания и держит ясную границу с `HostHardware`.
- **Новые sentinel вместо ErrInvalidID:** `ErrInvalidProject`/`ErrInvalidLocation` для пустых обязательных полей и zero parent-ID — семантически отличны от `ErrInvalidID` (битая uuid-строка при parse). Расширяет errors.go из 06-01.
- **Parent-ID неизменяем в Update:** `Module.DCID`/`Rack.ModuleID` фиксируются в фабрике; CRUD-`Update` меняет только редактируемые атрибуты (name/location/power). Перенос локации между родителями — вне scope этого CRUD.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Доменные sentinel для валидации фабрик**
- **Found during:** Task 1 (Project) и Task 2 (локации)
- **Issue:** План требует доменную ошибку при пустом обязательном поле (V5) и при zero parent-ID, но в errors.go (06-01) не было подходящего sentinel — был только `ErrInvalidID` (битый uuid при parse), семантически другой.
- **Fix:** Добавлены `ErrInvalidProject` (пустое name / переименование) и `ErrInvalidLocation` (пустое поле локации или висячая привязка по zero parent-ID) в errors.go — паттерн sentinel из 06-01 (style.md: предсказуемые исходы инвариантов через errors.Is).
- **Files modified:** services/inventory/internal/domain/errors.go
- **Verification:** `go build`/`go vet`/`go test ./internal/domain/...` зелёные; спеки ассертят `MatchError(ErrInvalidProject/ErrInvalidLocation)`.
- **Committed in:** `8d6b273` (Task 1, ErrInvalidLocation добавлен заранее для Task 2)

---

**Total deviations:** 1 auto-fixed (Rule 2)
**Impact on plan:** Минимальный — sentinel-ошибки необходимы для V5-валидации, которую план явно требует. Cross-plan стаб-контракт исполнен полностью.

## Cross-Plan Contract Fulfilled
- Прочитан 06-01 SUMMARY и комментарии в `aggregates_stub.go`.
- Удалены стабы `Project`/`DC`/`Module`/`Rack` (и их `ID()`-геттеры) из `aggregates_stub.go`.
- После удаления файл опустел (Host-стаб уже снят 06-02) → `git rm aggregates_stub.go`.
- Реальные агрегаты определены в `project.go`/`dc.go`/`module.go`/`rack.go`.
- Пакет `domain` компилируется; `go vet`/`go test ./internal/domain/...` зелёные; `ports.go` (ProjectRepository/DCRepository/ModuleRepository/RackRepository) ссылается на реальные типы без redeclaration.

## Issues Encountered
None — плановая работа без проблем сверх задокументированного sentinel-дополнения.

## User Setup Required
None — чистый домен на прямых unit-тестах, внешних сервисов/конфигурации не требуется.

## Known Stubs
None — `aggregates_stub.go` удалён; все 5 агрегатов (Host + Project/DC/Module/Rack) — реальные типы с полями/фабриками/событиями.

## Next Phase Readiness
- **Готово для 06-04/06-05 (usecases):** все агрегаты Project/DC/Module/Rack с фабриками-инвариантами, операциями и `PullEvents()`; порты репозиториев (ports.go) + моки уже есть. CRUD-usecase оркеструет `uow.Do` → `repo.Save` → `enrich(PullEvents())` → `outbox.Append`; delete-only-if-empty — через query-порт `HostsInProject` (Project) / аналогичные для локаций.
- **Готово для 06-06 (события):** имена семантических событий зафиксированы как контракт (ProjectCreated/.../RackDeleted) — войдут в DOC-07 glossary и protobuf-схемы Phase 8.
- **Концерн:** инвариант непустоты при Delete (DC без Module, Module без Rack, Rack без хостов) — агрегатный `Delete` лишь эмитит факт; enforcement delete-only-if-empty живёт в usecase (требует query-портов количества дочерних сущностей, которых пока нет для локаций — добавятся в 06-05 при необходимости).

## Self-Check: PASSED

Все заявленные файлы существуют на диске (project/dc/module/rack.go + project_test/location_test.go, errors.go модифицирован, aggregates_stub.go удалён); коммиты задач 8d6b273/b4bf9cb присутствуют в git log; `go build`/`go vet`/`go test ./internal/domain/...` exit 0.

---
*Phase: 06-inventory*
*Completed: 2026-06-30*
