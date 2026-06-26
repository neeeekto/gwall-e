# Architecture Research

**Domain:** Event-backbone (Kafka) поверх transactional-outbox/relay-канона — продюсер-этап для домена Inventory
**Researched:** 2026-06-26
**Confidence:** HIGH (Kafka log-compaction/tombstone семантика — официальный Confluent docs; outbox+idempotent-producer и dual-topic — несколько сходящихся источников; Go-клиент franz-go — корроборировано WarpStream + maintainer docs)

> **Скоуп этого ресёрча.** Только **архитектура event-backbone** для milestone v3.0:
> как Kafka ложится на уже зафиксированный канон `UoW → outbox → relay`, какие
> **новые** компоненты появляются, дизайн топиков/ключей/схем, и **build order**.
> Доменная модель Inventory (агрегаты Project/Host, hardware-VO, топология) —
> предмет FEATURES/PROJECT, здесь не проектируется. Консьюмеры (Search/Analytics/Audit) —
> вне v3.0 (продюсер-only), но **forward-compat** для них фиксируется.

---

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                       INVENTORY СЕРВИС (Go-модуль)                    │
│                                                                       │
│  api(gRPC) ─► usecases ─► domain (агрегат + PullEvents)               │
│                  │                                                    │
│                  ▼  одна Mongo-транзакция (порт UnitOfWork)           │
│        ┌──────────────────────────────────────────┐                  │
│        │  repositories: aggregate-collection  +    │  АТОМАРНО        │
│        │                outbox-collection          │                  │
│        └──────────────────────────────────────────┘                  │
│                  │ (txn commit)                                       │
│                  ▼                                                    │
│        ┌──────────────────────────────────────────┐                  │
│        │  cron/relay: poll outbox (status=pending) │  АСИНХРОННО      │
│        │  → map outbox-row → protobuf event(s)     │  (отдельный      │
│        │  → Kafka producer (idempotent, acks=all)  │   процесс/loop)  │
│        │  → mark published                         │                  │
│        └──────────────────┬───────────────────────┘                  │
└─────────────────────────────┼─────────────────────────────────────────┘
                              │ partition by entityID, key = entityID
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         KAFKA (destination)                          │
│                                                                       │
│  inventory.host.events     (cleanup=delete, retention=∞/long)        │
│      └ HostRegistered, HostHardwareChanged, HostReassigned,          │
│        HostDecommissioned, HostDeleted …   (история фактов)          │
│                                                                       │
│  inventory.host.state      (cleanup=compact, key=hostID)            │
│      └ снапшот текущей идентичности Host; tombstone при HostDeleted  │
│                                                                       │
│  inventory.project.events  (delete) / inventory.project.state (compact)
│                                                                       │
│  …топик-на-тип-агрегата (host / project / hw-module / location)      │
└─────────────────────────────────────────────────────────────────────┘
                              │ (v3.0: НЕТ консьюмеров, кроме тест-consumer)
                              ▼
        ┌──────────── будущие домены (Search / Analytics / Audit) ────────────┐
        │  consumer group + inbox (dedup by eventId) + ACL → локальный язык   │
        │  backfill: read compact-топик с earliest → материализовать проекцию │
        └─────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Новое / Модифицируемое |
