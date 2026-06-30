---
phase: 06-inventory
plan: 02
subsystem: domain
tags: [go, ddd, hexagonal, value-object, immutability, lifecycle-state-machine, domain-events, ginkgo, gomega]

# Dependency graph
requires:
  - phase: 06-inventory
    plan: 01
    provides: "aggregateBase (record/PullEvents/Version), DomainEvent-форма, типизированные ID-VO (HostID/ProjectID/RackID), доменные sentinel-ошибки, HostRepository-порт + Host-заглушка"
provides:
  - "Host-агрегат: фабрика NewHost с инвариантами идентичности/привязки + lifecycle state-machine + 6 операций-событий"
  - "lifecycleState — ровно 3 члена (shadow/registered/decommissioned); deleted НЕ enum"
  - "immutable HostHardware VO со структурированными под-VO (Motherboard/RAMModule/CPU/Drive/NIC/PSU/StorageController/GPU/Chassis) + HardwareSpec вход"
  - "6 доменных событий Host (HostRegistered/HostHardwareChanged/HostReassigned/HostRelocated/HostDecommissioned/HostDeleted)"
  - "доменные sentinel ErrInvalidHardware + ErrMissingProject"
affects: [06-03-locations, 06-04-usecases, 06-05-crud-usecases, 06-06-events, phase-07-mongo, phase-08-protobuf]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "immutable VO с defensive-copy слайсов в конструкторе И геттерах + глубокая копия вложенного NIC.MACs (Pitfall 2)"
    - "factory-holds-invariants: NewHost проверяет projectID.IsZero()/initial-state, генерирует ID, record(HostRegistered)"
    - "3-member lifecycle SM с iota; deleted = метод Delete()+repo.Delete, не enum-член (Pitfall 1/D-09)"
    - "одно семантическое событие на операцию (D-13), минимально-достаточный payload, не HostUpdated-дамп"
    - "ChangeHardware заменяет VO целиком → один HostHardwareChanged (D-07), не per-компонент"
    - "white-box тесты (package domain) для unexported lifecycleState/state-машины"

key-files:
  created:
    - services/inventory/internal/domain/hardware.go
    - services/inventory/internal/domain/hardware_test.go
    - services/inventory/internal/domain/host.go
    - services/inventory/internal/domain/host_test.go
  modified:
    - services/inventory/internal/domain/aggregates_stub.go
    - services/inventory/internal/domain/errors.go

key-decisions:
  - "host_test.go — white-box (package domain): lifecycleState и члены stateShadow/stateRegistered/stateDecommissioned неэкспортируемы, матрица переходов и NewHost(initial) требуют прямого доступа"
  - "Reassign на zero-value ProjectID → ErrMissingProject (привязка INV-02 остаётся обязательной и при переназначении, не только при создании)"
  - "Reassign/Relocate/ChangeHardware на списанном хосте → ErrAlreadyDecommissioned (терминальность D-10 распространяется на все мутации, не только повторный Decommission)"
  - "HostDeleted несёт полный snapshot-payload (projectID/fqdn/rackID/position) для аудит-следа (D-09); HostHardwareChanged несёт лишь HardwareName, не дамп компонентов (D-13)"

patterns-established:
  - "immutable VO: приватные поля + конструктор NewX(spec)+ошибка + value/copy-геттеры; defensive-copy слайсов (canonical для будущих VO)"
  - "агрегат: встраивание aggregateBase + фабрика-инвариант + операции, каждая через record(событие)"
  - "DescribeTable по матрице переходов lifecycle + It на one-event-per-op через BeAssignableToTypeOf"

requirements-completed: [INV-02, INV-04, INV-05, INV-06, INV-07, HW-01, HW-02, HW-03, HW-04, HW-05, HW-06, LOC-03]

# Metrics
duration: ~5min
completed: 2026-06-30
---

# Phase 6 Plan 02: Host-агрегат + HostHardware VO Summary

**Центральный агрегат фазы Host: immutable HostHardware VO со всеми структурированными под-компонентами железа (defensive-copy против Pitfall 2), фабрика NewHost с инвариантами идентичности/привязки, 3-членный lifecycle state-machine (терминальный decommission, deleted≠enum) и 6 операций-переходов — каждая эмитит ровно одно семантическое событие; go build/vet/test зелёные.**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-06-30T19:15:27Z
- **Completed:** 2026-06-30T19:20:15Z
- **Tasks:** 2
- **Files modified:** 6 (4 созданы + 2 изменены)

