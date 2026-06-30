# Phase 6: Доменная модель Inventory - Research

**Researched:** 2026-06-30
**Domain:** Go DDD/гексагональная доменная модель (агрегаты, VO, доменные события, порты) + interactor-usecases на in-memory фейках. Чистый Go-домен — без Mongo/gRPC/Kafka/protobuf (Phases 7/8).
**Confidence:** HIGH (решения локнуты в CONTEXT.md D-01…D-15; паттерны — идиоматичный Go, верифицированы против существующего кода репозитория и канона knowledge/; технических внешних зависимостей нет)

## Summary

Phase 6 — это **дизайн-фаза без новой инфраструктуры**: проектируем доменное ядро Inventory (агрегаты `Project`/`Host`/`DC`/`Module`/`Rack`, immutable VO `HostHardware`, типизированные ID-VO, семантические доменные события с envelope-границей) и interactor-usecases, оркестрирующие через порты `UnitOfWork`/`Outbox`/repo/uniqueness, реализованные **in-memory фейками**. Цель формы кода: Phase 7 = чистый свап фейков на Mongo, usecase не меняется (предотвращает Pitfall 8 dual-write на уровне формы).

Все ключевые решения уже зафиксированы в CONTEXT.md (D-01…D-15) и в `.planning/research/PITFALLS.md`. Исследование **не** переоткрывает их — оно даёт планнеру конкретные **идиоматичные Go-формы** для: типизированных ID-VO над `uuid.UUID`; immutable `HostHardware` VO со структурированными под-компонентами; aggregate-base с накоплением событий и `PullEvents()`; envelope-enrichment-границы между `PullEvents()` и `outbox.Append`; сигнатур портов и их фейков; рецепта interactor-usecase; и тестовой стратегии (прямые unit-тесты на агрегаты + mockery-моки на порты).