|-----------|----------------|------------------------|
| `domain` (агрегат + `PullEvents`) | Рождает **семантические** доменные события как часть инвариантов | Новое (домен Inventory с нуля); механика `PullEvents` — **канон, не меняется** |
| `repositories.outbox` | Append событий в outbox-коллекцию **в той же UoW-txn**, что и агрегат | Новое (реализация), контракт «append в txn» — канон |
| `repositories.UnitOfWork` | Транзакционная граница (Mongo-txn), кладёт txn в `ctx` | Канон v1.0 — **не меняется** (реализация пишется впервые) |
| **`cron/relay` (relay→Kafka publisher)** | Poll outbox → маппинг row→protobuf → **idempotent Kafka producer** → mark published; эмитит И event, И state-снапшот | **НОВОЕ** — ключевой компонент v3.0 |
| **Топик-дизайн (`*.events` + `*.state`)** | Per-aggregate: append-лог фактов + compacted-снапшот идентичности | **НОВОЕ** — инфраструктура (topic provisioning) |
| **Схемы событий (protobuf)** | Контракт событий: envelope (metadata) + payload-oneof по типу | **НОВОЕ** — `buf` скелет уже есть, .proto пишутся впервые |
| `tеst-consumer` | Проверка replay/backfill: читает compact-топик с earliest, материализует map | **НОВОЕ** (только для верификации, не прод-домен) |

---

## Recommended Project Structure

```
services/inventory/internal/
├── domain/
│   ├── host/
│   │   ├── host.go              # агрегат Host (фабрика держит инварианты)
│   │   ├── events.go            # доменные события: HostRegistered, HostHardwareChanged,
│   │   │                        #   HostReassigned, HostDecommissioned, HostDeleted
│   │   └── ports.go             # порты: HostRepository, UnitOfWork (если общий — выше)
│   ├── project/                 # ProjectCreated, OwnershipChanged, ProjectDecommissioned…
│   ├── shared/
│   │   ├── outbox.go            # порт Outbox.Append(ctx, []DomainEvent)
│   │   └── event.go             # доменное событие: тип + metadata (eventId, version, occurredAt, actor)
│   └── …
├── usecases/                    # 1 use case = struct + Execute; зовёт uow.Do{ save + outbox.Append }
├── query/                       # read-side (Mongo→DTO), к event-backbone не относится
├── repositories/
│   ├── mongo_uow.go             # реализация UnitOfWork (Mongo-txn)
│   ├── mongo_host.go            # реализация HostRepository
│   └── mongo_outbox.go          # реализация Outbox (та же txn из ctx)
├── relay/                       # ◄── НОВЫЙ слой (inbound от outbox, outbound в Kafka)
│   ├── publisher.go             # poll outbox → producer → mark published
│   ├── mapper.go                # outbox-row → protobuf (event + state-snapshot)
│   └── kafka_producer.go        # обёртка franz-go (idempotent, acks=all)
├── api/                         # gRPC-адаптеры
└── cron/                        # запуск relay-loop как джобы (или отдельный cmd)

services/inventory/proto/ (+ buf gen → gen/go/)
├── inventory/host/v1/
│   ├── events.proto            # HostEvent envelope + payload-oneof
│   └── state.proto             # HostState (compacted snapshot value)
└── inventory/project/v1/ …
```

### Structure Rationale

- **`relay/` как отдельный слой, а не часть `repositories`:** relay — это inbound-адаптер от outbox и outbound-адаптер к Kafka одновременно; он **не** в транзакционном пути записи агрегата (читает уже закоммиченные строки). Держать его рядом с UoW означало бы соблазн опубликовать в брокер внутри txn (= dual-write, **анти-паттерн канона**). Физическое разделение фиксирует «запись агрегата + outbox атомарно; публикация — после commit, асинхронно».
- **Топик-на-тип-агрегата (`proto/inventory/host`, `…/project`):** соответствует кросс-доменному инварианту «у каждой сущности свой агрегат» и даёт независимое версионирование схем и retention per-aggregate.
- **`events.proto` отдельно от `state.proto`:** два разных контракта (поток фактов vs снапшот текущего состояния) → разные топики, разный `cleanup.policy`. См. dual-topic ниже.

---

## Architectural Patterns

### Pattern 1: Transactional Outbox → async relay → idempotent Kafka producer

**What:** Агрегат и его события пишутся в **одну** Mongo-транзакцию (агрегат-коллекция + outbox-коллекция). Отдельный relay-процесс читает `pending` строки, публикует в Kafka с включённой **идемпотентностью продюсера** (`enable.idempotence=true`, `acks=all`), помечает `published`.

