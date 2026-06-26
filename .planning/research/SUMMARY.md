# Project Research Summary

**Project:** gwall-e — Hardware-as-a-Service платформа (домен Inventory + Event-backbone)
**Domain:** DC hardware inventory (identity/lifecycle/topology) + Kafka event-backbone via transactional outbox/relay
**Researched:** 2026-06-26
**Confidence:** HIGH

## Executive Summary

gwall-e v3.0 строит **источник истины об идентичности серверного парка** как Go-микросервис Inventory (DDD + гексагональная архитектура, MongoDB, UnitOfWork/outbox-канон) с **Kafka-продюсером событий** поверх уже зафиксированного relay-канона. Ключевое архитектурное решение этого milestone — **dual-topic per aggregate**: на каждый тип агрегата два Kafka-топика с разными политиками retention: `inventory.host.events` с `cleanup.policy=delete` для immutable-лога фактов/аудита и `inventory.host.state` с `cleanup.policy=compact` для compacted-снапшота идентичности по entityID. Это не косметическое решение: «история на событиях» и «снапшот по entityID» физически несовместимы в одном compacted-топике — compaction стирает промежуточные факты и аудит-след. Разделение закрывает core value «безопасное и согласованное управление» и даёт будущим доменам (Audit, Analytics, Search) бесплатный backfill через replay `*.state`-топика.

Доменная модель Inventory имеет конкретные пробелы относительно зрелых DCIM (Redfish/NetBox): отсутствуют NIC как компонент (есть только плоский `MACs[]`), PSU, BMC, firmware-версии компонентов, StorageController и внутренние GPU. Lifecycle не покрывает промежуточные состояния `failed`/`maintenance` (LC1) и `decommissioning` как in-progress (LC2), а граница «факт-статус ЖЦ vs process-state» не зафиксирована явно. Топология `connections` требует типизации (power/storage/data/parent-child) и двунаправленного read-model. Все эти пробелы включены в v3.0 Launch With — не P2/P3.

Главные риски milestone: (1) dual-write — событие записывается в outbox вне UoW-транзакции; (2) ключ compacted-топика = FQDN/INV вместо внутреннего ID; (3) отсутствие `actor/initiator` в envelope с дня 1 — переэмит immutable-лога задним числом невозможен. Все три требуют превентивного закрытия до первого relay-кода. Стек: franz-go v1.21.4 (pure-Go, idempotent by default), mongo-driver/v2 v2.7.0 (v1 deprecated 2026), testcontainers-go v0.43.0 с KRaft Kafka-модулем, без schema registry (продюсер-only, консьюмеров нет).

---

## Key Findings

### Recommended Stack

Архитектура (Go 1.24.6, DDD+гексагон, MongoDB, UnitOfWork, outbox/relay, Ginkgo/Gomega/mockery) **зафиксирована в v1.0 и не пересматривается**. v3.0 добавляет:

**Core technologies:**
- **franz-go v1.21.4** (`pkg/kgo` + `pkg/kadm`): Kafka-продюсер в relay — pure-Go (нет CGO, статический бинарь), идемпотентный продюсер включён по умолчанию (`acks=all`, sticky-key partitioner). Альтернативы отвергнуты: confluent-kafka-go (CGO), segmentio/kafka-go (нет idempotent-producer), IBM/sarama (legacy).
- **mongo-driver/v2 v2.7.0**: текущая версия services/inventory — v1.17.9, которая **deprecated в 2026**. Обновление обязательно до написания репозиториев/UoW. API транзакций изменился: callback получает `context.Context` вместо `mongo.SessionContext`.
- **testcontainers-go v0.43.0** + `modules/kafka` (KRaft, без ZooKeeper) + `modules/mongodb` (single-node RS — транзакции Mongo требуют replica set). Для интеграционных тестов UoW, outbox-записи, relay→Kafka end-to-end.
- **protobuf / buf** (уже в каноне): один IDL для gRPC и событий, без Schema Registry (продюсер-only; SR добавить при первом консьюмере в режиме BACKWARD_TRANSITIVE).