## Accomplishments

- **HostHardware** — единый immutable VO (`hardware.go`): приватные поля `name/platform/motherboard/ipmiMAC` + структурированные слайсы `ram/cpu/drives/nics/psus/storageCtl/gpus` + `chassis`. Под-VO `Motherboard`/`RAMModule`/`CPU`/`Drive`/`NIC`/`PSU`/`StorageController`/`GPU`/`Chassis` (HW-01…05).
- **NIC структурирован** (`Model`/`SpeedGbE`/`MACs []string`), а не плоский MACs[] на хосте (HW-03); все внешние ID (serial/inv/MAC/model) — raw `string` (HW-06).
- **Immutability (Pitfall 2):** defensive-copy слайсов в конструкторе И в каждом слайс-геттере; для NIC — **глубокая** копия вложенного `MACs[]` (отдельный `copyNICs`). Тесты подтверждают неизменность VO при мутации возвращённого слайса и при мутации исходного спека.
- **NewHostHardware** валидирует непустое `name` → `ErrInvalidHardware` (V5 input validation); вход — публичный `HardwareSpec` (DTO→спек на edge).
- **Host-агрегат** (`host.go`): встраивает `aggregateBase`; фабрика `NewHost` проверяет `projectID.IsZero()` (INV-02), генерирует `HostID` (INV-03), несёт `RackID`+позицию (LOC-03), допускает старт shadow/registered (D-10), эмитит ровно один `HostRegistered` (EVT-01).
- **lifecycleState** — ровно 3 члена (`stateShadow`/`stateRegistered`/`stateDecommissioned`); `deleted` НЕ член enum (Pitfall 1/D-09).
- **Операции-переходы**, каждая через `record()` ровно одно событие (D-13): `Reassign`→`HostReassigned`, `Relocate`→`HostRelocated`, `ChangeHardware`→`HostHardwareChanged` (замена VO целиком, D-07), `Decommission`→`HostDecommissioned` (терминально, ErrAlreadyDecommissioned при повторе), `Delete()`→`HostDeleted` (snapshot-payload, не трогает lifecycleState).
- **6 доменных событий** реализуют `DomainEvent` (`EventType()`/`EntityID()` = `HostID.String()`).
- Прямые unit-спеки (Ginkgo/Gomega): hardware — структура/immutability/raw-string/валидация; host — `DescribeTable` матрицы переходов, факторные инварианты, one-event-per-op (`BeAssignableToTypeOf`), Delete≠state, `Version()`==число операций (Pitfall 3).

## Task Commits

1. **Task 1: immutable HostHardware VO** — `49b60ca` (feat)
2. **Task 2: Host-агрегат — фабрика, lifecycle SM, операции-события** — `372c9e3` (feat)

_TDD-задачи: в Go тест-файлы не компилируются без типов, поэтому prod-код и спеки в одном коммите; RED-эквивалент — прогон `go test` после написания (зелёный = поведение покрыто), как в 06-01._

## Files Created/Modified

- `services/inventory/internal/domain/hardware.go` — immutable HostHardware VO + 9 под-VO + HardwareSpec + конструктор/геттеры с defensive-copy
- `services/inventory/internal/domain/hardware_test.go` — структура/immutability/raw-string/валидация (black-box)
- `services/inventory/internal/domain/host.go` — Host-агрегат: фабрика, lifecycle SM, 6 операций, 6 событий
- `services/inventory/internal/domain/host_test.go` — фабрика/lifecycle/события/version (white-box)
- `services/inventory/internal/domain/aggregates_stub.go` — удалена Host-заглушка (контракт Wave 1); Project/DC/Module/Rack нетронуты (06-03)
- `services/inventory/internal/domain/errors.go` — +ErrInvalidHardware, +ErrMissingProject

## Decisions Made