**When to use:** Всегда для доменных событий gwall-e — это прямая реализация core value «согласованность» без dual-write. Канон v1.0; Kafka здесь — лишь destination relay.

**Trade-offs:**
- (+) Нет dual-write: запись и публикация не могут разойтись (atomic в БД).
- (+) Идемпотентный продюсер устраняет **дубли от ретраев продюсера** (PID + sequence number на партицию) — но НЕ дубли от перезапуска relay (см. ниже).
- (−) Доставка **at-least-once**, не exactly-once: если relay упал между «publish ok» и «mark published», после рестарта то же событие отправится повторно. Это by design.
- (−) Eventual consistency: между commit и publish есть лаг (poll-интервал relay).

**Двухуровневая идемпотентность (важно для build order):**
1. **Idempotence продюсера** (флаг клиента) — дедуп ретраев внутри одной producer-сессии.
2. **eventId в metadata каждого события** — дедуп на стороне **консьюмера** (inbox), покрывает дубли от рестарта relay. v3.0 консьюмеров не пишет, но `eventId` **обязан** быть в схеме сейчас (иначе будущим доменам нечем дедуплицировать → переэмит событий).

**Example:**
```go
// usecase: запись агрегата + outbox атомарно (граница UoW)
func (uc *RegisterHostUseCase) Execute(ctx context.Context, in RegisterHostInput) (RegisterHostOutput, error) {
    host, err := host.NewHost(/* … инварианты */)
    if err != nil { return RegisterHostOutput{}, err }
    err = uc.uow.Do(ctx, func(ctx context.Context) error {
        if err := uc.hosts.Save(ctx, host); err != nil { return err }
        return uc.outbox.Append(ctx, host.PullEvents()) // та же txn из ctx
    })
    // публикация в Kafka — НЕ здесь; relay сделает это после commit
    return RegisterHostOutput{ID: host.ID()}, err
}

// relay (отдельный loop): poll → publish → mark; producer idempotent by default (franz-go)
for _, row := range outbox.PollPending(ctx, batch) {
    ev, state := mapper.ToKafka(row)        // event + (опц.) snapshot
    _ = producer.Produce(ctx, ev)           // key = entityID → partition by entityID
    if state != nil { _ = producer.Produce(ctx, state) }
    _ = outbox.MarkPublished(ctx, row.ID)   // если упали до сюда — повтор (at-least-once)
}
```

### Pattern 2: Dual-topic per aggregate — append-лог фактов + compacted-снапшот идентичности

**What:** На каждый тип агрегата — **два** топика:
- `inventory.host.events` — `cleanup.policy=delete`, длинный/бесконечный retention. Семантические события (`HostRegistered`, `HostHardwareChanged`, `HostReassigned`, `HostDecommissioned`, `HostDeleted`). Это **история фактов** и аудит-след (источник для Audit/Analytics).
- `inventory.host.state` — `cleanup.policy=compact`, `key = hostID`, `value = HostState`. Снапшот **текущей идентичности**. Tombstone (`value=null`, key=hostID) при `HostDeleted`. Это бесплатный **backfill-механизм**: новый домен читает топик с earliest → реконструирует «карту всех живых хостов».

**When to use:** Стандартный зрелый CQRS/event-sourcing-дизайн на Kafka. Разделение оправдано тем, что у двух потоков **разный жизненный цикл retention**: факты живут вечно (история на событиях — Key Decision PROJECT.md), снапшот — только «последнее на ключ».

