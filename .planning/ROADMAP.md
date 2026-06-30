# Roadmap: gwall-e

## Milestones

- ✅ **v1.0 Фундамент** — Phases 1-4 (shipped 2026-06-17) — база знаний `knowledge/` + enforcement-тулинг
- 🗺️ **v2.0 L2-видение** — карта 12 доменов + направление (без кода); см. [L2-ARCHITECTURE.md](L2-ARCHITECTURE.md). Эпики режутся отдельными milestone'ами
- 🚧 **v3.0 Inventory + Event-backbone** — Phases 5-10 (планирование) — первый реальный сервис (E1): домен Inventory (solo) + outbox→relay→Kafka продюсер

Full v1.0 detail archived in [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md).

## Phases

<details>
<summary>✅ v1.0 Фундамент (Phases 1-4) — SHIPPED 2026-06-17</summary>

- [x] Phase 1: Раскладка базы знаний и точки входа (2/2 plans) — completed 2026-06-17
- [x] Phase 2: Стабильные доки-основы (3/3 plans) — completed 2026-06-17
- [x] Phase 3: Доки конвенций и архитектуры (5/5 plans) — completed 2026-06-17
- [x] Phase 4: Enforcement-слой (тулинг) (4/4 plans) — completed 2026-06-17

**Resolved (Phase 5, Plan 05-05):** DOC-02 — `build.md` audit-рецепт переведён на `cd services/audit && go vet ./...` (exit 0, проверено эмпирически); падающий `go build ./...` снят. См. [MILESTONES.md](MILESTONES.md).

</details>

### 🗺️ v2.0 — L2-видение платформы (epic-уровень)

> **Тип milestone:** L2-видение (без кода). Это НЕ нарезка фаз — это карта эпиков и направление.
> Каждый эпик = отдельный будущий milestone, который режется через `/gsd-new-milestone`.
> Полная архитектура — [L2-ARCHITECTURE.md](L2-ARCHITECTURE.md). Порядок индикативный,
> уточняем по ходу.

**Принцип:** сервис = домен (DDD). Идентичность Project/Host/VM владеет Inventory; домены
синхронизируются через Kafka (choreography для идентичности + orchestration-саги для длинных
процессов).

**Эпики (each → отдельный milestone):**

| # | Эпик (домен) | Цель (L2) | Зависит от |
|---|--------------|-----------|------------|
| E1 | **Inventory + Event-backbone** | Источник истины Project/Host/VM (идентичность/ЖЦ, железо, локация, owner-ref); solo/sync-инвентарь; outbox→relay→**Kafka** | — *(фундамент)* |
| E2 | **Access** | Вся авторизация: owner-роли, права/роли, временные гранты, IDM-sync, SSH-гранты | E1 |
| E3 | **Network** | Свитчи, VLAN, IPAM, сетевые шаблоны, смена VLAN | E1 |
| E4 | **Health / Monitoring** | Runtime, health-checks, config-compliance (+ ingest от Host Agent) | E1 |
| E5 | **Coordination** ⭐ | approve-before-start с Owner CMS, CMS-конфиг+вызов, локи, предохранители/лимиты | E1, E2 |
| E6 | **Integrations** | Адаптеры к внешним провайдерам операций (profile/reimage/network/reboot) | E1 |
| E7 | **Actions** | Единичные операции над хостом + каталог наливок | E5, E6, E3 |
| E8 | **Orchestrator** | Кросс-доменные lifecycle-саги (provision, decommission с вето) | E1, E5, E7 |
| E9 | **Scenarios** | Кампании плановых массовых работ (окна, drain, shutdown на учения, move owner) | E7, E5, E8 |
| E10 | **Remediation** | Авто-починка по правилам SRE (Automation Plot), self-healing | E4, E7 |
| E11 | **Audit** | Лог всех действий в системе (consumer событий) | E1 |
| E12 | **Analytics** | Аналитика парка | E1 |
| E13 | **Search** | Поверх OpenSearch, много индексов, event-fed; для людей и машин | E1 |
| E14 | **API Gateway / BFF** | Вход для фронтенда | по мере доменов |
| E15 | **Notifications** | Уведомления | по мере доменов |

**Отложенный кластер:** **Host Agent** (агент на хосте: сбор данных → Health, исполнение ← Actions,
раздача SSH ← Access; + серверная часть). Решение «домен vs часть Health» — при проектировании E2/E4.

**Индикативный порядок:** `E1` → (`E2` ∥ `E3` ∥ `E4`) → `E5` → (`E6` → `E7`) → `E8` → `E9` → `E10`;
платформенные (`E11`–`E15`) вплетаются по мере появления данных/фронта.

