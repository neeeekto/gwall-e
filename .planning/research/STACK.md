# Stack Research

**Domain:** Go-микросервис Inventory (DDD/гексагон, MongoDB) + продюсер-бэкбон событий на Kafka (outbox→relay→Kafka)
**Researched:** 2026-06-26
**Confidence:** HIGH

> **Scope-замечание.** Это SUBSEQUENT-milestone (v3.0). Архитектура зафиксирована в v1.0 и
> **НЕ ре-ресёрчится**: Go 1.24.6, go.work, DDD+гексагон без CQRS-шины, слои
> domain/usecases/query/repositories/api(gRPC)/cron, MongoDB, UnitOfWork (Mongo-txn),
> transactional outbox внутри UoW + отдельный relay, Ginkgo v2 + Gomega + mockery, golangci-lint
> v2 + lefthook + buf. Ниже — **только новые/обновляемые** стек-добавления для (a) реального
> сервиса Inventory и (b) Kafka-продюсера поверх существующего outbox/relay-канона.

---

## Recommended Stack

### Core Technologies (новое / обновляемое для v3.0)

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| **github.com/twmb/franz-go** (`pkg/kgo`) | **v1.21.4** (Jun 24, 2026) | Kafka-продюсер в relay: publish outbox-записей в Kafka, partition-by-key, идемпотентная доставка | Pure-Go (CGO нет → статический бинарь, простой кросс-компайл, простой Docker — критично против confluent-kafka-go/librdkafka). Feature-complete до Kafka 4.2+. **Идемпотентный продюсер включён по умолчанию** (`acks=all` + sticky-key partitioner) — ровно то, что нужно relay-каналу с at-least-once. Производительность на уровне librdkafka. Активная разработка (релиз июнь 2026). |
| **go.mongodb.org/mongo-driver/v2** (`/v2/mongo`) | **v2.7.0** (Jun 17, 2026) | Mongo-клиент: репозитории, query-сервисы, sessions/transactions для UnitOfWork | **v1 формально deprecated в начале 2026** (deprecation-нотис в `go.mongodb.org/mongo-driver`); v2 — единственная активно развиваемая ветка. Сейчас в `services/inventory/go.mod` стоит **v1.17.9 — это должно быть обновлено на v2** до построения репозиториев/UoW (миграция дешевле сейчас, до написания кода, чем потом). API транзакций сменился: callback теперь получает `context.Context` вместо `mongo.SessionContext`. |
| **google.golang.org/protobuf** | latest (1.36.x line) | Сериализация payload доменных событий (и так уже proto для gRPC через buf) | proto уже в каноне (buf). Один формат сериализации для gRPC API и для событий → единый IDL, кодоген, версионирование полей по номерам. **Schema registry на этом этапе НЕ нужен** (см. ниже). |
| **github.com/twmb/franz-go/pkg/kadm** | совместимая с kgo (kadm v1.x) | Админ-операции из кода/тестов: создать **compacted** топик с нужным `cleanup.policy=compact`, проверить топологию топиков | Тот же экосистемный клиент, что и kgo; нужен потому что compacted-снапшот по `entityID` требует топиков с `cleanup.policy=compact` — удобно заводить декларативно из миграции/bootstrap, а не руками. |

### Supporting Libraries (новое для v3.0)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| **github.com/testcontainers/testcontainers-go** | **v0.43.0** (Jun 19, 2026) | Поднять реальные Mongo + Kafka в Ginkgo-интеграционных тестах | Для интеграционных suites (UoW-транзакции, outbox-запись, relay→Kafka end-to-end). Не для чисто доменных unit-тестов (там mockery-моки портов). |
| **…/testcontainers-go/modules/kafka** | (версия модуля = тег ядра) | Kafka-контейнер в **KRaft mode** (без Zookeeper) | `kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithClusterID(...))`. Юзает KRaft из коробки. |
| **…/testcontainers-go/modules/mongodb** | (версия модуля = тег ядра) | Mongo-контейнер для тестов UoW/репозиториев | **ВАЖНО:** транзакции Mongo требуют **replica set**. Дев/тест Mongo нужно поднимать как single-node RS (`--replSet`), иначе `StartTransaction` падает. |
| **github.com/twmb/franz-go/plugin/kprom** | совместимая с kgo | Prometheus-метрики продюсера (lag publish, ошибки, ретраи relay) | Когда понадобится наблюдаемость relay (stuck-rows, duplicate-risk видны через метрики). Можно отложить до первой эксплуатации, но плагин дешёвый. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| **docker-compose (локальный Kafka, KRaft)** | Локальный single-node Kafka без Zookeeper для ручного дева/отладки relay | Образ `confluentinc/confluent-local` или `apache/kafka` (нативный KRaft). Один брокер, `KAFKA_PROCESS_ROLES=broker,controller`. Compacted-топик(и) для снапшота по `entityID`. |
| **buf** (уже в каноне v1.0) | Кодоген proto для событий (тот же pipeline, что gRPC) | Добавить proto-пакет под доменные события рядом с gRPC-контрактами; lint/breaking-check buf'ом покрывает и события. |
| **mockery** (уже в каноне v1.0) | Моки портов (`EventPublisher`/relay-порт, репозитории, UoW) для unit-тестов | Доменные/usecase-тесты мокают порт публикации; реальный franz-go только в интеграционных тестах с testcontainers. |