**Trade-offs:**
- (+) `*.state` даёт онбординг нового сервиса «бесплатно»: replay earliest → текущее состояние без отдельных hydration-скриптов.
- (+) `*.events` сохраняет полную историю → `decommissioned`-факт и `deleted`-факт остаются в логе даже после того, как снапшот стёрт tombstone'ом. Это прямо реализует решение PROJECT.md «история живёт на потоке событий, не в soft-deleted записи».
- (−) Relay эмитит в два топика → две записи на одно изменение. Не атомарно между топиками, но это терпимо: оба идемпотентны по ключу/eventId, консьюмер last-writer-by-version.
- (−) Compaction **недетерминирована**: на голове лога возможны дубли ключа и ещё-не-стёртые старые значения. Консьюмер ОБЯЗАН применять **last-write-wins по version**, а не «один ключ = одна запись».

**Альтернатива (отвергнута для v3.0):** один топик со встроенными snapshot-событиями. Проще операционно (один топик), но смешивает retention фактов и снапшота — нельзя одновременно «события вечно» и «compaction по ключу». Для gwall-e история-навсегда + backfill-снапшот = разные требования → два топика.

### Pattern 3: Partition-by-entityID + per-entity version (last-writer-by-version)

**What:** `key = entityID` для **обоих** топиков. Гарантия Kafka: все записи с одним ключом → одна партиция → **строгий порядок в рамках сущности**. Каждое событие несёт монотонную `version` (seq на сущность, инкремент в агрегате). Консьюмеры применяют **last-writer-by-version**: запись с `version ≤ уже_применённой` отбрасывается.

**When to use:** Всегда. Это инвариант L2 «терпимость к out-of-order: версия/seq на сущность». Партиционирование даёт порядок *внутри* сущности; `version` страхует от out-of-order *между* партициями/при replay/при compaction-дублях.

**Trade-offs:**
- (+) Порядок per-entity без глобального упорядочивания (которое не масштабируется на 150k хостов).
- (+) `version` делает хендлеры replay-safe и идемпотентными к переупорядочиванию.
- (−) Партиций фиксированное число — горячий ключ теоретически перегружает партицию, но entityID распределён равномерно (UUID-подобный), это не риск.
- (−) `version` должна жить в **агрегате** (источник истины инкремента), а не назначаться relay'ем — иначе при двух эмитах рассинхрон. Версия = поле агрегата, растёт при каждом доменном изменении.

### Pattern 4: protobuf event envelope (metadata) + payload-oneof

**What:** Единый конверт события с metadata + типизированный payload через `oneof`.

```protobuf
message HostEvent {
  // --- metadata (общий конверт, forward-compat для Audit/Search) ---
  string  event_id    = 1;  // UUID; ключ дедупликации на консьюмере (inbox)
  string  entity_id   = 2;  // hostID; == Kafka message key (partition by)
  uint64  version     = 3;  // seq на сущность; last-writer-by-version
  google.protobuf.Timestamp occurred_at = 4;
  Actor   actor       = 5;  // SEED-002: кто инициатор (id + source)

  // --- payload (семантический факт) ---
  oneof payload {
    HostRegistered        registered        = 10;
    HostHardwareChanged   hardware_changed  = 11;
    HostReassigned        reassigned        = 12;
    HostDecommissioned    decommissioned    = 13;
    HostDeleted           deleted           = 14;
  }
}

message Actor {                 // SEED-002 forward-compat
  string id     = 1;
  Source source = 2;            // HUMAN | API | INTEGRATION | SYSTEM
}
```

**When to use:** Всегда. `buf`-скелет (`buf.yaml`/`buf.gen.yaml`, pin protoc-gen-go v1.36.5) уже стоит — активируется добавлением .proto.

**Trade-offs:**
- (+) Protobuf + `buf` — backward/forward-compat при правильной эволюции (только добавлять поля с новыми тегами, не переиспользовать/не удалять теги; `buf breaking` ловит нарушения).
- (+) `oneof` payload = одна схема на тип агрегата, легко расширять новыми событиями.
- (+) Metadata-конверт фиксирует `actor` СЕЙЧАС → Audit (E11) не требует переэмита (SEED-002).
- (−) `state.proto` (compacted value) — отдельная схема (`HostState`), не envelope; consumer её материализует, tombstone = `value=null` (key-only).