Главные граблезоны Phase 6 (из PITFALLS «Looks Done But Isn't»): (2) `decommissioned`=lifecycle-state vs `deleted`=hard-delete без `state=deleted`-строки; (5) семантические события vs `HostUpdated`-дамп; (6) `actor` в envelope с первого дня; (7) FQDN-конфликт = типизированный доменный конфликт через query-порт, не сырой DB-error; (8) форма usecase `uow.Do(fn){repo.Save + outbox.Append}` уже сейчас, чтобы Phase 7 был свапом.

**Primary recommendation:** Строить домен на трёх примитивах — (1) типизированный ID-VO per aggregate (`type HostID struct{ v uuid.UUID }` с фабрикой/parse/zero-check), (2) встраиваемый `aggregateBase` с `record()`/`PullEvents()`/`version`, (3) порты, объявленные в `domain`, с in-memory фейками в `usecases`-тестах И mockery-моками. Каждый usecase следует Рецепту 1 из patterns.md: фабрика/метод агрегата → `uow.Do(ctx, fn)` → внутри `fn`: `repo.Save` + `outbox.Append(agg.PullEvents()-enriched)`. Envelope навешивается на границе usecase, НЕ в агрегате.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Срез Phase 6 ↔ Phase 7 (глубина вертикали):**
- **D-01:** Вертикаль = **домен + usecases**. Строим агрегаты/VO/события/фабрики/инварианты **и** interactor-usecases на write-операции (RegisterHost, DecommissionHost, DeleteHost, ReassignHost, RelocateHost, ChangeHostHardware, CreateProject/…, CRUD локаций). Делает SC1/SC3/SC5 верифицируемыми в Phase 6.
- **D-02:** Usecases **уже оркестрируют через порты** `UnitOfWork.Do(ctx, fn)` + `Outbox.Append` ровно как в architecture.md: внутри `fn` — `repo.Save(агрегат)` + `PullEvents()`→`outbox.Append`. Порты реализованы **фейками** (in-memory uow = просто зовёт `fn`; in-memory outbox = слайс). **Phase 7 = свап фейков на Mongo-impl, usecase НЕ меняется.**
- **D-03:** Тест-дублёры — **только mockery-моки** для портов (repo/uow/outbox/query-порты). Домен/агрегаты/инварианты тестируются **прямыми** unit-тестами без моков. Mockery + Ginkgo v2 + Gomega уже провязаны в Phase 5.

**Агрегаты, идентичность, локации:**
- **D-04:** Локации DC/Module/Rack — **три независимых агрегата**, каждый свой корень с CRUD. Иерархия — через ссылки по **внутреннему ID**: Module несёт DCID, Rack несёт ModuleID. Host→RackID + позиция (юнит). НЕ один агрегат-дерево Location.
- **D-05:** Внутренние ID — **типизированные ID-VO на агрегат** (`HostID`/`ProjectID`/`DCID`/`ModuleID`/`RackID`), обёртка над `uuid.UUID` (`github.com/google/uuid` уже в go.mod) с фабрикой/парсингом. Внешние идентификаторы (`Owner`, `INV`, serial, MAC и т.п.) — сырой `string`.
- **D-06:** Ссылки **между** агрегатами — только по внутреннему ID (не вложенные объекты, не back-references). Reassign — операция на Host (меняет ProjectID), не на Project; Project не держит список своих хостов.
- **D-07:** `HostHardware` — **единый immutable VO** со всеми вложенными компонентами (Motherboard/RAM[]/CPU[]/Drives[]/NIC[]/PSU[]/storage-controller/внутренний GPU/chassis, IPMIMac). Изменение железа = собрать новый VO и заменить целиком → **одно** событие `HostHardwareChanged`. NIC — структурированный компонент (модель/скорость/MAC'и), не плоский `MACs[]`.

**ЖЦ, удаление, конфликты:**
- **D-08:** `lifecycleState ∈ {shadow, registered, decommissioned}`. `decommissioned` — смена состояния, хост **остаётся видим**, это НЕ tombstone.
- **D-09:** `deleted` = hard-удаление, НЕ член enum'а `lifecycleState`. Агрегатный метод `Delete()` → эмитит `HostDeleted` (полный payload + actor), usecase зовёт `repo.Delete` (физическое удаление). Никакой строки `state=deleted` не остаётся. Канон: Pitfall 2.
- **D-10:** Граф переходов — **гибкий вход + терминальный decommission**. Host создаётся либо в `shadow`, либо сразу `registered`. Переходы: `shadow→registered`; `shadow→decommissioned`; `registered→decommissioned`. `decommissioned` **терминально**. `Delete()` — из любого состояния.
- **D-11:** **FQDN-конфликт среди active** — доменный инвариант, проверяется **в usecase** через query-порт (напр. `ActiveHostByFQDN`/uniqueness-checker) → возвращает **типизированный доменный конфликт**. Канон: Pitfall 7.
- **D-12:** **Advisory-matching (INV-08)** — объявляется как **порт-хук** (интерфейс) + заглушка (no-op/пустые кандидаты) в Phase 6. Авто-restore/merge запрещён by design.

**Доменные события + envelope:**
- **D-13:** **Одно семантическое событие на бизнес-операцию** (минимально-достаточный payload, не `HostUpdated`-дамп). Регистрация → один `HostRegistered`. Далее отдельные: `HostHardwareChanged`, `HostReassigned`, `HostRelocated`, `HostDecommissioned`, `HostDeleted`. Аналогично для Project и локаций. Имена событий = ubiquitous language, фиксируются в DOC-07.
- **D-14:** **Факты в домене + envelope на границе.** Агрегат эмитит «голые» семантические факты (`eventType`/`entityId`/`version` + payload) и копит их; `PullEvents()` сливает. Envelope-мета (`eventId`, `occurredAt`, `actor{id, source}`) **навешивается между `PullEvents()` и `outbox.Append`** на границе usecase. **Нет** Clock/IDGen/actor внутри агрегата.
- **D-15:** Envelope несёт `eventId`, `entityId`, `eventType`, `version`, `occurredAt`, `actor{id, source: human|api|integration|system}` — **с первого дня**. `version` — поле агрегата, инкрементится при каждом изменении.

### Claude's Discretion
- **DOC-07 glossary:** точный состав/формулировки терминов и границы «факт существования vs динамическое состояние» — на усмотрение планнера/executor'а в рамках D-08…D-15 и authoring.md. Обязательно фиксирует: Project/Host/Owner/Module/Connection, идентичность, `decommission ≠ delete`, имена семантических событий.
- **Какие канон-слои получают код в Phase 6 (SVC-01):** ожидаемо `domain` (агрегаты/VO/события/порты), `usecases` (interactor'ы), query-порт(ы) для uniqueness; `repositories`/`api`/`cron`/`app` предположительно остаются скелетом/фейками до Phase 7. Финальную раскладку определяет планнер.
- **Project-агрегат — операции/события** (rename, смена Owner, delete) и **инвариант уникальности позиции host↔rack** — на усмотрение планнера, в рамках D-13 и D-11.
- Имена пакетов, точные сигнатуры портов, форма типизированных ID-VO, структура фабрик — планнеру/executor'у.

### Deferred Ideas (OUT OF SCOPE)
- **Полный advisory-matching движок** (составной нестабильный матч INV+FQDN+MAC+локация+окно) — отдельный интеграционный сервис, не Inventory (SEED-001). В Phase 6 — только порт-хук + заглушка (D-12).
- **Реальная persistence/UoW(Mongo-txn)/Outbox-коллекция/gRPC/query-на-Mongo/partial FQDN-index** — Phase 7.
- **protobuf-схемы событий + relay→Kafka + dual-topic + tombstone-эмиссия** — Phase 8 (имена/форма событий фиксируются сейчас как контракт, кодоген и публикация — там).
- **Топология `connections` + read-model + внешние HW-модули** — Phase 9 (MOD-01…03).
- **VM/VMGroup, sync из внешней инвентори, Audit-домен, Access/права** — будущие эпики.
- Реальная optimistic-concurrency enforcement по `version` — Phase 7 (поле заводим сейчас, enforcement — там).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| INV-01 | Оператор может завести Project (`ID`, `Name`, `Description`, `Owner`) | Project-агрегат + фабрика + `CreateProject` usecase (Pattern: Aggregate Factory, Interactor) |
| INV-02 | Зарегистрировать Host с обязательной привязкой к Project | `RegisterHost` usecase; ProjectID обязателен в фабрике `NewHost` (фабрика-инвариант) |
| INV-03 | Внутренний постоянный непереиспользуемый `ID` — единственный носитель идентичности | Типизированный ID-VO over `uuid.UUID v4` (Pattern: Typed ID VO); генерируется в фабрике |
| INV-04 | ЖЦ-статус `shadow→registered→decommissioned` + `deleted` | `lifecycleState` enum (3 члена) + отдельный `Delete()`; D-08/D-09; см. Pitfall 2, Lifecycle state-machine |
| INV-05 | Переназначить Host в другой Project | `ReassignHost` usecase → `host.Reassign(newProjectID)` → `HostReassigned` (D-06: операция на Host) |
| INV-06 | Decommission хоста (терминально; ≠ deleted) | `DecommissionHost` usecase → `host.Decommission()` → terminal state + `HostDecommissioned` |
| INV-07 | Удалить (`deleted`) запись; история на событиях; FQDN освобождается | `DeleteHost` usecase → `host.Delete()` (event) + `repo.Delete()` (physical); no `state=deleted` row |
| INV-08 | Повторное добавление = новый ID; матч советочный (хук) | `MatchAdvisor` порт-хук + no-op заглушка (D-12); re-add = новая фабрика, новый ID |
| INV-09 | `Owner` — непрозрачный внешний `string`-ID | `Owner string` (raw string per D-05) |
| INV-10 | FQDN уникален среди active (доменный конфликт); Project удалить только пустым | `ActiveHostByFQDN` query-порт + `ErrFQDNConflict`; Project-delete invariant via host-count query-порт |
| HW-01 | `HostHardware` VO (`Name`/`Platform`/`Motherboard`/`IPMIMac`) | Immutable `HostHardware` VO (Pattern: Immutable VO); D-07 |
| HW-02 | Структурированные RAM/CPU/Drives (`slot`/`model`/`vendor`/`lot`/`serial`/`Inv`/спеки) | Структурированные под-VO внутри `HostHardware`; slices |
| HW-03 | NIC = структурированный компонент (модель/скорость/MAC'и) | `NIC` под-VO с `MACs []string`, не плоский `MACs[]` на хосте; D-07 |
| HW-04 | PSU (узлы power-зависимости) | `PSU` под-VO slice |
| HW-05 | storage-controller/RAID, внутренние GPU, паспорт шасси отдельно от материнки | Отдельные под-VO `StorageController`/`GPU`/`Chassis` ≠ `Motherboard` |
| HW-06 | Все внешние идентификаторы компонентов — `string` | Все serial/inv/MAC/model — raw `string` (D-05) |
| LOC-01 | Завести/изменить DC, Module, Rack как первоклассные сущности (CRUD) | Три независимых агрегата + CRUD usecases (D-04) |
| LOC-02 | Иерархия `DC → Module → Rack` | Ссылки по ID: Module.DCID, Rack.ModuleID (D-04, D-06) |
| LOC-03 | Host ссылается на Rack + позицию (юнит) | Host.RackID + Host.Position (юнит) |
| LOC-04 | Rack несёт топологические атрибуты (источник питания и т.п.) | Поля/VO на Rack-агрегате |
| EVT-01 | Семантические гранулярные события на все изменения | Один semantic event per операция (D-13); aggregate-base record() |
| EVT-02 | Envelope с `eventId`/`version`/`actor/initiator`/`occurredAt` с первого дня | Envelope на границе usecase (D-14/D-15); EventEnvelope struct |
| DOC-07 | `knowledge/glossary.md` — ubiquitous language | Структура glossary (см. Pattern: Glossary authoring); authoring.md MUST/SHOULD/WON'T |
| SVC-01 | Канон-слои `domain/usecases/query/repositories/api/cron` + `app` | architecture.md layer layout; в Phase 6 код в domain+usecases+query-порт |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Идентичность (ID-VO генерация, типы) | `domain` (агрегат/фабрика) | — | ID — носитель идентичности, рождается в фабрике агрегата (architecture.md §события) |
| ЖЦ-инварианты (state-machine переходы) | `domain` (методы агрегата) | — | Инварианты держит агрегат; usecase лишь оркеструет |
| HostHardware-валидация (структура VO) | `domain` (VO-конструктор) | — | VO-семантика — чистый домен, без I/O |
| Эмиссия семантических фактов | `domain` (aggregate-base record) | — | События рождаются в агрегате при переходе (D-14) |
| Envelope-enrichment (eventId/occurredAt/actor) | `usecases` (граница) | `domain` (форма envelope) | actor — транспортная identity, никогда не входит в агрегат (D-14); enrich между PullEvents и Append |
| FQDN-uniqueness проверка | `usecases` (вызов query-порта) | `domain` (порт + ErrFQDNConflict) | Доменный инвариант, требующий чтения хранилища → query-порт (D-11, Pitfall 7) |
| Транзакционная оркестрация (Save+Outbox) | `usecases` (`uow.Do`) | `domain` (порты UoW/Outbox/repo) | Граница транзакции — порт UnitOfWork (architecture.md §UnitOfWork) |
| Advisory-matching | `domain` (порт-хук) | `usecases` (вызов) | Хук под будущую интеграцию + no-op заглушка (D-12) |
| Persistence (Save/Delete тела) | (Phase 7 — `repositories`) | в Phase 6: фейк/mock | D-02: Phase 6 — фейки; реальный Mongo — Phase 7 |
| gRPC-адаптеры, identity-interceptor | (Phase 7 — `api`) | — | Вне scope Phase 6 |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/google/uuid` | v1.6.0 | Основа типизированных ID-VO; `uuid.New()` (v4 random) для генерации внутренних ID | `[VERIFIED: go list -m]` уже в go.mod (indirect → станет direct при первом импорте в domain). Канонический UUID-пакет Go-экосистемы. |
| Go stdlib `errors` | go1.25.0 | Sentinel-ошибки (`errors.New`, `errors.Is/As`), `fmt.Errorf("...: %w", err)` | `[VERIFIED: go version]` style.md канон: sentinel + `%w` (errorlint hook) |
| Go stdlib `context` | go1.25.0 | `context.Context` первым аргументом во всех портах/usecase Execute | Идиома Go; architecture.md сигнатуры портов |

### Supporting (тест-стек — уже провязан в Phase 5)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/onsi/ginkgo/v2` | v2.32.0 | BDD-раннер; `Describe/Context/It`, `DescribeTable` | Все спеки (testing.md канон); dot-import |
| `github.com/onsi/gomega` | v1.42.1 | Matcher-ассерты `Expect().To(...)`, `MatchError`, `BeAssignableToTypeOf` | Все ассерты; dot-import |
| `github.com/vektra/mockery` | v3.7.1 (pinned в Makefile) | Кодоген моков портов (testify-template, expecter-API) | `make generate-mocks`; моки для repo/uow/outbox/query-портов (D-03) |
| `github.com/stretchr/testify/mock` | v1.11.1 | `mock.Anything`, expecter runtime для сгенерированных моков | Обычный (не dot) импорт в спеках с моками |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `type HostID struct{ v uuid.UUID }` | `type HostID string` (как ExampleID/style.md плейсхолдер) | string-форма проще и в style.md показана на `Order`, но struct-over-uuid даёт строгий zero-value контроль и parse-валидацию. **Рекомендация: struct-обёртка** — D-05 явно говорит «обёртка над `uuid.UUID`», не над string. Планнер фиксирует точную форму. |
| `uuid.New()` (v4) | `uuid.NewV7()` (time-ordered) | v7 даёт сортируемость по времени (плюс для Kafka/Mongo-ключей), но v3.0 ключ = ID без требования сортировки; v4 проще и достаточен. **Рекомендация: v4** (`uuid.New()`); v7 — опционально, на усмотрение планнера. |
| Встраиваемый `aggregateBase` struct | Интерфейс `DomainEventRecorder` + ручной слайс в каждом агрегате | Встраивание убирает дублирование `record()/PullEvents()/version` между 5 агрегатами (DRY, и канон проекта «общее — в одном месте»). **Рекомендация: встраиваемый base** в `domain` (НЕ в pkg — он доменно-специфичен для inventory, не generic для всех сервисов). |
| Ручные in-memory фейки портов в `_test.go` | Только mockery-моки | D-02 требует in-memory **фейки** для uow/outbox (uow = просто зовёт fn; outbox = слайс), а D-03 — mockery-**моки** для expectation-проверок. Это не конфликт: фейки нужны там, где usecase реально гоняет fn и копит события (проверяем итог); моки — где проверяем сам факт/порядок вызовов. testing.md SHOULD «не писать fake вручную» относится к мокам портов с поведением; uow/outbox-фейки настолько тривиальны, что допустимы как тестовый хелпер. **Планнер решает баланс.** |

**Installation:** Новые внешние пакеты **не устанавливаются**. `github.com/google/uuid` уже в go.mod (indirect); при первом импорте в `internal/domain` он станет direct после `go mod tidy`. Тест-стек (ginkgo/gomega/mockery/testify) провязан в Phase 5.

```bash
# проверка (не установка):
cd services/inventory && go list -m github.com/google/uuid   # => v1.6.0
```

**Version verification:** `github.com/google/uuid v1.6.0` `[VERIFIED: go list -m в services/inventory]`. Go toolchain `go1.25.0` `[VERIFIED: go version]`. Тест-версии — из go.mod (Phase 5).

## Package Legitimacy Audit

> Phase 6 устанавливает **ноль** новых внешних пакетов. Все используемые пакеты уже присутствуют в go.mod после Phase 5 и верифицированы тогда. Аудит легитимности новых пакетов не требуется.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| github.com/google/uuid | Go modules (proxy.golang.org) | стабильный (v1.6.0, многолетний) | широко используемый | github.com/google/uuid | OK | Уже в go.mod (Phase 5); промоутится indirect→direct |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

Поток write-операции (на примере RegisterHost) — концептуальные компоненты, не файлы:

```text
  gRPC DTO (Phase 7, вне scope)
        │  маппинг DTO→домен на edge (style.md)
        ▼
  ┌─────────────────────── usecases (interactor) ────────────────────────┐
  │  RegisterHostUseCase.Execute(ctx, in DomainInput, actor Actor)        │
  │   1. host, err := NewHost(in...)        ── фабрика держит инварианты   │
  │   2. uow.Do(ctx, func(ctx) error {                                    │
  │        a. existing := uniq.ActiveHostByFQDN(ctx, in.FQDN)             │
  │           if found → return ErrFQDNConflict{candidates}  (Pitfall 7)  │
  │        b. repo.Save(ctx, host)                                        │
  │        c. events := host.PullEvents()        ── голые факты (D-14)    │
  │        d. enriched := envelope(events, actor, clock, idgen) ── граница │
  │        e. outbox.Append(ctx, enriched)       ── та же tx (Pitfall 8)  │
  │      })                                                               │
  │   3. return RegisterHostOutput{ID: host.ID()}                        │
  └──────────────────────────────────────────────────────────────────────┘
        │ зависит ТОЛЬКО от портов (интерфейсов) из domain
        ▼
  ┌───────────────────────────── domain ──────────────────────────────────┐
  │  Host (aggregateBase) ── lifecycleState, version, []event, HostHardware │
  │  фабрики NewHost/NewProject/NewDC/NewModule/NewRack                     │
  │  ID-VO: HostID/ProjectID/DCID/ModuleID/RackID (over uuid.UUID)         │
  │  semantic events: HostRegistered/HostHardwareChanged/HostReassigned/…  │
  │  EventEnvelope{eventId,entityId,eventType,version,occurredAt,actor}    │
  │  ПОРТЫ (интерфейсы): UnitOfWork, Outbox, HostRepository,               │
  │                       ActiveHostByFQDN (uniqueness), MatchAdvisor      │
  └────────────────────────────────────────────────────────────────────────┘
        ▲                          ▲                         ▲
        │ реализуют (Phase 6)      │ реализуют (Phase 6)     │ реализуют (Phase 7)
   in-memory FAKE uow         mockery MOCK repo         Mongo impl (свап, D-02)
   in-memory FAKE outbox      mockery MOCK query        outbox-коллекция (Phase 7)
```

Ключевое: **форма usecase идентична Phase 7** — меняются только реализации портов, переданные в composition root. Это «эталон виден целиком уже сейчас» (D-02, Specifics).

### Recommended Project Structure

```text
services/inventory/internal/
├── domain/
│   ├── id.go               # типизированные ID-VO (HostID/ProjectID/DCID/ModuleID/RackID) + parse/zero
│   ├── aggregate.go        # встраиваемый aggregateBase: record()/PullEvents()/version/bumpVersion()
│   ├── events.go           # semantic event типы + DomainEvent интерфейс + EventEnvelope + Actor
│   ├── errors.go           # sentinel-ошибки: ErrFQDNConflict, ErrInvalidTransition, ErrProjectNotEmpty…
│   ├── host.go             # Host-агрегат: фабрика NewHost, Reassign/Relocate/ChangeHardware/Decommission/Delete
│   ├── hardware.go         # HostHardware VO + под-VO (RAM/CPU/Drive/NIC/PSU/StorageController/GPU/Chassis/Motherboard)
│   ├── project.go          # Project-агрегат + операции (Create/Rename/ChangeOwner/Delete)
│   ├── dc.go│module.go│rack.go  # три локационных агрегата + CRUD-методы
│   └── ports.go            # порты: UnitOfWork, Outbox, HostRepository, ProjectRepository, *LocationRepository,
│                           #         ActiveHostByFQDN (uniqueness query-порт), MatchAdvisor (хук), Clock, IDGenerator
├── usecases/
│   ├── register_host.go│decommission_host.go│delete_host.go│reassign_host.go│relocate_host.go│change_hardware.go
│   ├── create_project.go│delete_project.go│…
│   ├── crud_dc.go│crud_module.go│crud_rack.go
│   ├── envelope.go         # хелпер envelope-enrichment (PullEvents → EventEnvelope[]) на границе
│   └── mocks/              # сгенерированные mockery-моки портов (D-03)
└── (query/repositories/api/cron/app — скелет/фейки до Phase 7; SVC-01 раскладка)
```

Точную раскладку (один файл на агрегат vs группировка) определяет планнер (Claude's Discretion на имена пакетов). `query`-порт для uniqueness может жить в `domain/ports.go` (объявление) — реализация-фейк в тестах. **Generic-элементов для `pkg/` в Phase 6 не ожидается** — всё доменно-специфично для inventory (MEMORY: shared-code-must-live-in-pkg применяется только к по-настоящему generic коду; aggregateBase инвентори-специфичен).

### Pattern 1: Типизированный ID-VO over uuid.UUID (D-05)
**What:** Каждый агрегат имеет свой ID-тип-обёртку над `uuid.UUID` с фабрикой (генерация нового), parse (из string на edge/repo) и zero-value контролем. Компилятор ловит `Host.RackID ≠ ProjectID`.
**When to use:** Все 5 внутренних идентичностей (Host/Project/DC/Module/Rack). НЕ для внешних ID (Owner/INV/serial/MAC — raw string, D-05/HW-06).
**Example:**
```go
// Source: идиоматичный Go (паттерн typed-id over uuid) [ASSUMED — общий Go-паттерн];
// форма обёртки выбрана под D-05 «обёртка над uuid.UUID». Иллюстрация, не из репо.
package domain

import "github.com/google/uuid"

// HostID — внутренний носитель идентичности хоста (D-05). Непрозрачная обёртка над uuid.UUID;
// компилятор не даст перепутать с ProjectID/RackID.
type HostID struct{ v uuid.UUID }

// NewHostID генерирует новый постоянный непереиспользуемый ID (INV-03). v4 random.
func NewHostID() HostID { return HostID{v: uuid.New()} }

// ParseHostID восстанавливает ID из строки (для repo/edge в Phase 7); пустой/битый — ошибка.
func ParseHostID(s string) (HostID, error) {
    u, err := uuid.Parse(s)
    if err != nil {
        return HostID{}, fmt.Errorf("parse host id %q: %w", s, ErrInvalidID) // %w — style.md
    }
    return HostID{v: u}, nil
}

func (id HostID) String() string { return id.v.String() }
func (id HostID) IsZero() bool   { return id.v == uuid.Nil } // zero-value guard в инвариантах
```
**Note:** Альтернатива `type HostID string` (как `ExampleID` в example/provisioner.go) — проще, но теряет parse-валидацию. D-05 формулировка «обёртка над `uuid.UUID`» склоняет к struct-форме. Планнер фиксирует. Метод `String()` нужен для будущего Kafka-ключа (Phase 8, EVT-05: key = internal ID) и логов.

### Pattern 2: Встраиваемый aggregateBase + record() + PullEvents() (D-14)
**What:** Общий встраиваемый struct, дающий каждому агрегату накопление событий, слив через `PullEvents()` (отдаёт-и-очищает), и `version`-поле (D-15).
**When to use:** Все 5 агрегатов встраивают его.
**Example:**
```go
// Source: канон architecture.md §«Доменные события» + patterns.md Рецепт 3 (PullEvents).
// [CITED: knowledge/architecture.md, knowledge/patterns.md] Иллюстрация формы base.
package domain

// DomainEvent — голый семантический факт БЕЗ envelope-меты (D-14). Envelope навешивается
// в usecase. Метод EventType()/EntityID() нужны для enrichment на границе.
type DomainEvent interface {
    EventType() string   // напр. "HostRegistered" — ubiquitous language (D-13, DOC-07)
    EntityID() string    // строковая форма внутреннего ID сущности
}

// aggregateBase встраивается в каждый агрегат: копит события + несёт version.
type aggregateBase struct {
    version int           // поле агрегата (D-15); инкремент при каждом изменении; enforcement — Phase 7
    events  []DomainEvent // буфер накопленных фактов
}

// record добавляет факт в буфер и инкрементит version (вызывается из методов агрегата на переходе).
func (b *aggregateBase) record(e DomainEvent) {
    b.version++
    b.events = append(b.events, e)
}

// PullEvents отдаёт накопленные события и ОЧИЩАЕТ буфер (slice→nil), чтобы повторный pull был пуст.
func (b *aggregateBase) PullEvents() []DomainEvent {
    out := b.events
    b.events = nil
    return out
}

func (b *aggregateBase) Version() int { return b.version }
```
**Anti-pattern:** инкремент `version` вне `record()` (расходится с числом событий) — держать инкремент в одной точке.

### Pattern 3: Immutable HostHardware VO со структурированными под-компонентами (D-07, HW-01…06)
**What:** Единый неизменяемый VO. Все поля приватны, доступ через геттеры или через value-копию; изменение железа = собрать **новый** VO целиком. NIC/RAM/CPU/Drive/PSU/StorageController/GPU/Chassis — структурированные под-VO, не плоские слайсы примитивов.
**When to use:** Внутри Host. Изменение → `host.ChangeHardware(newHW)` → одно событие `HostHardwareChanged`.
**Example:**
```go
// Source: D-07 (immutable VO, NIC структурирован) + HW-01…06. Иллюстрация формы VO.
// [CITED: CONTEXT.md D-07] Все внешние идентификаторы — string (HW-06).
package domain

// HostHardware — единый immutable VO состава железа (HW-01). Сборка только через конструктор;
// поля приватны, мутации нет — изменение = новый VO целиком (D-07) → одно событие.
type HostHardware struct {
    name        string        // HW-01
    platform    string
    motherboard Motherboard   // паспорт материнки отдельно (HW-05)
    ipmiMAC     string        // HW-01 (raw string, HW-06)
    ram         []RAMModule   // HW-02 структурированные
    cpu         []CPU         // HW-02
    drives      []Drive       // HW-02
    nics        []NIC         // HW-03 структурированный, НЕ плоский MACs[]
    psus        []PSU         // HW-04
    storageCtl  []StorageController // HW-05 RAID/контроллер
    gpus        []GPU         // HW-05 внутренние GPU
    chassis     Chassis       // HW-05 шасси отдельно от материнки
}

// NIC — структурированный сетевой компонент (HW-03): модель/скорость/набор MAC'ов, не плоский MACs[].
type NIC struct {
    model    string
    speedGbE int
    macs     []string  // raw string (HW-06)
}

// NewHostHardware собирает immutable VO; место для простых инвариантов (напр. непустое name).
func NewHostHardware(in HardwareSpec) (HostHardware, error) { /* валидация + сборка */ }
```
**Note:** Слайсы внутри VO нужно копировать в конструкторе/геттерах (defensive copy), иначе вызывающий мутирует «immutable» VO через ссылку на слайс. Это известная Go-грабля immutable-VO (см. Pitfall ниже).

### Pattern 4: Envelope-enrichment на границе usecase (D-14, D-15, EVT-02)
**What:** Агрегат эмитит голые `DomainEvent`. После `PullEvents()` usecase оборачивает каждый в `EventEnvelope`, добавляя `eventId` (через `IDGenerator` порт), `occurredAt` (через `Clock` порт), `actor` (из параметра Execute — транспортная identity). Затем `outbox.Append(envelopes)`.
**When to use:** В каждом write-usecase, ровно между `PullEvents()` и `outbox.Append`.
**Example:**
```go
// Source: D-14/D-15 (envelope на границе, actor никогда не в агрегате) + Pitfall 6 (actor с 1-го дня).
// [CITED: CONTEXT.md D-14/D-15, PITFALLS Pitfall 6] Иллюстрация хелпера.
package usecases

// Actor — транспортная identity вызывающего (D-15). source: human|api|integration|system.
// В Phase 6 приходит параметром Execute (заглушка); в Phase 7 — из gRPC-interceptor (SVC-08).
type Actor struct {
    ID     string
    Source string // human|api|integration|system
}

// enrich навешивает envelope-мету на голые доменные факты на ГРАНИЦЕ usecase (домен чист).
func enrich(events []domain.DomainEvent, actor Actor, clock domain.Clock, idgen domain.IDGenerator) []domain.EventEnvelope {
    out := make([]domain.EventEnvelope, 0, len(events))
    for _, e := range events {
        out = append(out, domain.EventEnvelope{
            EventID:    idgen.NewID(),        // порт (детерминизм в тестах)
            EntityID:   e.EntityID(),
            EventType:  e.EventType(),
            Version:    /* версия агрегата на момент события */ 0,
            OccurredAt: clock.Now(),          // порт (детерминизм в тестах)
            Actor:      domain.Actor{ID: actor.ID, Source: actor.Source},
            Payload:    e,                    // голый семантический факт
        })
    }
    return out
}
```
**Critical:** `actor` НИКОГДА не входит в агрегат (D-14). `Clock`/`IDGenerator` — порты в `domain`, чтобы тесты были детерминированы (фиксированное время/ID). Это убирает недетерминизм из unit-тестов событий.

### Pattern 5: Interactor usecase через uow.Do + repo.Save + outbox.Append (D-01, D-02, Рецепт 1)
**What:** 1 usecase = 1 struct + `Execute`. Зависит только от портов. Внутри `uow.Do(ctx, fn)` — uniqueness-проверка → `repo.Save` → `PullEvents` → enrich → `outbox.Append`. Всё в одной «транзакции» (в Phase 6 фейк-uow просто зовёт fn).
**When to use:** Каждая write-операция.
**Example:**
```go
// Source: patterns.md Рецепт 1 + architecture.md §Write-side/UnitOfWork. [CITED: knowledge/patterns.md]
// Форма расширена под FQDN-check (D-11/Pitfall 7) и envelope (D-14). Иллюстрация.
package usecases

type RegisterHostUseCase struct {
    hosts  domain.HostRepository    // порт
    uniq   domain.ActiveHostByFQDN  // query-порт уникальности (D-11)
    outbox domain.Outbox            // порт
    uow    domain.UnitOfWork        // порт транзакционной границы
    clock  domain.Clock
    idgen  domain.IDGenerator
}

func (uc *RegisterHostUseCase) Execute(ctx context.Context, in RegisterHostInput, actor Actor) (RegisterHostOutput, error) {
    host, err := domain.NewHost(in.ProjectID, in.FQDN, in.Hardware, in.Lifecycle) // фабрика-инварианты (INV-02)
    if err != nil {
        return RegisterHostOutput{}, fmt.Errorf("register host: %w", err)
    }
    err = uc.uow.Do(ctx, func(ctx context.Context) error {
        // Pitfall 7: FQDN-конфликт среди active — доменный конфликт через query-порт, НЕ DB-error.
        if existing, found, err := uc.uniq.ActiveHostByFQDN(ctx, in.FQDN); err != nil {
            return err
        } else if found {
            cands, _ := /* MatchAdvisor no-op (D-12) */ nil, nil
            return domain.ErrFQDNConflict{FQDN: in.FQDN, ExistingID: existing, Candidates: cands}
        }
        if err := uc.hosts.Save(ctx, host); err != nil {
            return err
        }
        envelopes := enrich(host.PullEvents(), actor, uc.clock, uc.idgen) // граница (D-14)
        return uc.outbox.Append(ctx, envelopes)                            // та же tx (Pitfall 8)
    })
    if err != nil {
        return RegisterHostOutput{}, fmt.Errorf("register host: %w", err)
    }
    return RegisterHostOutput{ID: host.ID()}, nil
}
```
**Phase 7 свап:** меняется только composition root (`app`) — фейк-uow→Mongo-uow, mock-repo→Mongo-repo. `Execute` не трогается (D-02).

### Pattern 6: Порты + in-memory фейки (D-02) — форма под swap
**What:** Порты объявлены в `domain`. Фейки для uow/outbox в Phase 6 тривиальны.
**Example:**
```go
// Source: architecture.md §UnitOfWork/события + D-02 (фейки). [CITED: knowledge/architecture.md]
package domain

type UnitOfWork interface {
    Do(ctx context.Context, fn func(ctx context.Context) error) error
}
type Outbox interface {
    Append(ctx context.Context, events []EventEnvelope) error
}
type HostRepository interface {
    Save(ctx context.Context, h *Host) error
    Delete(ctx context.Context, id HostID) error   // INV-07: физическое удаление (D-09), не state=deleted
}
// ActiveHostByFQDN — query-порт уникальности (D-11). Read-side hook в write-usecase.
type ActiveHostByFQDN interface {
    ActiveHostByFQDN(ctx context.Context, fqdn string) (existing HostID, found bool, err error)
}
// MatchAdvisor — порт-хук советочного матчинга (D-12, INV-08). В Phase 6 — no-op заглушка.
type MatchAdvisor interface {
    Candidates(ctx context.Context, fqdn string, hw HostHardware) ([]HostID, error)
}

// --- фейки (Phase 6; в тестах или internal/usecases/fakes) ---
// fakeUoW просто выполняет fn без транзакции (D-02: in-memory uow = просто зовёт fn).
type fakeUoW struct{}
func (fakeUoW) Do(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }
// fakeOutbox копит envelopes в слайс — тест проверяет, ЧТО эмитнуто (D-02: in-memory outbox = слайс).
type fakeOutbox struct{ appended []EventEnvelope }
func (o *fakeOutbox) Append(_ context.Context, evs []EventEnvelope) error {
    o.appended = append(o.appended, evs...); return nil
}
```
**Note (D-03 vs фейки):** repo/uniqueness/MatchAdvisor → mockery-моки (expectation на вызовы). uow/outbox → in-memory фейки (нужно реально гонять fn и собрать события для проверки итога). Это не противоречит testing.md: тривиальные uow/outbox-фейки — тестовый хелпер, поведение которого детерминировано и не дрейфует. Планнер фиксирует баланс; обе формы валидны.

### Anti-Patterns to Avoid
- **`HostUpdated`-дамп-событие** на любое изменение (Pitfall 5, D-13): каждое изменение → своё семантическое событие. WON'T один тип на все мутации.
- **`actor`/`Clock`/`IDGen` внутри агрегата** (D-14): домен-ядро остаётся чистым; envelope-мета — на границе usecase. WON'T инъекция Clock в фабрику.
- **`state=deleted` как член enum** (D-09, Pitfall 2): `deleted` — hard-delete (`repo.Delete` + событие), НЕ lifecycle-состояние. WON'T строка-tombstone в живой коллекции.
- **FQDN-конфликт как сырой DB-error** (D-11, Pitfall 7): типизированный `ErrFQDNConflict` через query-порт. WON'T полагаться на E11000.
- **outbox.Append вне uow.Do** (Pitfall 8): даже на фейках держать форму «внутри fn», чтобы Phase 7 был свапом. WON'T публикация после fn.
- **Воскрешение из decommissioned** (D-10): `decommissioned` терминально; re-add = новый ID. WON'T авто-restore/merge.
- **Мутабельные под-компоненты HostHardware с per-компонент событиями** (D-07): один immutable VO, одно событие `HostHardwareChanged`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Генерация уникального ID | Свой счётчик/случайная строка | `github.com/google/uuid` (`uuid.New()`) | Коллизии, переиспользование (нарушит INV-03); uuid уже в go.mod |
| Парсинг/валидация ID из строки | Ручной regex | `uuid.Parse` | RFC-валидация из коробки; ошибки — обёрнуть в доменный sentinel |
| Тест-дублёры портов (repo/uniq/advisor) | Ручные fake-структуры с поведением | `mockery v3` (`make generate-mocks`) | testing.md: ручной фейк дрейфует от интерфейса; D-03 — только mockery-моки |
| BDD-структура/ассерты тестов | Свой раннер / `t.Fatalf` | Ginkgo v2 + Gomega | testing.md канон; уже провязано |
| Sentinel-ошибки + цепочка | `errors.New` в месте + строковое сравнение | `var Err… = errors.New(...)` + `errors.Is` + `%w` | style.md (errorlint hook); типизированный конфликт сравним |
| Детерминизм времени/ID в тестах | `time.Now()`/`uuid.New()` прямо в коде | Порты `Clock`/`IDGenerator` с фейк-реализацией | Недетерминированные unit-тесты событий; D-14 уже выносит мету на границу |

**Key insight:** В Phase 6 «не хэндроллить» относится прежде всего к **тест-дублёрам** (mockery, D-03) и **детерминизму** (Clock/IDGen-порты). Сам домен (агрегаты/VO/инварианты) пишется руками — это и есть продукт фазы; но он должен быть **чистым** (без I/O, без time/uuid-сайд-эффектов внутри), чтобы тестироваться прямыми unit-тестами без моков (D-03).

## Runtime State Inventory

> Phase 6 — **greenfield доменный код** (чистый лист `internal/` после v1.0; только throwaway example-пакет). Нет rename/refactor/migration существующих данных. Раздел применим минимально — фиксируем затронутые артефакты Phase 5.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — нет persistence в Phase 6 (фейки in-memory); Mongo — Phase 7. | none |
| Live service config | None — нет запущенных сервисов/топиков, затрагиваемых доменным кодом. | none |
| OS-registered state | None. | none |
| Secrets/env vars | None — домен не читает env/секреты. | none |
| Build artifacts | `internal/example/` (throwaway smoke-пакет для mockery) — остаётся как есть; `.mockery.yaml` будет **расширен** реальными доменными портами (добавятся packages-записи). `google/uuid` indirect→direct после первого импорта + `go mod tidy`. | Расширить `.mockery.yaml` (добавить domain-порты); `go mod tidy` после импорта uuid; example-пакет можно оставить или удалить по решению планнера (он доказал smoke — больше не нужен, но не мешает). |

**Nothing found in categories Stored/Live/OS/Secrets:** verified — Phase 6 не трогает runtime-состояние (D-01/D-02: фейки, без инфраструктуры).

## Common Pitfalls

### Pitfall 1: `deleted` смоделирован как lifecycle-состояние (нарушение D-09/Pitfall 2)
**What goes wrong:** Соблазн добавить `deleted` четвёртым членом `lifecycleState` enum и хранить «удалённые» строки с этим флагом (soft-delete).
**Why it happens:** Привычка к soft-delete; кажется симметричным с decommissioned.
**How to avoid:** `lifecycleState ∈ {shadow, registered, decommissioned}` — ровно три. `deleted` = агрегатный метод `Delete()` → событие `HostDeleted` (полный payload + actor) + usecase зовёт `repo.Delete` (физическое удаление). История живёт **только** в append-only `*.events` (Phase 8). Никакой строки `state=deleted`.
**Warning signs:** enum с 4 членами; `repo.Save` с флагом deleted вместо `repo.Delete`; тест ищет «удалённый» хост в live-store и находит.

### Pitfall 2: «Immutable» HostHardware протекает через слайс-ссылку
**What goes wrong:** VO с приватным слайсом `nics []NIC`, но конструктор сохраняет переданный слайс по ссылке, или геттер возвращает внутренний слайс — вызывающий мутирует «immutable» VO.
**Why it happens:** Go-слайсы — reference-типы; «приватное поле» не равно «immutable».
**How to avoid:** Defensive copy в конструкторе (`append([]NIC(nil), in...)`) и в геттерах, возвращающих слайсы. Либо не отдавать слайсы наружу, а только агрегированные значения. Тест: получить слайс из VO, мутировать, проверить что VO не изменился.
**Warning signs:** геттер `NICs() []NIC { return hw.nics }` без копии; общий слайс между двумя VO.

### Pitfall 3: `version` и число событий разъезжаются
**What goes wrong:** `version` инкрементится в разных местах (часть в фабрике, часть в методах), не совпадает с числом `record()`-вызовов; в Phase 7 optimistic-concurrency сломается.
**How to avoid:** Инкремент `version` — **только** внутри `record()` (Pattern 2). Один путь. Тест: после N операций `Version() == N` и `len(PullEvents()) == N` (до pull).
**Warning signs:** `b.version++` вне `record()`; событие эмитнуто без инкремента version.

### Pitfall 4: FQDN-конфликт-проверка вне транзакции / не через порт (Pitfall 7)
**What goes wrong:** Проверка `ActiveHostByFQDN` сделана до `uow.Do` (вне «транзакции») или конфликт возвращается как `errors.New("duplicate")` без типа — теряется контекст для human-in-the-loop и advisory-кандидатов.
**How to avoid:** Проверка — **внутри** `uow.Do(fn)` (форма под Phase 7, где это будет в той же tx + defense-in-depth partial index). Конфликт — типизированный `ErrFQDNConflict{FQDN, ExistingID, Candidates}` (Candidates от MatchAdvisor no-op в Phase 6). Сравним через `errors.As`.
**Warning signs:** uniqueness-check перед `uow.Do`; сырой error без полей; нет `MatchAdvisor`-хука в пути.

### Pitfall 5: `PullEvents()` не очищает буфер
**What goes wrong:** `PullEvents()` возвращает слайс, но не обнуляет буфер — повторный вызов/повторное сохранение дублирует события в outbox.
**How to avoid:** `PullEvents()` отдаёт-и-очищает (`b.events = nil`). Тест: второй `PullEvents()` подряд возвращает пустой слайс.
**Warning signs:** `return b.events` без обнуления; дубли в fakeOutbox.appended.

### Pitfall 6: Недетерминированные тесты событий (time.Now/uuid в коде)
**What goes wrong:** `occurredAt = time.Now()` или `eventId = uuid.New()` прямо в usecase/агрегате → тесты не могут ассертить точные значения envelope, флакают.
**How to avoid:** `Clock`/`IDGenerator` — порты в `domain`; в тестах — фейк с фиксированным временем/последовательностью ID. D-14 уже выносит эту мету на границу usecase — там она инъектируется через порты.
**Warning signs:** прямой `time.Now()`/`uuid.New()` в usecase; тест envelope с `Expect(...).ToNot(BeEmpty())` вместо точного значения.

## Code Examples

### Lifecycle state-machine (D-08, D-10, INV-04/06)
```go
// Source: D-08/D-10 (3 состояния + терминальный decommission, гибкий вход). Иллюстрация.
// [CITED: CONTEXT.md D-08/D-10]
type lifecycleState int
const (
    stateShadow lifecycleState = iota  // заготовка/обнаружен
    stateRegistered                    // ручная полная регистрация
    stateDecommissioned                // терминально (D-10: нет воскрешения)
)

// Decommission списывает хост (INV-06): из shadow или registered; терминально.
func (h *Host) Decommission(reason string) error {
    if h.state == stateDecommissioned {
        return ErrAlreadyDecommissioned  // sentinel
    }
    h.state = stateDecommissioned
    h.record(HostDecommissioned{ID: h.id, Reason: reason}) // version++ внутри record
    return nil
}

// Delete — hard-удаление из ЛЮБОГО состояния (D-09): эмитит факт, usecase зовёт repo.Delete.
func (h *Host) Delete() {
    h.record(HostDeleted{ID: h.id, Snapshot: h.snapshot()}) // полный payload для аудита (Pitfall 2)
}
```

### Project: delete-only-if-empty инвариант (INV-10)
```go
// Source: INV-10 (Project удалить только пустым). Проверка — в usecase через query-порт
// (D-11 паттерн: инвариант, требующий чтения хранилища → query-порт). [CITED: CONTEXT.md INV-10]
func (uc *DeleteProjectUseCase) Execute(ctx context.Context, id domain.ProjectID, actor Actor) error {
    return uc.uow.Do(ctx, func(ctx context.Context) error {
        n, err := uc.hostCount.HostsInProject(ctx, id) // query-порт
        if err != nil { return err }
        if n > 0 { return domain.ErrProjectNotEmpty{ProjectID: id, HostCount: n} }
        proj, err := uc.projects.Load(ctx, id) // load для эмиссии события + Delete
        if err != nil { return err }
        proj.Delete()
        if err := uc.projects.Delete(ctx, id); err != nil { return err }
        return uc.outbox.Append(ctx, enrich(proj.PullEvents(), actor, uc.clock, uc.idgen))
    })
}
```

### Семантические имена событий (D-13, фиксируются в DOC-07)
```go
// Source: D-13 + ROADMAP SC5. Одно событие на операцию; имена = ubiquitous language.
// [CITED: CONTEXT.md D-13, ROADMAP Phase 6 SC5]
// Host:    HostRegistered, HostHardwareChanged, HostReassigned, HostRelocated,
//          HostDecommissioned, HostDeleted
// Project: ProjectCreated, ProjectRenamed, ProjectOwnerChanged, ProjectDeleted
// Loc:     DCCreated/DCUpdated/DCDeleted, ModuleCreated/…, RackCreated/RackUpdated/RackDeleted
// (точный набор Project/Loc-операций — Claude's Discretion, в рамках D-13)
```

### DOC-07 glossary — структура (authoring.md MUST/SHOULD/WON'T)
```markdown
<!-- Source: authoring.md (стандарт), DOC-07, boundaries.md (glossary — исключение из no-phantom).
     [CITED: knowledge/authoring.md §«Никаких phantom-правил» — glossary единственное исключение] -->
# Глоссарий домена Inventory
## Граница домена
- **Факт существования** (Inventory владеет): идентичность, ЖЦ-как-актив, состав железа, локация, топология.
- **Динамическое состояние** (НЕ Inventory → State/Health): runtime-health, failed/maintenance, firmware.
## Термины
- **Project** — …  | **Host** — …  | **Owner** — непрозрачный внешний string-ID группы (INV-09)
- **Module / Rack / DC** — …  | **Connection** — (Phase 9; упоминается как forward-term)
- **Идентичность** — внутренний постоянный непереиспользуемый ID; единственный носитель (INV-03)
- **decommission ≠ delete** — decommission = смена ЖЦ (хост видим, терминально); delete = hard-удаление (история в *.events)
## Семантические события (D-13)
- HostRegistered / HostHardwareChanged / HostReassigned / HostRelocated / HostDecommissioned / HostDeleted / …
```
**Authoring-note:** glossary — единственное место, где разрешено фиксировать термины «вперёд» (boundaries.md/authoring.md: glossary — исключение из no-phantom). Connection (Phase 9) можно упомянуть как forward-term. Файл идёт в `knowledge/glossary.md`, индексируется в `knowledge/README.md` (статус «существует» вместо «отложено»).

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `type ID string` голый | Типизированный ID-VO over uuid (struct/named type) | канон style.md | Компилятор ловит перепутанные ссылки (D-05) |
| Анемичная модель (сеттеры) | Rich aggregate с фабрикой + методами-переходами + инвариантами | DDD-канон architecture.md | Инварианты в одном месте, события рождаются в агрегате |
| `EntityUpdated`-дамп | Семантические гранулярные события | L2-инвариант 5 / D-13 | Консьюмеры понимают, что изменилось; нет жирных дампов (Pitfall 5) |
| mockery v2 (`.PackageName`) | mockery v3 (`.SrcPackageName`, template testify, моки рядом) | Phase 5 (.mockery.yaml) | Использовать v3-синтаксис; не v2 |
| Clock/ID внутри домена | Clock/IDGen-порты, envelope на границе | D-14 | Чистое ядро, детерминированные тесты |

**Deprecated/outdated:**
- mockery v2 синтаксис — НЕ использовать; `.mockery.yaml` уже на v3 (pinned v3.7.1).
- Soft-delete (`state=deleted`-строка) — by design запрещён (D-09); hard-delete + append-only лог.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | ID-VO как `struct{ v uuid.UUID }` (не `type ID string`) предпочтительнее под D-05 «обёртка над uuid.UUID» | Pattern 1, Standard Stack | LOW — обе формы рабочие; планнер фиксирует. `type ID string` (как ExampleID) проще, но теряет parse-валидацию uuid |
| A2 | `uuid.New()` (v4) достаточен; v7 не требуется в v3.0 | Standard Stack, Pattern 1 | LOW — v3.0 не требует сортируемости ID; v7 — опциональное улучшение для будущих Kafka/Mongo-ключей |
| A3 | `aggregateBase` встраивается (не выносится в pkg) — он inventory-специфичен, не generic | Standard Stack, Structure | LOW — MEMORY-правило «shared→pkg» относится к truly-generic; доменный base инвентори-специфичен. Если позже analytics/audit заведут агрегаты — можно вынести generic-часть в pkg (будущее) |
| A4 | uow/outbox — in-memory фейки, repo/uniq/advisor — mockery-моки (смешанный подход) | Pattern 6, Alternatives | LOW — D-02 требует фейки (uow=зовёт fn, outbox=слайс), D-03 — моки портов; баланс на планнере. Обе формы соответствуют канону |
| A5 | `Clock`/`IDGenerator` — отдельные порты в domain для детерминизма envelope-тестов | Pattern 4, Pitfall 6 | LOW — альтернатива: передавать `now`/`eventID` параметрами в enrich. Порты чище для тестов; планнер выбирает |
| A6 | example-пакет (`internal/example`) можно удалить после Phase 6 (smoke больше не нужен) | Runtime State Inventory | LOW — он не мешает; удаление опционально. Реальные порты заменяют его как mockery-цель |
| A7 | Точный набор Project/Loc-операций и событий (rename/changeOwner/update) — Claude's Discretion | Code Examples (события) | LOW — CONTEXT.md явно отдаёт это планнеру в рамках D-13 |

**Все остальные доменные решения — НЕ assumptions:** они локнуты в CONTEXT.md D-01…D-15 и/или процитированы из knowledge/ канона и PITFALLS.md (CITED).

## Open Questions

1. **Форма ID-VO: `struct{ uuid.UUID }` vs `type ID string`?**
   - What we know: D-05 говорит «обёртка над `uuid.UUID` с фабрикой/парсингом»; style.md плейсхолдер показывает `type OrderID string`; example-код использует `type ExampleID string`.
   - What's unclear: буквальная форма обёртки.
   - Recommendation: `struct{ v uuid.UUID }` (строже под D-05 формулировку, parse-валидация); но `type HostID = uuid.UUID`-стиль тоже приемлем. Планнер фиксирует одну форму для всех 5 ID единообразно. Низкий риск — компиляция и тесты одинаковы.

2. **Инвариант уникальности позиции host↔rack (два хоста в одном юните)?**
   - What we know: CONTEXT.md Claude's Discretion отдаёт это планнеру (в рамках D-11).
   - What's unclear: вводить ли его в Phase 6 или отложить.
   - Recommendation: смоделировать как query-порт `HostAtRackPosition` + доменный конфликт `ErrPositionOccupied` по той же форме, что FQDN-конфликт (D-11). Дёшево и консистентно. Планнер решает включение.

3. **Удалять ли `internal/example` smoke-пакет?**
   - Recommendation: удалить после того, как реальные порты станут mockery-целями (он выполнил роль доказательства). Опционально — оставить не мешает. Решение планнера.

## Environment Availability

> Phase 6 — чистый Go-домен на уже установленном toolchain/стеке. Внешних сервисов (Mongo/Kafka/Docker) код Phase 6 НЕ требует (фейки in-memory).

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | компиляция/тесты домена | ✓ | go1.25.0 | — |
| `github.com/google/uuid` | ID-VO (INV-03/D-05) | ✓ | v1.6.0 (в go.mod indirect) | — |
| Ginkgo v2 / Gomega | unit-спеки | ✓ | v2.32.0 / v1.42.1 | — |
| mockery | кодоген моков портов (D-03) | ✓ | v3.7.1 (pinned Makefile) | — |
| `make generate-mocks` | регенерация моков | ✓ | — | ручной запуск `mockery` |
| Docker / Mongo / Kafka | НЕ требуются в Phase 6 | n/a | — | фейки in-memory (D-02) |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** none — всё провязано в Phase 5.

## Validation Architecture

> `workflow.nyquist_validation` = true (config.json). Раздел включён.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Ginkgo v2 (v2.32.0) + Gomega (v1.42.1); mockery v3 (v3.7.1) для моков портов |
| Config file | `services/inventory/.mockery.yaml` (будет расширен domain-портами); suite-бутстрап в каждом `*_test.go` |
| Quick run command | `cd services/inventory && go test ./internal/domain/... ./internal/usecases/...` |
| Full suite command | `go test ./...` из корня workspace (pre-push гейт inventory unit) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| INV-01 | Создать Project → ID присвоен, ProjectCreated эмитнут | usecase (фейк uow/outbox) | `go test ./internal/usecases/... -run CreateProject` | ❌ Wave 0 |
| INV-02 | RegisterHost требует ProjectID (фабрика-инвариант) | unit (агрегат) | `go test ./internal/domain/... -run NewHost` | ❌ Wave 0 |
| INV-03 | ID непустой, уникальный, типизированный | unit (ID-VO) | `go test ./internal/domain/... -run HostID` | ❌ Wave 0 |
| INV-04 | ЖЦ переходы shadow→registered→decommissioned валидны/невалидны | unit (state-machine) DescribeTable | `go test ./internal/domain/... -run Lifecycle` | ❌ Wave 0 |
| INV-05 | Reassign меняет ProjectID + HostReassigned | unit (агрегат) + usecase | `-run Reassign` | ❌ Wave 0 |
| INV-06 | Decommission терминален; повторный — ошибка; HostDecommissioned | unit (агрегат) | `-run Decommission` | ❌ Wave 0 |
| INV-07 | Delete эмитит HostDeleted + usecase зовёт repo.Delete (mock expectation); НЕ state=deleted | usecase (mock repo) | `-run DeleteHost` | ❌ Wave 0 |
| INV-08 | Re-add = новый ID; MatchAdvisor вызван (no-op кандидаты) | usecase (mock advisor) | `-run ReAdd` | ❌ Wave 0 |
| INV-09 | Owner — raw string, хранится как есть | unit (агрегат) | `-run Owner` | ❌ Wave 0 |
| INV-10 | FQDN-конфликт среди active → ErrFQDNConflict (не DB-error); Project delete-only-empty | usecase (mock uniq/hostCount) | `-run FQDNConflict`, `-run DeleteProject` | ❌ Wave 0 |
| HW-01…06 | HostHardware VO структура; NIC структурирован; immutability (defensive copy); внешние ID = string | unit (VO) DescribeTable | `go test ./internal/domain/... -run Hardware` | ❌ Wave 0 |
| LOC-01…04 | DC/Module/Rack CRUD; иерархия по ID; Host→RackID+позиция; Rack-атрибуты | unit + usecase | `-run Location` | ❌ Wave 0 |
| EVT-01 | Каждая операция эмитит ровно одно семантическое событие | unit (агрегат) + usecase (fakeOutbox.appended) | `-run Event` | ❌ Wave 0 |
| EVT-02 | Envelope несёт eventId/version/actor/occurredAt; детерминизм через Clock/IDGen фейки | usecase (fake clock/idgen) | `-run Envelope` | ❌ Wave 0 |
| DOC-07 | glossary.md существует, содержит границу/термины/decommission≠delete/имена событий | manual review + grep | `grep -q "decommission" knowledge/glossary.md` | ❌ Wave 0 |
| SVC-01 | Код в канон-слоях; domain не импортирует наружу | structural (компиляция + ревью) + depguard (dormant→active) | `go build ./... && go vet ./...` | partial (depguard написан, активируется появлением кода) |

### Sampling Rate
- **Per task commit:** `cd services/inventory && go test ./internal/domain/... ./internal/usecases/...` (быстрые unit/usecase, без контейнеров)
- **Per wave merge:** `go test ./...` из workspace-корня (полный inventory unit)
- **Phase gate:** полный suite зелёный + `go vet ./...` + glossary.md present перед `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/domain/*_test.go` — спеки агрегатов/VO/ID/lifecycle (прямые unit, без моков; D-03)
- [ ] `internal/usecases/*_test.go` — спеки interactor'ов (mockery-моки repo/uniq/advisor + in-memory fake uow/outbox/clock/idgen; D-02/D-03)
- [ ] Расширить `.mockery.yaml`: добавить packages-записи на реальные доменные порты (HostRepository/ProjectRepository/Outbox/UnitOfWork/ActiveHostByFQDN/MatchAdvisor/…), затем `make generate-mocks`
- [ ] `internal/usecases/fakes` (или inline в `_test.go`) — fakeUoW/fakeOutbox/fakeClock/fakeIDGen хелперы
- [ ] Каждый suite-файл: `TestXxxSuite` бутстрап (`RegisterFailHandler(Fail)` + `RunSpecs`) — testing.md канон
- [ ] `knowledge/glossary.md` — DOC-07 (не тест, но gate-артефакт; ревью + grep-проверка)

*Framework install: не требуется — Ginkgo/Gomega/mockery провязаны в Phase 5.*

## Security Domain

> `security_enforcement` = true, ASVS level 1. Phase 6 — чистый доменный код без транспорта/persistence/внешнего ввода. Большинство ASVS-категорий не применимы на этом слое (они лягут на Phase 7 gRPC-edge). Релевантна валидация доменных инвариантов и forward-compat атрибуции (actor).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | Транспорт/identity — Phase 7 (gRPC interceptor, SVC-08) |
| V3 Session Management | no | Нет сессий в домене |
| V4 Access Control | no (forward) | Проверки прав — будущий Access-домен; в Phase 6 actor пробрасывается параметром (stub, SEED-003) |
| V5 Input Validation | yes | Фабрики/конструкторы агрегатов и VO валидируют доменные инварианты (ProjectID обязателен, ЖЦ-переходы, непустые поля). Транспортная валидация — на edge (Phase 7, style.md) |
| V6 Cryptography | no | Нет крипто в домене; ID — uuid v4 (не секрет) |
| V7 Error Handling | yes | Sentinel-ошибки + `%w` (style.md/errorlint); типизированные доменные конфликты (ErrFQDNConflict) не протекают реализацией БД наружу (Pitfall 7) |
| V8 Logging/Audit | yes (forward) | `actor{id,source}` в envelope с первого дня (EVT-02/D-15, SEED-002) — иначе исторические события не атрибутируемы (Pitfall 6) |

### Known Threat Patterns for Go DDD-домен

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Невалидный переход ЖЦ (воскрешение decommissioned) | Tampering | State-machine инвариант в агрегате; терминальный decommissioned (D-10) |
| Утечка реализации БД через сырой конфликт (E11000 наружу) | Information Disclosure | Типизированный `ErrFQDNConflict` через query-порт (Pitfall 7) |
| Событие без `actor` → неатрибутируемое действие | Repudiation | `actor` в envelope с 1-го дня (EVT-02, Pitfall 6, SEED-002) |
| Авто-restore/merge вернувшегося хоста (ложная атрибуция данных) | Spoofing/Tampering | Re-add = новый ID, авто-мердж запрещён by design (D-12, INV-08) |
| Мутация «immutable» HostHardware через слайс-ссылку | Tampering | Defensive copy в VO-конструкторе/геттерах (Common Pitfall 2) |
| Потеря аудит-следа удаления (state=deleted затирается) | Repudiation | HostDeleted с полным payload в append-only лог; hard-delete, не soft (D-09, Pitfall 2) |

**Security note:** В Phase 6 security сводится к **корректности доменных инвариантов** (V5) и **forward-compat атрибуции** (V8/actor). Транспортные контроли (authn/authz/rate-limit) — Phase 7 edge. `security_block_on: high` — высокорисковых внешних поверхностей в Phase 6 нет.

## Sources

### Primary (HIGH confidence)
- `.planning/phases/06-inventory/06-CONTEXT.md` — D-01…D-15, scope-fences, canonical refs (локнутые решения)
- `.planning/research/PITFALLS.md` — Pitfalls 2/5/6/7/8 + «Looks Done But Isn't» чеклист (приёмочные точки)
- `knowledge/architecture.md` — канон слоёв, Execute, UnitOfWork, PullEvents, outbox, MUST NOT CQRS
- `knowledge/patterns.md` — Рецепты 1/3 (use case / aggregate+PullEvents)
- `knowledge/style.md` — типизированные ID, sentinel+`%w`, маппинг DTO→домен на edge
- `knowledge/testing.md` — Ginkgo v2 + Gomega + mockery v3 конвенции, спек-структура
- `knowledge/authoring.md` — MUST/SHOULD/WON'T, glossary = исключение из no-phantom
- `knowledge/boundaries.md` / `structure.md` — карта владения, inventory как компилируемый член go.work
- `.planning/ROADMAP.md` § Phase 6/7/8 — Success Criteria + scope-границы
- `.planning/REQUIREMENTS.md` — тексты INV/HW/LOC/EVT/DOC/SVC
- `services/inventory/internal/example/` (provisioner.go, _test.go, mocks/) — реальный mockery v3 expecter-shape + Ginkgo-бутстрап (компилируемый эталон)
- `services/inventory/.mockery.yaml`, `Makefile` (mockery v3.7.1 pinned) — провязка кодогена
- `services/inventory/go.mod` + `go list -m github.com/google/uuid` (v1.6.0) — верификация зависимостей

### Secondary (MEDIUM confidence)
- `.planning/L2-ARCHITECTURE.md` — инвариант семантических событий (5), single identity owner

### Tertiary (LOW confidence)
- Общие идиоматичные Go-паттерны (typed-id over uuid, embedded aggregate-base, defensive-copy для immutable VO) — `[ASSUMED]` из training knowledge, отмечены в Assumptions Log A1–A5. Форма выбрана консистентно с существующим кодом репо и каноном knowledge/.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — нулевая новизна (uuid уже в go.mod, тест-стек из Phase 5), верифицировано `go list -m`
- Architecture: HIGH — формы прямо процитированы из architecture.md/patterns.md и существующего кода; решения локнуты D-01…D-15
- Pitfalls: HIGH — из PITFALLS.md (HIGH-confidence research) + идиоматичные Go-грабли immutable-VO/PullEvents
- Идиоматичные Go-формы (ID-VO struct, aggregate-base): MEDIUM/LOW — `[ASSUMED]`, отмечены в Assumptions Log; обе альтернативы рабочие, планнер фиксирует

**Research date:** 2026-06-30
**Valid until:** 2026-07-30 (стабильный домен; внешних быстро-движущихся зависимостей нет — паттерны и канон стабильны)