## Installation

```bash
# из services/inventory (модуль вне go.work → GOWORK=off)
GOWORK=off go get github.com/twmb/franz-go@v1.21.4
GOWORK=off go get github.com/twmb/franz-go/pkg/kadm@latest
GOWORK=off go get go.mongodb.org/mongo-driver/v2@v2.7.0   # замена v1.17.9

# proto (если ещё не подтянут транзитивно через gRPC)
GOWORK=off go get google.golang.org/protobuf@latest

# dev/test
GOWORK=off go get github.com/testcontainers/testcontainers-go@v0.43.0
GOWORK=off go get github.com/testcontainers/testcontainers-go/modules/kafka@v0.43.0
GOWORK=off go get github.com/testcontainers/testcontainers-go/modules/mongodb@v0.43.0

# наблюдаемость продюсера (можно отложить)
GOWORK=off go get github.com/twmb/franz-go/plugin/kprom@latest
```

> **Уже в репозитории (не трогаем):** Ginkgo v2.23.4, Gomega v1.38.0 (в `pkg/go.mod`),
> `github.com/google/uuid v1.6.0` (генерация постоянного внутреннего `ID` хоста — см. SEED-001).

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| **franz-go** | **confluent-kafka-go** (официальный, librdkafka) | Только если нужна максимально зрелая транзакционная EOS-реализация И CGO допустим. Для gwall-e CGO — минус (статика/кросс-компайл/Docker), а продюсер-only at-least-once не требует librdkafka-зрелости → не оправдано. |
| **franz-go** | **segmentio/kafka-go** | Только для простого pub/sub без идемпотентного продюсера/транзакций. У него **нет** idempotent-producer guarantees, sync-Writer медленный, тестировался против Kafka ≤2.7.1. Не подходит для надёжного relay. |
| **franz-go** | **IBM/sarama** | Широко распространён, но «больше всего подводных камней» (community consensus); franz-go — современная замена. Не выбирать для нового кода. |
| **mongo-driver/v2** | **mongo-driver v1** | Не использовать в новом коде — v1 deprecated (2026). Остаться на v1 можно лишь если миграция блокирует, но здесь код пишется с нуля → сразу v2. |
| **testcontainers Kafka module** | **testcontainers Redpanda module** | Если позже понадобится SASL/SCRAM в тестах — у Kafka-модуля известны проблемы с SASL-конфигом; Redpanda (Kafka-API-совместим, KRaft-like, без JVM) — обходной путь. Для plaintext-дева Kafka-модуля достаточно. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| **Schema Registry (Confluent/apicurio)** на этапе v3.0 | Продюсер-only, консьюмеров нет → главная ценность SR (контракт producer↔consumer, forward-compat) **не к чему применить**. Лишний операционный компонент. | proto-IDL под версионным контролем (buf) + дисциплина «только backward-совместимые правки полей по номерам». Завести SR позже, когда появится первый консьюмер (Search/Analytics/Audit) — рекомендуемый режим **BACKWARD_TRANSITIVE** для protobuf. |
| **confluent-kafka-go / librdkafka** | CGO → ломает статические бинари, усложняет кросс-компайл и Docker-сборку; не оправдано для продюсер-only. | franz-go (pure-Go). |
| **segmentio/kafka-go** для relay | Нет idempotent-producer; at-least-once без идемпотентности → риск дублей от ретраев продюсера на брокере. | franz-go (идемпотентность по умолчанию). |
| **mongo-driver v1** в новом коде | Deprecated 2026; пакеты `primitive`/`gridfs`/`bsonrw` удалены в v2, API транзакций изменился. | mongo-driver/v2. |
| **Exactly-once / Kafka-транзакции (EOS)** в relay сейчас | Канон зафиксировал **at-least-once** (инвариант L2: «не exactly-once брокера», идемпотентность — на стороне консьюмера). EOS-транзакции продюсера усложняют relay без выгоды, пока консьюмеров нет. | Идемпотентный продюсер (`acks=all`) + dedup по `eventId` на будущих консьюмерах (inbox). |
| **Consumer/inbox-инфра** (consumer groups, dedup-таблицы) | v3.0 — **полный продюсер, консьюмеров нет** (PROJECT.md, L2). | Отложить до первого consumer-домена (SEED-002 Audit, Search). |
| **CDC/Debezium** для relay | Канон — **polling-relay из outbox внутри UoW** (не log-based CDC). Debezium тянет Kafka Connect + отдельную операционку. | Существующий polling-relay (канон v1.0) → franz-go produce. |