**Критическое ограничение relay:** не ограничивать число ретраев franz-go (ретраи безопасны при idempotent-продюсере; лимит + идемпотентность = ложное «data loss»).

### Expected Features

**Must have (table stakes — v3.0 Launch With):**
- Host identity + lifecycle: постоянный внутренний ID, статусы shadow/active/decommissioning/decommissioned/deleted + **failed/maintenance** (LC1); FQDN uniqueness среди active (partial index)
- HostHardware VO (расширенный): базовый паттерн `{slot,model,vendor,serial,Inv}` + **NIC** (HW1), **PSU** (HW2), **BMC** (HW4), **firmwareVersion** на компонентах (HW3), **StorageController** (HW5), **внутренний GPU** (HW7), **chassis-паспорт** (HW8)
- Project + Owner (непрозрачный внешний ID) — контейнер владения; `ProjectCreated`
- Самостоятельные HW-модули без owner (дисковые полки, внешние GPU)
- Локация DC→Module→Rack + `position{startUnit,heightU,face,depth}` (L2) + паспортные capacity (L3) + `facilityId` (L5)
- Топология `connections` с типизацией enum (power/storage/data/parent-child/pcie) (T1) + M:N модуль↔хосты (T5)
- Read-model зависимостей **двунаправленный**: «что зависит от X» и «от чего зависит X» (T4)
- Event-backbone: outbox→relay→Kafka, dual-topic (`*.events` append-only + `*.state` compacted), envelope с `eventId`/`version`/`actor` с дня 1
- DOC-07 glossary: ubiquitous language, факт-статусы vs process-state (LC4)

**Should have (v3.x после валидации):**
- Power-топология PSU→feed→PDU→UPS→generator (T2, L1, L6) — триггер: реальный blast-radius
- `impacted` vs `failed` в read-model (T3)
- `decommissioning` как явное переходное состояние saga (LC2) — зависит от Orchestrator
- Fault-domain / A-B feed redundancy (L4)

**Defer (future / отдельные эпики):**
- Sync + советочный reconciliation (SEED-001) — отдельный интеграционный сервис
- VM / VMGroup — модель работы не определена
- Procurement/warranty/EOX — вне Core Value
- Schema Registry — добавить при первом консьюмере (Search/Analytics/Audit)
- Consumer groups, inbox-dedup — вне v3.0
- Power-телеметрия реального времени, cooling/thermal мониторинг — домен Health
- Cable management/patch-панели — домен Network

**Осознанные anti-features (не реализуем намеренно):**
- Авто-мердж / restore-with-merge идентичности (запрещён by design, SEED-001)
- Каскадные действия по топологии (домен Actions/Orchestrator)
- Provisioning-state внутри Inventory (только факт-статусы, не process-state)

### Architecture Approach

Relay-поток: `domain(PullEvents) → UoW(Mongo-txn) → outbox-коллекция` (атомарно) → `relay(poll pending, ORDER BY sequence) → franz-go produce(key=entityID)` → два топика per aggregate → `mark published`. Строгое отделение relay от UoW-пути предотвращает dual-write: relay читает уже закоммиченные строки из БД и публикует асинхронно. Версия (монотонный инкремент) живёт в агрегате, relay только переносит её в metadata.

**Сквозное решение — dual-topic per aggregate (ARCHITECTURE + PITFALLS сошлись на этом):**
- `inventory.host.events` (`cleanup.policy=delete`, длинный retention): append-only лог семантических фактов (`HostRegistered`, `HostHardwareChanged`, `HostReassigned`, `HostDecommissioned`, `HostDeleted`). Источник для Audit/Analytics. История живёт здесь — не в compacted-топике.
- `inventory.host.state` (`cleanup.policy=compact`, `key=hostID`): compacted-снапшот текущей идентичности; tombstone при `HostDeleted`. Источник для backfill/онбординга нового домена (replay с earliest). Tombstone только на `deleted`, не на `decommissioned`.
- Топик аналогично для каждого агрегата (project, hw-module, location).