- **host_test.go — white-box (`package domain`):** `lifecycleState` и его члены неэкспортируемы by design; матрица переходов через `DescribeTable` и `NewHost(initial)` требуют прямого доступа к константам. hardware_test.go остаётся black-box (`package domain_test`) — VO полностью покрывается публичным API. Оба пакета сосуществуют в каталоге (как `aggregate_test.go`/`id_test.go` в 06-01).
- **Reassign требует непустой ProjectID:** инвариант обязательной привязки (INV-02) держится и при переназначении, не только при создании — zero-value → `ErrMissingProject`.
- **Терминальность распространяется на все мутации:** Reassign/Relocate/ChangeHardware на списанном хосте → `ErrAlreadyDecommissioned` (D-10: decommissioned терминально, воскрешения/изменений нет). Delete() — единственная операция из любого состояния.
- **Payload-гранулярность (D-13):** HostDeleted несёт полный snapshot (projectID/fqdn/rackID/position) для аудит-следа (D-09); HostHardwareChanged несёт лишь HardwareName, а не дамп компонентов.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Доменные sentinel ErrInvalidHardware + ErrMissingProject**
- **Found during:** Task 1 (валидация конструктора) и Task 2 (фабрика NewHost)
- **Issue:** План требует «доменную ошибку (V5 input validation)» при пустом обязательном поле HostHardware и при zero-value ProjectID в NewHost (INV-02), но в errors.go (06-01) таких sentinel не было — были только ErrInvalidID/ErrInvalidTransition/ErrAlreadyDecommissioned.
- **Fix:** Добавлены `ErrInvalidHardware` (V5 input validation для HostHardware) и `ErrMissingProject` (обязательная привязка INV-02) как доменные sentinel в errors.go, сравнимые через errors.Is.
- **Files modified:** services/inventory/internal/domain/errors.go
- **Committed in:** `49b60ca` (Task 1) — ErrInvalidHardware/ErrMissingProject объявлены вместе.

**2. [Cross-plan contract] Удаление Host-заглушки из aggregates_stub.go**
- **Found during:** Task 2 (определение реального Host)
- **Issue:** 06-01 оставил placeholder `type Host struct{aggregateBase; id HostID}`+`ID()` в aggregates_stub.go, чтобы ports.go компилировался в Wave 1. Реальный Host в host.go вызвал бы "Host redeclared in this block".
- **Fix:** Удалён Host-stub-тип и его `ID()` из aggregates_stub.go (заменён комментарием-маркером); реальный Host определён в host.go. Заглушки Project/DC/Module/Rack оставлены нетронутыми (владелец — 06-03).
- **Files modified:** services/inventory/internal/domain/aggregates_stub.go
- **Committed in:** `372c9e3` (Task 2)

---

**Total deviations:** 2 (1 Rule 2, 1 cross-plan contract — обе ожидаемы).
**Impact on plan:** Нулевой на scope; обе правки прямо предписаны планом (V5-валидация) и cross-plan-контрактом (удаление заглушки). Влияние на 06-03: Project/DC/Module/Rack-заглушки нетронуты, контракт расширения сохранён.

## Issues Encountered

None — планируемая работа прошла без проблем. gofmt-выравнивание struct-полей применено автоматически (gofumpt -w).

## User Setup Required

None — чистый домен на фейках, внешних сервисов/конфигурации не требуется.

## Known Stubs

None в рамках этого плана. Host-заглушка устранена (реальный агрегат); оставшиеся Project/DC/Module/Rack-заглушки в aggregates_stub.go — НЕ этого плана (резолвятся 06-03), документированы в 06-01 SUMMARY.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| (none) | — | Новой security-relevant поверхности сверх плана threat_model не введено. T-06-05…08 покрыты: терминальный decommission, hard-delete+аудит-событие, immutable VO defensive-copy, обязательная привязка в фабрике. |

## Next Phase Readiness

- **Готово для 06-03 (Project/DC/Module/Rack):** паттерн immutable-VO + factory-holds-invariants + record-на-операцию установлен; расширять placeholder-типы в aggregates_stub.go, не переобъявлять.
- **Готово для 06-04 (usecases):** Host-агрегат + 6 событий + HostRepository-порт (06-01) + Clock/IDGenerator для envelope-enrichment. RegisterHost/Decommission/Delete/Reassign/Relocate/ChangeHardware-interactors дёргают соответствующие методы Host.
- **Готово для 06-06 (events):** 6 Host-событий реализуют DomainEvent; envelope-обёртка на границе usecase.
- **Концерн:** FQDN-уникальность (ActiveHostByFQDN) и optimistic-concurrency по Version() — enforcement в Phase 7 (Mongo partial index + tx); домен лишь несёт version и типизированный ErrFQDNConflict.

## Self-Check: PASSED

Все заявленные файлы существуют на диске (hardware.go/hardware_test.go/host.go/host_test.go созданы; errors.go/aggregates_stub.go изменены); коммиты задач 49b60ca/372c9e3 присутствуют в git log; go build ./... / go vet / go test ./internal/domain/... зелёные.

---
*Phase: 06-inventory*
*Completed: 2026-06-30*