---

## Data Flow

### Write Flow (продюсер, v3.0)

```
gRPC RegisterHost
    ↓
api → usecases.RegisterHostUseCase.Execute
    ↓  uow.Do(ctx, fn):
       ├─ hosts.Save(ctx, host)              ─┐
       └─ outbox.Append(ctx, PullEvents())    ├─ ОДНА Mongo-txn (атомарно)
                                              ─┘
    ↓  (txn commit)
[relay loop, асинхронно]
    poll outbox WHERE status=pending
    ↓
    mapper: outbox-row → HostEvent (protobuf) + HostState (если меняется идентичность)
    ↓
    producer.Produce(key=hostID): → inventory.host.events
    producer.Produce(key=hostID): → inventory.host.state   (или tombstone при HostDeleted)
    ↓
    outbox.MarkPublished           (падение до сюда ⇒ повтор = at-least-once)
```

### Backfill / Replay Flow (проверяется test-consumer, прод-консьюмеры — будущее)

```
Новый домен стартует
    ↓
consumer group (уникальная), auto.offset.reset=earliest
    ↓
read inventory.host.state с offset 0 → head
    ├─ value≠null → upsert hostID в локальную проекцию (last-writer-by-version)
    └─ value=null (tombstone) → удалить hostID из проекции
    ↓
переключиться на tail (новые события)
ОГРАНИЧЕНИЕ: backfill ДОЛЖЕН завершить скан быстрее delete.retention.ms,
            иначе tombstone стёрт → «ghost»-хост останется в проекции.
```

### Key Data Flows

1. **Семантический факт → Audit/Analytics:** `*.events`-топик = иммутабельный аудит-след; `actor` в metadata атрибутирует «кто». Audit-домен (E11) — будущий консьюмер.
2. **Идентичность → backfill нового домена:** `*.state`-compacted = «снапшот всех живых сущностей»; replay earliest = онбординг сервиса без отдельной миграции. Прямая реализация L2 «log-compaction → снапшот идентичности бесплатно».
3. **decommission vs delete (разные события, разный эффект на снапшот):**
   - `HostDecommissioned` (списание) → событие в `*.events`; в `*.state` хост **остаётся** (snapshot со статусом `decommissioned`) — он ещё известен системе.
   - `HostDeleted` (убрать запись) → событие в `*.events` (история сохраняется) **+ tombstone** (`value=null`) в `*.state` → хост исчезает из снапшота/backfill, но факт удаления навсегда в логе фактов.
   - FQDN-uniqueness среди `active` — доменный инвариант записи; на backfill consumer строит проекцию `active` из снапшота (decommissioned ≠ active).

---

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| ~Сотни хостов (старт solo) | Один relay-loop, малый poll-интервал; число партиций можно держать скромным (например 6–12) |
| ~Тысячи–десятки тысяч | entityID-партиционирование держит порядок; масштаб — добавлением партиций (но менять число партиций ломает key→partition-маппинг — закладывать запас сразу) |
| ~150k хостов (целевой парк) | Партиций достаточно для параллелизма консьюмеров; compaction-лаг мониторить (`delete.retention.ms` ≥ worst-case backfill-времени); relay горизонтально — но тогда нужен claim/lock строк outbox чтобы не дублировать публикацию |

### Scaling Priorities

1. **Первый bottleneck:** relay-throughput (poll+publish). Решение: батч-poll, idempotent producer с `acks=all` + параллелизм по партициям. **Число партиций задать с запасом сразу** — увеличение позже ломает `key→partition` (entityID попадёт в другую партицию → порядок per-entity ломается для старых ключей).
2. **Второй bottleneck:** compaction не успевает за write-rate → раздувание `*.state`. Решение: мониторить compaction-lag; тюнить `segment.ms`/`min.cleanable.dirty.ratio`. И `delete.retention.ms` ≥ максимального времени backfill нового домена, иначе tombstone'ы пропадут до прочтения.