**Kafka key = внутренний постоянный ID** (не FQDN, не INV, не MAC). Hard rule: FQDN рециклируется, INV/MAC меняется при замене железа — нестабильный ключ compacted-топика ломает идентификацию.

**Major components:**
1. `domain/host`, `domain/project` — агрегаты с `PullEvents()`, доменные события, `version` как поле агрегата
2. `repositories/mongo_uow.go`, `repositories/mongo_outbox.go` — UnitOfWork (Mongo-txn, session в ctx) + Outbox (append в той же txn)
3. `relay/` (новый слой) — poll outbox → mapper(row→protobuf event + state/tombstone) → franz-go idempotent producer → mark published
4. `proto/inventory/host/v1/events.proto` + `state.proto` — envelope (`eventId`, `entityId`, `version`, `occurredAt`, `actor{id,source}`) + payload oneof + HostState
5. Kafka топики (provisioned через `kadm`): `*.events` (delete) + `*.state` (compact, `delete.retention.ms` >= worst-case backfill lag)
6. `test-consumer` — quality-gate: читает `*.state` с earliest, материализует map (last-writer-by-version, tombstone=delete)

**Двунаправленный read-model топологии:** отдельный query-сервис (CQRS-lite канон), читает Mongo в DTO; при росте — материализовать по событиям топологии идемпотентно.

### Critical Pitfalls

1. **История и compacted-снапшот в одном топике** — compaction стирает промежуточные события; история невосстановима после первого прохода клинера. Решение: строго два топика per aggregate: `*.events` (delete) для истории и `*.state` (compact) для снапшота. Никогда — `Never`.
2. **Tombstone как реализация `decommissioned`** — tombstone в Kafka = «забыть ключ»; после `delete.retention.ms` аудит удаления исчезает. Решение: tombstone ТОЛЬКО при терминальном `HostDeleted` и ТОЛЬКО в `*.state`; `HostDecommissioned` = смена `lifecycleState` (хост остаётся в снапшоте). `delete.retention.ms` >= 24h.
3. **Envelope без `actor/initiator` до Audit-эпика** — immutable-лог нельзя обогатить задним числом. Решение: поля `eventId`/`version`/`actor{id,source}` в схеме с дня 1, даже без консьюмеров (SEED-002 forward-compat). Никогда — `Never`.
4. **Kafka key = FQDN / INV / MAC** — нестабильный ключ compacted-топика ломает идентификацию при рециклинге FQDN или замене материнки. Решение: key = внутренний постоянный ID во всех продюсерах.
5. **Dual-write: outbox-вставка вне UoW-транзакции** — событие теряется при краше между commit домена и записью outbox. Решение: `Outbox.Append` берёт txn-`ctx` от UoW; интеграционный тест atomicity (инъекция паники между Save и Append). Никогда — `Never`.
6. **Relay `ORDER BY created_at` вместо `ORDER BY sequence`** — clock-skew/параллельные txn переставляют события одной сущности. Решение: монотонный BIGSERIAL `sequence` в outbox, relay читает строго по нему.
7. **`decommissioned` и `deleted` слиты в один механизм** — потеря семантики. Решение: `HostDecommissioned` != `HostDeleted`; tombstone только на `deleted`.
8. **Re-add выбрасывает сырой DB-error** вместо доменного конфликта с advisory-кандидатами. Решение: проверка FQDN-уникальности в usecase, partial unique index `{fqdn:1, lifecycleState:active}`.
9. **Connections: висячие ссылки + каскадные действия** — decommission не убирает connections; соблазн реализовать каскадные операции (вне scope). Решение: событие `ConnectionRemoved`; read-model — query-сервис, не оркестратор.

---

## Implications for Roadmap

Build order жёстко диктуется зависимостями: домен → запись → события → транспорт → верификация. Каждый шаг — фундамент следующего.