**Tech debt v1.0** (внести при нарезке E1): DOC-02 fix, Nyquist sign-off, live-firing UAT, DOC-07 glossary.

> Нарезка E1 → milestone **v3.0** ниже.

### 🚧 v3.0 — Inventory + Event-backbone (Phases 5-10)

> **Тип milestone:** первый реальный сервис (нарезка эпика E1). Эталонная реализация архитектуры
> v1.0: домен Inventory (solo) + продюсер-бэкбон событий на Kafka (outbox→relay→Kafka,
> dual-topic per aggregate). Консьюмеров нет (продюсер-only). Контекст — [PROJECT.md](PROJECT.md);
> требования — [REQUIREMENTS.md](REQUIREMENTS.md); ресёрч — [research/SUMMARY.md](research/SUMMARY.md).

**Build order (диктуется зависимостями):** инфра/стек → доменная модель + glossary → эталон
записи (UoW+Outbox+gRPC) → схемы + relay→Kafka → топология connections → верификация backbone.
**Жёсткий инвариант:** forward-compat envelope (`eventId`/`version`/`actor`, SEED-002/003)
заложен в доменные события (Phase 6) и фиксируется в protobuf-схеме (Phase 8) **до** relay-кода —
переэмит immutable-лога задним числом невозможен.

- [ ] **Phase 5: Dev-инфра и стек** — mongo-driver/v2 миграция, docker-compose (Kafka KRaft + Mongo RS), provisioning топиков, тест-тулинг; DOC-02 fix
- [ ] **Phase 6: Доменная модель Inventory** — агрегаты Project/Host (+ HW VO, локации DC·Module·Rack), идентичность/ЖЦ, семантические доменные события с envelope, glossary DOC-07
- [ ] **Phase 7: Эталон записи и чтения** — UnitOfWork (Mongo-txn) + transactional Outbox (атомарно), gRPC-адаптеры + identity-interceptor, query-сервисы (CQRS-lite)
- [ ] **Phase 8: Event-backbone — схемы + relay → Kafka** — protobuf events/state, relay (idempotent producer, key=ID, ORDER BY sequence), dual-topic (`*.events` delete + `*.state` compact), decommission≠delete на потоке
- [ ] **Phase 9: Топология connections + read-model** — самостоятельные HW-модули, типизированные M:N connections, двунаправленный read-model зависимостей
- [ ] **Phase 10: Верификация backbone** — test-consumer: replay/backfill из `*.state`, чтение истории из `*.events`, порядок per-entity, tombstone-семантика

## Phase Details

### Phase 5: Dev-инфра и стек

