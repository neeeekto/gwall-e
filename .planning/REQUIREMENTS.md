# Requirements: gwall-e — Milestone v3.0 (Inventory + Event-backbone)

**Defined:** 2026-06-19
**Core Value:** Безопасное и согласованное управление парком серверов как услугой — единый источник правды о хостах.

> Скоуп v3.0 — первый реальный сервис **Inventory** (solo-режим) + продюсер-бэкбон событий на Kafka, как эталонная реализация архитектуры v1.0. Контекст и решения — в [PROJECT.md](PROJECT.md); доменный ресёрч — в [research/SUMMARY.md](research/SUMMARY.md).
>
> **Граница домена:** Inventory хранит **факт существования** железа (идентичность, ЖЦ-как-актив, состав железа, локацию, топологию), а **НЕ динамические эксплуатационные состояния** хоста (runtime-здоровье, failed/maintenance) — это State/Health-сервис.

## Milestone v3.0 Requirements

### INV — Inventory: идентичность и жизненный цикл

- [ ] **INV-01**: Оператор может завести Project (`ID`, `Name`, `Description`, `Owner`)
- [ ] **INV-02**: Оператор может зарегистрировать Host с обязательной привязкой к Project
- [x] **INV-03**: Система присваивает хосту/проекту внутренний постоянный `ID` (генерится системой, не переиспользуется) — единственный носитель идентичности
- [ ] **INV-04**: Host имеет ЖЦ-статус `shadow → registered → decommissioned` + `deleted` (только факт существования железа, не динамическое состояние)
- [ ] **INV-05**: Оператор может переназначить Host в другой Project (инвентарная коррекция; безопасный перенос-с-затиркой — вне scope)
- [ ] **INV-06**: Оператор может decommission хоста (списание железа; терминально; ≠ `deleted`)
- [ ] **INV-07**: Оператор может удалить (`deleted`) запись хоста/проекта; история сохраняется на событиях; FQDN освобождается
- [x] **INV-08**: Повторное добавление хоста = новый `ID` без авто-мерджа; матч с прошлыми записями по ключу — только советочный (хук под будущую интеграцию)
- [ ] **INV-09**: `Owner` хранится как непрозрачный внешний `string`-ID группы; резолв — наружу (вне scope)
- [x] **INV-10**: `FQDN` уникален только среди `active`-хостов (partial unique index); Project можно удалить только пустым

### HW — Модель железа хоста