### Phase 1: Инфраструктура и обновление стека
**Rationale:** Нулевой шаг перед любым кодом. mongo-driver v1 deprecated — обновление дешевле до написания репозиториев. Локальный dev-стенд (Kafka KRaft, Mongo RS) блокирует интеграционные тесты.
**Delivers:** Обновлённый `go.mod` (mongo-driver/v2, franz-go, testcontainers), docker-compose с Kafka KRaft + Mongo RS, скрипт provisioning топиков (`kadm`).
**Addresses:** Стек-зависимости (STACK.md).
**Avoids:** Стоимость миграции mongo-driver/v2 после написания кода; null-ключ на compacted-топике.
**Research flag:** Стандартные паттерны, глубокий ресёрч не нужен.

### Phase 2: Доменная модель — агрегаты, события, lifecycle, glossary
**Rationale:** Domain-first. Без доменных событий нечего класть в outbox; без версии в агрегате relay не может гарантировать порядок. Это фундамент всего backbone.
**Delivers:** Агрегаты `Host` (+ расширенный HW VO: NIC/PSU/BMC/firmware/StorageController/GPU/chassis) и `Project` с доменными событиями, `version` как поле агрегата; `PullEvents()`; lifecycle статусы (shadow/active/failed/decommissioning/decommissioned/deleted); glossary DOC-07.
**Addresses:** HW1-HW8, LC1-LC4 (FEATURES.md).
**Avoids:** Жирные `HostUpdated`-дампы; тощие notification; смешение факт-статусов и process-state (Pitfall 5, LC4).
**Research flag:** Паттерны зафиксированы; требует итерации по DOC-07 glossary (LC1: failed vs health-flag граница с доменом Health).

### Phase 3: Эталон UnitOfWork + Outbox (первая реализация канона)
**Rationale:** UoW — первая реальная реализация канона v1.0, reference-implementation для всех будущих сервисов. Outbox должен быть в той же Mongo-txn — до relay, иначе dual-write.
**Delivers:** `mongo_uow.go` (WithTransaction, writeconcern.Majority), `mongo_outbox.go` (Append в той же txn), интеграционный тест atomicity (инъекция паники), partial unique index FQDN + `lifecycleState:active`.
**Addresses:** Фундамент согласованности.
**Avoids:** Dual-write (Pitfall 8).
**Research flag:** Канон зафиксирован. Критично сделать образцово.

### Phase 4: Protobuf-схемы событий (envelope + state)
**Rationale:** Схема фиксирует forward-compat контракт до relay-кода. `actor/initiator` в envelope нельзя добавить задним числом в immutable-лог.
**Delivers:** `proto/inventory/host/v1/events.proto` (HostEvent envelope: `eventId`/`entityId`/`version`/`occurredAt`/`actor{id,source}` + payload oneof: 5 событий) + `state.proto` (HostState); аналогично для Project. Активирует buf codegen.
**Addresses:** Forward-compat для Audit/Analytics (SEED-002).
**Avoids:** Переэмит immutable-лога (Pitfall 6); нет eventId для будущих консьюмеров (Pitfall 2/anti-pattern 2).
**Research flag:** Стандартный protobuf + buf; buf breaking-check автоматически гейтит схему.

### Phase 5: Relay → Kafka (ключевой новый компонент)
**Rationale:** Центральный компонент v3.0. Зависит от UoW+outbox (Phase 3) и схем (Phase 4). Реализует dual-topic emit, idempotent producer, `ORDER BY sequence`.
**Delivers:** `relay/publisher.go` (poll ORDER BY sequence → produce → mark published), `relay/mapper.go` (outbox-row → HostEvent + HostState/tombstone), `relay/kafka_producer.go` (franz-go, key=entityID, без лимита ретраев). Топики provisioned: `inventory.host.events` (delete) + `inventory.host.state` (compact, `delete.retention.ms` >= 24h).
**Addresses:** Event-backbone (FEATURES.md P1); dual-topic per aggregate.
**Avoids:** Dual-write; ORDER BY created_at (Pitfall 4); ключ=FQDN (Pitfall 3); tombstone вместо decommission-события (Pitfall 2).
**Research flag:** Паттерны зафиксированы. Критично: интеграционный тест порядка per-entityID.