---

## Anti-Patterns

### Anti-Pattern 1: Публикация в Kafka внутри UoW-транзакции (dual-write)

**What people do:** В use case после `Save` сразу зовут `producer.Produce` (внутри или сразу после txn-функции).
**Why it's wrong:** Запись в БД и публикация в брокер не атомарны — упадёт публикация → событие потеряно; упадёт commit после публикации → фантомное событие. Это ровно тот dual-write, ради устранения которого существует outbox. Прямо нарушает канон `architecture.md`.
**Do this instead:** В txn — только `Save` + `outbox.Append`. Публикацию делает **relay** после commit, читая закоммиченные строки.

### Anti-Pattern 2: Полагаться на exactly-once брокера / отсутствие eventId

**What people do:** Считать, что idempotent producer = exactly-once end-to-end, и не класть `eventId` в схему.
**Why it's wrong:** Idempotence продюсера дедуплицирует только ретраи внутри producer-сессии. Рестарт relay между publish и mark-published → повторная отправка (at-least-once). Без `eventId` будущий консьюмер не сможет дедуплицировать → дубли проекций. Нарушает инвариант L2 «идемпотентность потребителей, inbox по eventId». Переэмит схемы задним числом — дорого.
**Do this instead:** `eventId` (+ `version`) в metadata **с первого дня**, даже без консьюмеров. Документировать доставку как at-least-once. Готовить inbox-контракт для будущих доменов.

### Anti-Pattern 3: Жирный `HostUpdated`-дамп или тощий notification

**What people do:** Один универсальный `HostChanged` с полным дампом полей, ЛИБО событие без полезной нагрузки («host X изменился, сходи спроси»).
**Why it's wrong:** Жирный дамп = event-carried-state «пере» (консьюмеры не знают, что именно сменилось; теряется доменный смысл). Тощий = «недо» (заставляет sync-запрос в Inventory → distributed monolith). Оба в анти-паттернах L2.
**Do this instead:** Семантические гранулярные события (`HostHardwareChanged`, `HostReassigned`, `HostDecommissioned`) с полями, релевантными факту. Снапшот текущего состояния — отдельно, в `*.state`-compacted.

### Anti-Pattern 4: Soft-delete-флаг вместо tombstone + история-на-событиях

**What people do:** Хранить удалённый хост в Mongo с `deleted=true`, читать историю из soft-deleted записи; в Kafka не различать decommission/delete.
**Why it's wrong:** Противоречит решению PROJECT.md/SEED-001: история живёт на event-backbone, не в записи; soft-delete провоцирует restore-with-merge (запрещён by design — ложный матч на рециклинге FQDN). Одно `HostChanged` на оба случая стирает разницу «списан» vs «убран».
**Do this instead:** `HostDecommissioned` ≠ `HostDeleted` — разные события. `delete` = факт в `*.events` (навсегда) + tombstone в `*.state`. Историю восстанавливают из `*.events`-лога, не из записи.

### Anti-Pattern 5: Назначать version в relay, а не в агрегате

**What people do:** Relay инкрементит version при публикации.
**Why it's wrong:** Relay эмитит в два топика и может рестартовать — version рассинхронизируется, last-writer-by-version ломается.
**Do this instead:** `version` — поле агрегата, инкремент при доменном изменении (источник истины — UoW-txn). Relay только переносит её в metadata.

---

## Integration Points

### Точки интеграции с существующим каноном (UoW / outbox / relay)