- [ ] **HW-01**: Host несёт `HostHardware` как VO внутри агрегата (`Name`, `Platform`, `Motherboard`, `IPMIMac`)
- [ ] **HW-02**: Hardware включает структурированные компоненты RAM / CPU / Drives (`slot`/`model`/`vendor`/`lot`/`serial`/`Inv string` + спеки)
- [ ] **HW-03**: NIC моделируется как структурированный компонент (модель, скорость, MAC'и) вместо плоского `MACs[]`
- [ ] **HW-04**: Hardware включает PSU (блоки питания) — узлы power-зависимости
- [ ] **HW-05**: Hardware покрывает storage-controller/RAID, внутренние GPU и паспорт шасси отдельно от материнки
- [ ] **HW-06**: Все внешние идентификаторы компонентов хранятся как `string` (любой формат — числа…UUID)

### LOC — Локации

- [ ] **LOC-01**: Оператор может завести/изменить DC, Module, Rack как первоклассные сущности (CRUD)
- [ ] **LOC-02**: Локации образуют иерархию `DC → Module → Rack`
- [ ] **LOC-03**: Host ссылается на Rack + позицию (юнит) в стойке
- [ ] **LOC-04**: Rack несёт атрибуты (напр. источник питания / дизель-генератор) как узлы топологии

### MOD — Внешние HW-модули и топология

- [ ] **MOD-01**: Внешний HW-модуль (дисковая полка / внешний GPU) — самостоятельный агрегат (`type`, внешний `ID` string, **без owner**)
- [ ] **MOD-02**: Connections хост↔модуль — типизированные (`power`/`storage`/`data`/`pcie`/`parent-child`), отношение **M:N**, кросс-ссылки в Mongo
- [ ] **MOD-03**: Read-model топологии зависимостей — двунаправленный («что зависит от модуля/стойки/генератора»), различает `impacted` vs `failed`; только знание, без действий

### EVT — Event-backbone (Kafka, продюсер)

- [x] **EVT-01**: Сервис эмитит семантические гранулярные доменные события на все изменения (идентичность/ЖЦ/железо/локация/топология)
- [x] **EVT-02**: Каждое событие несёт envelope с `eventId`, `version` (поле агрегата), `actor/initiator` и `occurredAt` — с первого дня (forward-compat для Audit)
- [ ] **EVT-03**: События публикуются через transactional outbox внутри UoW + отдельный relay → Kafka (idempotent producer, at-least-once)
- [ ] **EVT-04**: Dual-topic на тип агрегата — `*.events` (append-only, immutable-история фактов: Analytics/replay/backfill/debug) + `*.state` (compacted by `entityID`, снапшот/backfill). *Как поверх этого строится Audit (envelope-consume vs выделенный audit-stream) — research при E11, см. [SEED-002](seeds/SEED-002-audit-logging.md); v3.0 от этого выбора не зависит*
- [ ] **EVT-05**: Kafka message key = внутренний `ID`; partition by `entityID` (порядок per-entity); `decommissioned` = событие, `deleted` = tombstone в `*.state`
- [ ] **EVT-06**: Схемы событий — protobuf (buf codegen активируется); schema registry не вводится (продюсер-only)
- [ ] **EVT-07**: Тестовый консьюмер подтверждает replay/backfill из `*.state` и чтение истории из `*.events` (prod-консьюмеров нет)

### SVC — Эталонный сервис и dev-инфра

- [x] **SVC-01**: Сервис реализует канон-слои `domain / usecases / query / repositories / api(gRPC) / cron` + composition root (`app`)
- [ ] **SVC-02**: Запись идёт через порт `UnitOfWork` (Mongo-транзакция; требует replica set)
- [ ] **SVC-03**: Use cases доступны через gRPC-адаптеры (хендлеры зовут use case напрямую, без диспетчера)
- [ ] **SVC-04**: Read-side — query-сервисы читают Mongo напрямую в DTO (CQRS-lite)
- [x] **SVC-05**: Персистентность — MongoDB через mongo-driver **v2** (миграция с v1 до написания репозиториев)
- [x] **SVC-06**: Тесты — Ginkgo v2 + Gomega + mockery; интеграционные через testcontainers (Kafka KRaft + Mongo single-node RS)
- [x] **SVC-07**: Dev-инфра — docker-compose (Kafka KRaft + Mongo RS) + bootstrap провижна топиков (`*.events`/`*.state` с нужной cleanup-policy)
- [ ] **SVC-08**: gRPC-слой извлекает идентичность вызывающего (caller identity) и пробрасывает её до use case через единую точку перехвата (interceptor); проверки прав НЕ реализуем (stub под будущий Access), но identity питает `actor/initiator` (forward-compat, см. [SEED-003](seeds/SEED-003-authorization-on-all-actions.md))

### DOC — Документация и техдолг v1.0

- [ ] **DOC-07**: `knowledge/glossary.md` — ubiquitous language домена Inventory; фиксирует границу «факт существования vs динамическое состояние» и термины (Project/Host/Owner/Module/Connection/идентичность/decommission≠delete)
- [ ] **DOC-02**: Починить `build.md` audit-рецепт (carry-over из v1.0) — по ходу разработки

## Future Requirements

Признаны, но отложены в будущие эпики/milestone'ы (не в текущем роадмапе).

### VM / интеграции / аудит

- **VM-01**: VM и VMGroup как полноценные сущности домена (модель работы уточняется)
- **SYNC-01**: Интеграционный сервис sync из внешней инвентори (матчинг/reconciliation — см. [SEED-001](seeds/SEED-001-inv-matching-instability.md))
- **AUD-01**: Домен Audit как consumer событий (см. [SEED-002](seeds/SEED-002-audit-logging.md))
- **ACCESS-01**: Домен Access — проверка прав на все операции (без этого любой может сделать что угодно; см. [SEED-003](seeds/SEED-003-authorization-on-all-actions.md))

### Расширения домена

- **TOPO-FUT-01**: Полная power-цепочка топологии (PSU → feed → PDU → UPS → generator как узлы)
- **LC-FUT-01**: Дополнительные ЖЦ-состояния (`planned`/`staged`) — зависят от паттерна solo-наполнения
- **EVT-FUT-01**: Schema registry + consumer/inbox-инфра (с первым реальным доменом-потребителем); EOS — по необходимости

## Out of Scope

| Feature | Reason |
|---------|--------|
| Динамические эксплуатационные состояния (`failed`/`maintenance`/runtime-health) | Граница домена: Inventory = факт существования, не динамика → State/Health-сервис |
| Версии прошивок (firmware/BIOS/BMC) | Перепрошиваются динамически — эксплуатационное состояние, не факт существования → State/Health-сервис |
| Авторизация/гранты (owner-роли, SSH, IDM) | Домен Access |
| Действия над хостом (reboot/reimage/profile), затирки | Домен Actions |
| Сеть: VLAN/IPAM/конфиг свитчей, выдача FQDN | Домен Network (Inventory только хранит присвоенный FQDN) |
| Каскадные действия по топологии («погасить пострадавшие») | Read-model только знает зависимости; действия — Actions/Remediation |
| Безопасный перенос хоста (reimage+wipe+approve) | Сага Orchestrator/Coordination/Actions |
| Авто-мердж идентичности / restore-with-merge при повторном добавлении | By design запрещены (ложный матч на рециклинге FQDN/смене материнки) |
| Prod-консьюмеры, inbox, exactly-once (EOS) | Продюсер-only этап; консьюмеры приходят с доменами-потребителями |
| Фронтенд (React/Nx) | После бэкенд-доменов |

## Traceability

Каждое требование → ровно одна фаза (v3.0 = Phases 5-10). Маппинг — из [ROADMAP.md](ROADMAP.md).

| Requirement | Phase | Status |
|-------------|-------|--------|
| INV-01 | Phase 6 | Pending |
| INV-02 | Phase 6 | Pending |
| INV-03 | Phase 6 | Complete |
| INV-04 | Phase 6 | Pending |
| INV-05 | Phase 6 | Pending |
| INV-06 | Phase 6 | Pending |
| INV-07 | Phase 6 | Pending |
| INV-08 | Phase 6 | Complete |
| INV-09 | Phase 6 | Pending |
| INV-10 | Phase 6 | Complete |
| HW-01 | Phase 6 | Pending |
| HW-02 | Phase 6 | Pending |
| HW-03 | Phase 6 | Pending |
| HW-04 | Phase 6 | Pending |
| HW-05 | Phase 6 | Pending |
| HW-06 | Phase 6 | Pending |
| LOC-01 | Phase 6 | Pending |
| LOC-02 | Phase 6 | Pending |
| LOC-03 | Phase 6 | Pending |
| LOC-04 | Phase 6 | Pending |
| MOD-01 | Phase 9 | Pending |
| MOD-02 | Phase 9 | Pending |
| MOD-03 | Phase 9 | Pending |
| EVT-01 | Phase 6 | Complete |
| EVT-02 | Phase 6 | Complete |
| EVT-03 | Phase 7 | Pending |
| EVT-04 | Phase 8 | Pending |
| EVT-05 | Phase 8 | Pending |
| EVT-06 | Phase 8 | Pending |
| EVT-07 | Phase 10 | Pending |
| SVC-01 | Phase 6 | Complete |
| SVC-02 | Phase 7 | Pending |
| SVC-03 | Phase 7 | Pending |
| SVC-04 | Phase 7 | Pending |
| SVC-05 | Phase 5 | Complete |
| SVC-06 | Phase 5 | Complete |
| SVC-07 | Phase 5 | Complete |
| SVC-08 | Phase 7 | Pending |
| DOC-07 | Phase 6 | Pending |
| DOC-02 | Phase 5 | Pending |

**Coverage:**

- v3.0 requirements: **40** total (INV 10 · HW 6 · LOC 4 · MOD 3 · EVT 7 · SVC 8 · DOC 2)
- Mapped to phases: **40 / 40** ✓
- Unmapped: **0**

**По фазам:**

- Phase 5 (Dev-инфра и стек): SVC-05, SVC-06, SVC-07, DOC-02 — 4
- Phase 6 (Доменная модель Inventory): INV-01…10, HW-01…06, LOC-01…04, EVT-01, EVT-02, DOC-07, SVC-01 — 24
- Phase 7 (Эталон записи и чтения): SVC-02, SVC-03, SVC-04, SVC-08, EVT-03 — 5
- Phase 8 (Event-backbone — схемы + relay): EVT-06, EVT-04, EVT-05 — 3
- Phase 9 (Топология connections): MOD-01, MOD-02, MOD-03 — 3
- Phase 10 (Верификация backbone): EVT-07 — 1

---
*Requirements defined: 2026-06-19*
*Last updated: 2026-06-27 — traceability заполнена при создании роадмапа v3.0 (Phases 5-10, 40/40 mapped)*