### Phase 6: Топология connections + read-model зависимостей
**Rationale:** Зависит от Host identity (Phase 2). После backbone, чтобы ConnectionRemoved-события проходили через уже работающий relay.
**Delivers:** Типизированные M:N connections (enum: power/storage/data/parent-child/pcie); двунаправленный read-model («что зависит от X» + «от чего зависит X»); событие `ConnectionRemoved`; очистка connections при decommission/delete. Самостоятельные HW-модули без owner (строковые внешние ID).
**Addresses:** T1, T4, T5 (FEATURES.md); граница «Inventory даёт read-model, не каскадные действия».
**Avoids:** Висячие ссылки (Pitfall 9); каскадные операции (out of scope); live N+1 join.
**Research flag:** Двунаправленный read-model при масштабе — материализация vs live join; решать при приближении к prod-нагрузке.

### Phase 7: Верификация backbone — test-consumer + интеграционные тесты
**Rationale:** Quality-gate продюсера. Проверяет backfill из `*.state`, порядок per-entityID, tombstone-семантику, восстановимость истории из `*.events` после compaction.
**Delivers:** test-consumer (Ginkgo-suite, testcontainers KRaft): читает `*.state` earliest → материализует map; тест-кейсы: create→update→decommission→delete per host, replay after compaction, backfill onboarding. Проверка «Looks Done But Isn't» чеклиста.
**Addresses:** Верификация всех 9 pitfalls; приёмочный критерий event-backbone.
**Avoids:** «Кажется работает, но история теряется после компакции» (Pitfall 1 warning sign).
**Research flag:** Стандартные тест-паттерны; testcontainers KRaft документирован.

### Phase Ordering Rationale

- **Инфраструктура первой:** mongo-driver/v2 migration дешевле до кода; docker-compose нужен с первых интеграционных тестов.
- **Domain перед persistence:** DDD-канон — сначала доменная модель, потом реализация портов.
- **Схемы перед relay:** forward-compat контракт (`actor`/`eventId`) должен быть зафиксирован до первой публикации в immutable-лог.
- **Топология после backbone:** ConnectionRemoved-события проходят через уже работающий relay.
- **Верификация последней:** полный end-to-end только после всех компонентов.
- **Группировка по инвариантам:** Pitfalls 1+2+4 (backbone) → Phase 5; Pitfalls 3+6 (ключ+actor) → Phases 4+5; Pitfall 8 (dual-write) → Phase 3; Pitfall 9 (connections) → Phase 6.

### Research Flags

**Phases likely needing deeper research during planning:**
- **Phase 6 (Topology read-model):** двунаправленный read-model при масштабе — паттерны материализации по событиям vs live join не специфицированы для нашего размера парка; ресёрч при подходе к prod-нагрузке.
- **Phase 2 (DOC-07 / LC1):** конкретная трактовка `failed` vs `maintenance` как факт-статус или health-flag требует итерации по граничному случаю с доменом Health — зафиксировать в glossary до кода.