| Точка | Что уже зафиксировано (НЕ менять) | Что добавляется в v3.0 |
|-------|-----------------------------------|------------------------|
| `UnitOfWork.Do` | Транзакционная граница (Mongo-txn), txn в ctx — канон v1.0 | Первая **реализация** (`mongo_uow.go`) |
| `Outbox.Append(ctx, events)` | Append в той же txn — канон | Первая реализация (`mongo_outbox.go`) + Mongo-схема outbox-коллекции |
| `PullEvents()` на агрегате | Механика «агрегат копит → сливает» — канон | Первые доменные события Inventory |
| **relay** | Существует концептуально («отдельный async relay публикует, помечает published») | **Реализация relay→Kafka**: producer, mapper, two-topic emit, mark-published |
| `buf` codegen | Скелет `buf.yaml`/`buf.gen.yaml`, pin v1.36.5 — инертен | Первые `.proto` (events + state) → активирует codegen |

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Kafka (брокер) | Destination relay'я; idempotent producer, `acks=all` | Топики provision'ятся: `*.events` (delete), `*.state` (compact, key=entityID). `delete.retention.ms` ≥ worst-case backfill |
| MongoDB | Outbound (агрегаты + outbox), txn-сессии для UoW | Mongo-транзакции требуют replica set — учесть в локальном/dev окружении |
| Go Kafka client | **franz-go (twmb/franz-go)** | Pure Go (без CGO), **idempotent by default**, лучший по ordering/idempotency per WarpStream. Альтернатива — confluent-kafka-go (CGO/librdkafka) если нужна офиц. поддержка Confluent. Sarama не рекомендован для нового idempotent-продюсера |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| usecases ↔ outbox | через порт `Outbox` (domain), в UoW-txn | Запись агрегата + событий атомарна |
| outbox ↔ relay | relay читает закоммиченные строки (poll), вне txn записи | Развязка через БД = нет dual-write |
| relay ↔ Kafka | idempotent producer, key=entityID | at-least-once; eventId для дедупа на консьюмере |
| Inventory ↔ будущие домены (Search/Analytics/Audit) | **только через Kafka-события** (choreography) | v3.0 продюсер-only; консьюмеры — отдельные milestone'ы; контракт схем фиксируется сейчас |

---

## Build Order (для ROADMAP — с учётом зависимостей)

Порядок выводится из «домен → запись → события → транспорт наружу → проверка», каждый шаг — фундамент следующего:

1. **Доменные события + версия в агрегате.** Сначала `domain/*/events.go` и `version` как поле агрегата (источник истины). Зависит от: доменной модели Inventory (Project/Host). Без этого нечего класть в outbox.
2. **UoW + Outbox-порт и Mongo-реализация.** `UnitOfWork.Do` (Mongo-txn) + `Outbox.Append` в той же txn. Зависит от: (1). Это «запись агрегата + outbox атомарно».
3. **protobuf-схемы (events envelope + state).** `.proto` per aggregate: `HostEvent` (metadata+oneof, **с `eventId`/`version`/`actor`**) и `HostState`. Активирует `buf` codegen. Зависит от: (1) (знать набор событий). Можно частично параллелить с (2).
4. **Топик-дизайн / provisioning.** `*.events` (delete) + `*.state` (compact, key=entityID, `delete.retention.ms` с запасом). Зависит от: (3) (имена/ключи). Инфра-шаг.
5. **relay→Kafka publisher.** poll outbox → mapper(row→protobuf, event + state/tombstone) → franz-go idempotent producer (key=entityID) → mark-published. Зависит от: (2),(3),(4). **Ключевой новый компонент.**
6. **decommission vs delete на потоке.** Развести `HostDecommissioned` (snapshot остаётся) и `HostDeleted` (факт + tombstone в `*.state`). Зависит от: (1),(5).
7. **test-consumer (верификация replay/backfill).** Читает `*.state` с earliest, материализует map (last-writer-by-version, tombstone=delete), проверяет онбординг. Зависит от: (5),(6). Не прод-домен — только quality-gate продюсера.