## Механика relay → Kafka (интеграция с outbox-каноном)

> Канон v1.0 **не меняется**: доменные события собираются `PullEvents`, пишутся в outbox в той же
> Mongo-транзакции UnitOfWork (нет dual-write), отдельный relay вычитывает неопубликованные записи.
> Kafka — это лишь **destination** публикации relay.

1. **Запись (write-side).** Use case → `UnitOfWork`: бизнес-изменение агрегата **и** outbox-запись
   коммитятся одной Mongo-транзакцией. Outbox-строка несёт: `eventId` (uuid), `entityID`
   (ключ партиционирования/компакции), сериализованный proto-payload, `version/seq` сущности,
   metadata (`actor/initiator`), `publishedAt=null`.
2. **Relay (polling).** Relay сканирует неопубликованные строки (партиальный индекс по
   `publishedAt`), сохраняя **порядок по `entityID`** (одна сущность — упорядоченно).
3. **Produce (franz-go).** Для каждой строки — `kgo.Record{Key: entityID, Value: protoBytes,
   Topic: ...}`:
   - **Key = `entityID`** → sticky-key partitioner кладёт все события одной сущности в **одну
     партицию** → порядок per-entity сохранён; этот же ключ — ключ **log-compaction**
     (compacted-снапшот «последнее состояние идентичности» по `entityID`).
   - **Идемпотентный продюсер по умолчанию** (`acks=all`, sticky-key) → брокер дедуплицирует
     ретраи продюсера в рамках сессии.
   - **Ретраи — не ограничивать** (franz-go ретраит «вечно» и безопасно для sequence-номеров;
     лимит ретраев + идемпотентность может дать ложное «data loss»).
4. **Mark published.** После успешного ack от Kafka relay проставляет `publishedAt`. Если relay
   упал между produce и mark → повторная публикация (**at-least-once**) — допустимо по инварианту;
   дедуп — забота будущих консьюмеров по `eventId` (inbox), не v3.0.
5. **Топики — compacted.** Через `kadm` (или bootstrap/миграцию) топик(и) событий заводятся с
   `cleanup.policy=compact` (по необходимости `compact,delete`), партиции — по `entityID`.
   Compaction даёт «снапшот идентичности» для бесплатного онбординга будущего домена + replay
   с `offset=earliest` для backfill.

**Семантический payload (инвариант L2):** осмысленные доменные события (`HostRegistered`,
`HostDecommissioned`, `OwnershipChanged`), не тощие `HostChanged`; не «жирные `EntityUpdated`-дампы».
proto-сообщение на каждый тип события + отдельный compacted-снапшот по `entityID`.

## MongoDB v2: специфика для UnitOfWork и снапшота

- **Транзакции = sessions.** `client.StartSession()` → `session.WithTransaction(ctx, fn)`
  (рекомендуется: авто-ретрай transient-ошибок) **или** ручной
  `mongo.WithSession`+`StartTransaction`/`CommitTransaction`. В v2 callback получает
  **`context.Context`** (не `mongo.SessionContext` из v1) — порт `UnitOfWork` пробрасывает
  session-aware `ctx`; репозитории берут транзакцию из `ctx` (как и зафиксировано в каноне).
- **Replica set обязателен** для транзакций — и в проде, и в тестах (single-node RS с `--replSet`).
  testcontainers Mongo-модуль нужно конфигурировать как RS.
- **Read/Write concern** для согласованности outbox: транзакцию заводить с
  `writeconcern.Majority()` (+ при нужде `readconcern.Snapshot()`/`Majority`).