**Goal**: Готовое окружение и обновлённый стек, на котором можно писать репозитории и гонять интеграционные тесты — до первого доменного кода.
**Depends on**: Nothing (нулевой шаг v3.0; продолжает v1.0-фундамент)
**Requirements**: SVC-05, SVC-06, SVC-07, DOC-02
**Success Criteria** (what must be TRUE):

  1. `services/inventory` собирается с mongo-driver/**v2** (v1 удалён из `go.mod`), `go build ./...` / `go vet ./...` зелёные из корня workspace (inventory — полноправный член `go.work`, D-01)
  2. `docker compose up` поднимает локальный стенд: Kafka (KRaft, без ZooKeeper) + MongoDB как single-node replica set (транзакции доступны)
  3. Bootstrap-скрипт провижнит топики `inventory.*.events` (cleanup=delete) и `inventory.*.state` (cleanup=compact) с заданной cleanup-policy
  4. Интеграционный тест на testcontainers (Kafka KRaft + Mongo RS) стартует и подключается; Ginkgo v2 + Gomega + mockery подключены и проходят smoke-прогон
  5. `build.md` audit-рецепт (DOC-02) исправлен: документированная команда сборки реально проходит (exit 0)

**Plans**: 5 plansPlans:
**Wave 1**

- [x] 05-01-PLAN.md — go.mod→mongo-driver/v2 swap + Mongo connection-helper + topology-пакет (Bootstrap на kadm) + unit-тест констант
- [x] 05-02-PLAN.md — docker-compose (confluent-local + mongo:7 RS) + Makefile dev-таргеты (mockery pin, dev-up/topics/test-integration/generate-mocks) + ручной SC2-smoke

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 05-03-PLAN.md — .mockery.yaml (v3) + throwaway example-интерфейс + сгенерированный мок + unit-spec (mockery smoke)
- [x] 05-04-PLAN.md — bootstrap-CLI в cmd/ + integration-тест (testcontainers, build-tag integration) — оба зовут общую topology.Bootstrap
- [ ] 05-05-PLAN.md — lefthook de-exclusion (inventory unit в pre-push) + каноны build/structure/boundaries + ROADMAP SC1 + DOC-02 audit-рецепт (go vet)

### Phase 6: Доменная модель Inventory

**Goal**: Спроектированный домен Inventory как агрегаты с инвариантами идентичности/ЖЦ и семантическими доменными событиями — фундамент всего backbone (без событий нечего класть в outbox).
**Depends on**: Phase 5
**Requirements**: INV-01, INV-02, INV-03, INV-04, INV-05, INV-06, INV-07, INV-08, INV-09, INV-10, HW-01, HW-02, HW-03, HW-04, HW-05, HW-06, LOC-01, LOC-02, LOC-03, LOC-04, EVT-01, EVT-02, DOC-07, SVC-01
**Success Criteria** (what must be TRUE):

  1. Оператор может (через usecase) завести Project (`ID`/`Name`/`Description`/`Owner` как непрозрачный внешний string-ID) и зарегистрировать Host с обязательной привязкой к Project; система присваивает внутренний постоянный, непереиспользуемый `ID`
  2. Host несёт `HostHardware` VO (Name/Platform/Motherboard/IPMIMac) со структурированными компонентами RAM/CPU/Drives/NIC/PSU/storage-controller/внутренний GPU/chassis; все внешние идентификаторы — `string`
  3. Host проходит ЖЦ `shadow → registered → decommissioned` + `deleted`; `decommissioned` ≠ `deleted`; повторное добавление = новый `ID` без авто-мерджа (матч — только советочный); FQDN-конфликт среди `active` возвращает доменный конфликт, не сырой DB-error
  4. Локации DC→Module→Rack заводятся как первоклассные сущности с иерархией; Host ссылается на Rack + позицию (юнит); Rack несёт топологические атрибуты (напр. источник питания)
  5. Каждое изменение (идентичность/ЖЦ/железо/локация) рождает **семантическое** доменное событие (`HostRegistered`/`HostHardwareChanged`/`HostReassigned`/`HostDecommissioned`/`HostDeleted`/…), не дамп `HostUpdated`; событие несёт envelope `eventId`/`version`/`actor`/`occurredAt` с первого дня
  6. `knowledge/glossary.md` (DOC-07) фиксирует ubiquitous language: границу «факт существования vs динамическое состояние», термины Project/Host/Owner/Module/Connection, идентичность и `decommission ≠ delete`; код раскладывается по канон-слоям `domain/usecases/query/repositories/api/cron` + composition root `app` (SVC-01)

**Plans**: TBD
**UI hint**: no

### Phase 7: Эталон записи и чтения (UnitOfWork + Outbox + gRPC)

**Goal**: Первая реальная реализация канона записи v1.0 — атомарная запись агрегата и событий в одной Mongo-транзакции, доступная через gRPC; reference-implementation для всех будущих сервисов.
**Depends on**: Phase 6
**Requirements**: SVC-02, SVC-03, SVC-04, SVC-08, EVT-03
**Success Criteria** (what must be TRUE):

  1. Запись агрегата идёт через порт `UnitOfWork` (Mongo-транзакция, writeconcern majority); репозитории берут транзакцию из `ctx`
  2. Доменные события пишутся в outbox-коллекцию **в той же** UoW-транзакции, что и агрегат (transactional outbox, at-least-once, нет dual-write); интеграционный тест atomicity (инъекция паники между Save и Append) проходит
  3. Use cases вызываются через gRPC-адаптеры напрямую (хендлеры зовут use case без диспетчера); read-side — query-сервисы читают Mongo напрямую в DTO (CQRS-lite)
  4. gRPC-слой извлекает identity вызывающего через единый interceptor и пробрасывает её до use case, питая `actor/initiator` события (проверки прав — stub под будущий Access, SEED-003)
  5. Partial unique index FQDN среди `lifecycleState:active` создан и реально освобождает FQDN при decommission/delete

**Plans**: TBD
**UI hint**: no

### Phase 8: Event-backbone — protobuf-схемы + relay → Kafka

**Goal**: Центральный компонент v3.0 — события доходят до Kafka через relay в правильном порядке, по двум топикам с разным retention, по стабильному ключу; forward-compat контракт зафиксирован в схеме до публикации.
**Depends on**: Phase 7
**Requirements**: EVT-06, EVT-04, EVT-05
**Success Criteria** (what must be TRUE):

  1. Схемы событий — protobuf через buf codegen: `events.proto` (HostEvent envelope `eventId`/`entityId`/`version`/`occurredAt`/`actor{id,source}` + payload-oneof) и `state.proto` (HostState) per aggregate; schema registry не вводится
  2. Relay (отдельный loop) читает outbox строго `ORDER BY sequence`, публикует через franz-go idempotent producer (`acks=all`, без лимита ретраев), помечает published; интеграционный тест порядка `create→update→decommission→delete` на одном `entityID` зелёный
  3. Каждое изменение эмитится в dual-topic: `inventory.*.events` (append-only, immutable-история фактов) + `inventory.*.state` (compacted by `entityID`, снапшот); Kafka message key = внутренний `ID`, partition by `entityID`
  4. `HostDecommissioned` = событие + смена `lifecycleState` (хост остаётся в `*.state`-снапшоте); `HostDeleted` = событие в `*.events` + tombstone в `*.state`; `delete.retention.ms` ≥ 24ч

**Plans**: TBD
**UI hint**: no

### Phase 9: Топология connections + read-model зависимостей

**Goal**: Inventory владеет физической топологией парка — типизированные связи хост↔модуль и знание «что зависит от X» — без операций (каскадные действия живут в других доменах).
**Depends on**: Phase 6 (Host identity), Phase 8 (события топологии идут через работающий relay)
**Requirements**: MOD-01, MOD-02, MOD-03
**Success Criteria** (what must be TRUE):

  1. Внешний HW-модуль (дисковая полка / внешний GPU) — самостоятельный агрегат (`type`, внешний `ID` string, **без owner**)
  2. Connections хост↔модуль — типизированные (`power`/`storage`/`data`/`pcie`/`parent-child`), отношение M:N, кросс-ссылки в Mongo по внутренним ID (для хостов) и string-ID (для модулей)
  3. Read-model отвечает двунаправленно: «что зависит от модуля/стойки/генератора» и «от чего зависит хост»; различает `impacted` vs `failed`; только знание, без действий
  4. Decommission/delete хоста или модуля эмитит `ConnectionRemoved` и снимает связи — нет висячих ссылок; каскадных **действий** нет (вне scope)

**Plans**: TBD
**UI hint**: no

### Phase 10: Верификация backbone (test-consumer)

**Goal**: Приёмочный quality-gate продюсера — доказать, что backfill, история, порядок и tombstone-семантика работают end-to-end, и история восстановима после прогона компакции.
**Depends on**: Phase 8 (relay), Phase 9 (события топологии)
**Requirements**: EVT-07
**Success Criteria** (what must be TRUE):

  1. Test-consumer (Ginkgo-suite, testcontainers KRaft) читает `inventory.*.state` с earliest и материализует карту живых сущностей (last-writer-by-version, tombstone = удалить из проекции)
  2. Сценарий онбординга нового домена через backfill из `*.state` проходит; `deleted`-сущность отсутствует в проекции, `decommissioned` — присутствует со статусом
  3. История изменений хоста (`create→hardware→reassign→decommission→delete`) восстановима из `inventory.*.events` **после** прогона компакции `*.state`-топика
  4. Чеклист «Looks Done But Isn't» из ресёрча (append-only история, decommission≠tombstone, atomicity outbox, ORDER BY sequence, actor в envelope, key=ID, partial FQDN-index, re-add-конфликт, очистка connections) пройден

**Plans**: TBD
**UI hint**: no

## Progress

| Phase | Milestone | Plans Complete | Status   | Completed  |
| ----- | --------- | -------------- | -------- | ---------- |
| 1. Раскладка базы знаний и точки входа | v1.0 | 2/2 | Complete | 2026-06-17 |
| 2. Стабильные доки-основы | v1.0 | 3/3 | Complete | 2026-06-17 |
| 3. Доки конвенций и архитектуры | v1.0 | 5/5 | Complete | 2026-06-17 |
| 4. Enforcement-слой (тулинг) | v1.0 | 4/4 | Complete | 2026-06-17 |
| 5. Dev-инфра и стек | v3.0 | 3/5 | In Progress|  |
| 6. Доменная модель Inventory | v3.0 | 0/? | Not started | - |
| 7. Эталон записи и чтения | v3.0 | 0/? | Not started | - |
| 8. Event-backbone — схемы + relay → Kafka | v3.0 | 0/? | Not started | - |
| 9. Топология connections + read-model | v3.0 | 0/? | Not started | - |
| 10. Верификация backbone | v3.0 | 0/? | Not started | - |