**Критический зависимостный риск:** `eventId`/`version`/`actor` в схеме (шаг 3) — **до** релиза relay (шаг 5). Если отложить — будущие консьюмеры/Audit потребуют переэмита событий. Forward-compat фиксируется в продюсере, хотя консьюмеров в v3.0 нет.

---

## Соответствие инвариантам и анти-паттернам L2 (quality-gate)

| Инвариант L2 | Как соблюдён в этом дизайне |
|--------------|------------------------------|
| Single identity owner — только Inventory создаёт ID | Только Inventory эмитит `*RegisterED`/`*Created`; снапшот идентичности живёт в `inventory.*.state`; другие домены лишь читают |
| Идемпотентность потребителей (inbox по eventId, at-least-once) | `eventId` в metadata с дня 1; доставка документирована как at-least-once; inbox-контракт готов для будущих доменов |
| Терпимость к out-of-order (версия/seq на сущность) | `version` в агрегате + metadata; last-writer-by-version; partition-by-entityID для порядка внутри сущности |
| ACL на потребителе; осмысленные доменные факты | Семантические события (`HostDecommissioned`, не `HostChanged`); консьюмер маппит в свой язык — анти-паттерн «жирный/тощий» исключён |
| Разделение «факт идентичности» (event) vs «процесс» (saga) | v3.0 эмитит **только факты**; decommission-saga с вето — Orchestrator (будущий домен), здесь не реализуется |
| Replay-safe хендлеры (backfill без side-effects) | test-consumer материализует проекцию из compacted-снапшота без внешних эффектов; прод-консьюмеры — будущее |
| Нет distributed monolith | Inventory→домены только через события; нет sync-команд на провижн проекций; семантические события не требуют back-query |

---

## Sources

- [Kafka Log Compaction — Confluent Documentation](https://docs.confluent.io/kafka/design/log_compaction.html) — HIGH (официальная семантика compaction, tombstone, `delete.retention.ms`, offset-стабильность)
- [Key-Based Retention Using Topic Compaction — Confluent Developer](https://developer.confluent.io/courses/architecture/compaction/) — HIGH
- [Kafka Idempotent Consumer & Transactional Outbox — Lydtech](https://www.lydtechconsulting.com/blog/kafka-idempotent-consumer-transactional-outbox) — HIGH (комбинирование outbox + idempotent consumer; failure modes; уже цитировался в L2-ресёрче)
- [Building Reliable Event-Driven Architectures with Kafka: Outbox, Exactly-Once, Idempotent Consumers — Java Code Geeks](https://www.javacodegeeks.com/2025/09/understanding-event-driven-architectures-kafka-outbox-pattern-and-exactly-once-guarantees.html) — MEDIUM
- [Event Sourcing Patterns with Kafka — Conduktor](https://www.conduktor.io/glossary/event-sourcing-patterns-with-kafka) — MEDIUM (dual-topic: append-лог + compacted snapshot per aggregate)
- [WarpStream — Tuning Kafka clients for Performance/idempotency](https://docs.warpstream.com/warpstream/kafka/configure-kafka-client/tuning-for-performance) — HIGH (franz-go рекомендация для idempotency/ordering)
- [franz-go — producing-and-consuming docs (twmb)](https://github.com/twmb/franz-go/blob/master/docs/producing-and-consuming.md) — HIGH (idempotent by default; record-retry семантика)
- Внутренние: `.planning/L2-ARCHITECTURE.md` (инварианты/анти-паттерны/event-backbone), `.planning/PROJECT.md` (Key Decisions: идентичность, удаление, события), `knowledge/architecture.md` (канон UoW/outbox/relay/PullEvents), SEED-001 (идентичность/матчинг), SEED-002 (actor/initiator forward-compat)

---
*Architecture research for: Event-backbone (Kafka) поверх outbox/relay-канона — Inventory продюсер-этап (milestone v3.0)*
*Researched: 2026-06-26*