- **Compacted-снапшот (Mongo-сторона).** Сам снапшот «последнего состояния» по `entityID` — это
  семантика Kafka log-compaction; в Mongo это обычный текущий стейт агрегата + событие в outbox.
  Никакой отдельной снапшот-коллекции под Kafka-компакцию не требуется.
- **Топология `connections` (cross-refs)** — обычные Mongo-документы со ссылками; read-model
  «что зависит от X» строится query-сервисом напрямую (CQRS-lite, канон). Driver-специфики нет.

## Stack Patterns by Variant

**Пока консьюмеров нет (текущий v3.0):**
- franz-go только как **продюсер** (`kgo` produce + `kadm` для топиков). Без consumer-групп.
- Без schema registry: proto + buf breaking-check как контракт-гейт.
- Идемпотентный продюсер (default), at-least-once, дедуп откладываем.

**Когда появится первый консьюмер (Search/Analytics/Audit — будущие milestones):**
- Добавить **Schema Registry** (режим **BACKWARD_TRANSITIVE** для protobuf).
- Добавить **inbox/dedup по `eventId`** на консьюмере (идемпотентность — на потребителе).
- franz-go consumer-группы; ACL «событие → локальный ubiquitous language» (инвариант L2).

**Если тестам понадобится SASL:**
- Перейти на testcontainers **Redpanda**-модуль (Kafka-API-совместим, KRaft-like) — у Kafka-модуля
  известны проблемы с SASL-конфигом.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| franz-go v1.21.4 | Kafka 0.8.0 … 4.2+ | Покрывает любой реалистичный брокер; KRaft-режим Kafka — без оговорок. (Прим.: ветка v1.19.0 имела баг с pre-4.0 брокерами, исправлено в последующих патчах — v1.21.4 чист.) |
| franz-go `pkg/kgo` v1.21.4 | `pkg/kadm` v1.x | Сабмодули версионируются независимо; брать совместимый тег kadm к kgo. |
| mongo-driver/v2 v2.7.0 | Go ≥ 1.19 (репо), MongoDB ≥ 4.4 | v2.7 требует **MongoDB 4.4+** (v2.6 был последним с поддержкой 4.2). Go 1.24.6 проекта — ок. |
| testcontainers-go v0.43.0 | Go-модули kafka/mongodb того же тега | Модули версионируются вместе с ядром (один тег `v0.43.0`). Kafka-модуль `Run(...)` (не deprecated `RunContainer`). |
| Ginkgo v2.23.4 / Gomega v1.38.0 | testcontainers-go v0.43.0 | Совместимы; testcontainers — в `BeforeSuite`/`AfterSuite` для подъёма/сноса контейнеров. |

## Sources

- pkg.go.dev `github.com/twmb/franz-go` — latest **v1.21.4**, Jun 24 2026 (HIGH, official registry)
- pkg.go.dev `go.mongodb.org/mongo-driver/v2` — latest **v2.7.0**, Jun 17 2026; Go ≥1.19, MongoDB ≥4.4 (HIGH)
- pkg.go.dev `github.com/testcontainers/testcontainers-go` — **v0.43.0**, Jun 19 2026 (HIGH)
- golang.testcontainers.org/modules/kafka — KRaft `kafka.Run(...)`, `confluentinc/confluent-local` (HIGH, official docs)
- golang.testcontainers.org/modules/redpanda — SASL-fallback альтернатива (MEDIUM)
- github.com/twmb/franz-go docs (producing-and-consuming.md) — идемпотентность по умолчанию, `acks=all`, sticky-key partitioner, ретраи-без-лимита (HIGH, maintainer docs)
- mongodb.com/docs/drivers/go (v2 transactions) + pkg.go.dev v2/mongo — `WithSession`/`WithTransaction`, callback на `context.Context`, replica-set требование (HIGH, official)
- mongodb.com release notes — v1 deprecated в 1.17.8 (2026), миграция на /v2 (HIGH, official)
- docs.confluent.io schema-registry (serdes-protobuf, schema-evolution) — SR не нужен без консьюмеров; BACKWARD_TRANSITIVE для protobuf (MEDIUM-HIGH, vendor docs)
- microservices.io transactional outbox + AutoMQ/SoByte Go-Kafka-client comparisons — relay at-least-once, дедуп на консьюмере, CGO-трейдофф клиентов (MEDIUM, community/vendor, cross-checked)

---
*Stack research for: Go Inventory microservice + Kafka producer event-backbone (gwall-e v3.0)*
*Researched: 2026-06-26*