**Phases with standard patterns (skip research-phase):**
- **Phase 1:** стек-обновление — официальные инструкции миграции mongo-driver/v2.
- **Phase 3:** UoW + outbox — канон v1.0, задокументирован в knowledge/architecture.md.
- **Phase 4:** protobuf envelope — buf-скелет уже стоит, паттерн oneof envelope стандартный.
- **Phase 5:** relay → franz-go — идемпотентный продюсер по умолчанию, задокументирован в ARCHITECTURE.md.
- **Phase 7:** testcontainers KRaft — официальные docs modules/kafka.

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Версии верифицированы по pkg.go.dev (franz-go v1.21.4 Jun 2026, mongo-driver/v2 v2.7.0 Jun 2026, testcontainers v0.43.0 Jun 2026); deprecation mongo-driver v1 — официальный нотис |
| Features | HIGH / MEDIUM | Hardware-модель верифицирована по Redfish/NetBox официальной документации; приоритеты HW-gaps — контекст-специфичны для solo-режима (MEDIUM по приоритизации) |
| Architecture | HIGH | dual-topic паттерн верифицирован по Confluent docs + независимым практикам; outbox/relay/franz-go cross-checked WarpStream + maintainer docs |
| Pitfalls | HIGH | Каждый pitfall верифицирован по >= 2 независимым источникам (Confluent docs, Ted Naleid, Conduktor, Decodable, Microsoft Learn, SEED-001/002 внутренние решения) |

**Overall confidence:** HIGH

### Gaps to Address

- **LC1/LC2 граница с Health:** `failed`/`maintenance` как факт-статус vs health-flag — пограничный случай между доменами Inventory и Health. Решить в DOC-07 glossary (Phase 2) до кода статусной машины.
- **LC3 planned/staged:** достаточно ли одного `shadow` или нужны два статуса — решить при написании glossary DOC-07 после наблюдения реального solo-наполнения.
- **T2/L1/L6 power-топология:** отложена в P2; архитектура connections (Phase 6) должна допускать добавление power-цепочки без переделки схемы.
- **T3 impacted vs failed:** потребуется при появлении первого консьюмера read-model (Health/Scenarios); заложить extensibility в Phase 6.
- **delete.retention.ms конкретное значение:** зависит от worst-case lag будущих консьюмеров. Правило: >= 24h на старте; пересмотреть при появлении первого консьюмера с его backfill-окном.
- **Число партиций Kafka:** задать с запасом сразу (увеличение позже ломает key→partition маппинг); конкретное значение — при Phase 5, исходя из целевого 150k парка.

---

## Sources

### Primary (HIGH confidence)
- pkg.go.dev `github.com/twmb/franz-go` v1.21.4 (Jun 2026) — Kafka-продюсер, idempotent по умолчанию, ретраи
- pkg.go.dev `go.mongodb.org/mongo-driver/v2` v2.7.0 (Jun 2026) — UoW/транзакции, deprecation v1
- pkg.go.dev `github.com/testcontainers/testcontainers-go` v0.43.0 (Jun 2026) — KRaft Kafka + Mongo RS
- Kafka Log Compaction — Confluent docs — tombstone, delete.retention.ms, compaction семантика
- Key-Based Retention — Confluent Developer — dual-topic паттерн
- franz-go producing-and-consuming docs (maintainer) — idempotent by default, record-retry семантика
- mongodb.com/docs/drivers/go (v2 transactions) — WithSession/WithTransaction, replica-set requirement
- NetBox DCIM Devices/Inventory — hardware модель, lifecycle статусы
- Redfish (DMTF) / Supermicro Redfish — NIC/PSU/BMC/firmware компоненты
- Event Sourcing Pattern — Microsoft Learn — append-only лог vs снапшот
- Внутренние: SEED-001, SEED-002, L2-ARCHITECTURE, knowledge/architecture.md

### Secondary (MEDIUM confidence)
- WarpStream Kafka client tuning — franz-go рекомендация для idempotency/ordering
- Conduktor: Transactional Outbox — aggregate_id как ключ, порядок per-partition
- Decodable: Revisiting Outbox — polling ORDER BY sequence vs timestamp
- Ted Naleid: Understanding Kafka Compaction — tombstone требует двух проходов, гонка delete.retention
- USPTO 7711980 — impacted vs failed, dependency graph
- Nautobot Tenancy — tenant = single-owner, uniqueness
- microservices.io transactional outbox — at-least-once, дедуп на консьюмере

---
*Research completed: 2026-06-26*
*Ready for roadmap: yes*
