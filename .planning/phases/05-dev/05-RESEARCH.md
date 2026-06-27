# Phase 5: Dev-инфра и стек — Research

**Researched:** 2026-06-27
**Domain:** Go dev-инфра (go.work, build/lint/test recipes) + локальный стенд (Kafka KRaft + Mongo single-node RS через docker-compose) + bootstrap топиков на `franz-go/kadm` + миграция `mongo-driver` v1→v2 + интеграционные тесты на `testcontainers-go` (Ginkgo v2/Gomega/mockery)
**Confidence:** HIGH

## Summary

Phase 5 — нулевой шаг v3.0: готовое окружение и обновлённый стек **до** первого доменного
кода. Все архитектурные решения уже залочены в `05-CONTEXT.md` (D-01…D-15) и
`.planning/research/STACK.md`; этот ресёрч **не переоткрывает** их, а добывает конкретные
идиоматические Go-паттерны и точные сигнатуры API для четырёх load-bearing задач:
(1) свап `go.mod` на `mongo-driver/v2` + тонкий connection-helper; (2) единый Go-модуль
топологии топиков с bootstrap-функцией на `kadm` + тонкий CLI; (3) docker-compose c Kafka
(`confluentinc/confluent-local`, KRaft) + Mongo `mongo:7` single-node RS с healthcheck-driven
`rs.initiate()`; (4) интеграционный smoke-тест на `testcontainers-go` (модули kafka+mongodb),
плюс mockery-обвязка и починка DOC-02.

Все версии стека верифицированы по Go-proxy в этой сессии (HIGH): `franz-go v1.21.4`,
`pkg/kadm v1.18.0`, `mongo-driver/v2 v2.7.0`, `testcontainers-go v0.43.0` (+ модули kafka/mongodb
того же тега), `mockery v3.7.1`. Сигнатуры `kadm.CreateTopics`, `kafka.Run/Brokers`,
`mongodb.Run/WithReplicaSet/ConnectionString`, `mongo.Connect`(v2), `kgo.NewClient/Ping`
извлечены из официальных pkg.go.dev / исходников тегов и приведены ниже verbatim.

**DOC-02 эмпирически решён в этой сессии:** `cd services/audit && go build ./...` падает
`build output "cmd" already exists and is a directory` (exit 1). Кандидат `go build ./cmd`
**тоже падает тем же сообщением** (вопреки записи в STATE.md). Зелёный exit 0 без артефакта
дают **`go vet ./...`** (рекомендуется — каноническая проверка без бинаря) и
`go build -o /dev/null ./...`. См. § Validation Architecture / § DOC-02.

**Primary recommendation:** один Go-модуль топологии топиков (константы `+` bootstrap-функция
на `kadm`) — единственный источник истины, который зовут И тонкий CLI в `cmd/` (для compose
через make-таргет), И интеграционные тесты напрямую; миграция Mongo = свап `go.mod` + helper
на `mongo.Connect(options.Client().ApplyURI(uri).SetWriteConcern(writeconcern.Majority()))`
(в v2 `Connect` без `ctx`); compose-Mongo = `mongo:7 --replSet rs0` + healthcheck
`try{rs.status()}catch{rs.initiate(...)}`, клиент с `directConnection=true`; DOC-02 → `go vet ./...`.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Граница inventory / go.work**
- **D-01:** inventory становится **полноправным членом `go.work`** (остаётся в `use`-блоке).
  Канон `GOWORK=off` **отменяется**; сборка/вет — общий `go build ./...` / `go vet ./...` из
  корня workspace.
- **D-02:** pre-push **больше не исключает** inventory (снять намеренное исключение из `lefthook.yml`).
- **D-03:** Осознанная цена D-01/D-02: WIP-ошибка компиляции в inventory ломает сборку **всего**
  workspace → инвариант «inventory всегда компилируется на каждом коммите».
- **D-04:** Переписать под новое решение каноны: `knowledge/build.md` (раздел `GOWORK=off` +
  строка про pre-push), `knowledge/structure.md` (раздел inventory вне workspace),
  `knowledge/boundaries.md`, **и формулировку Success Criterion 1 в ROADMAP** (убрать
  «с `GOWORK=off`»). In-scope Phase 5, не deferred.

**Локальный стенд (compose ↔ testcontainers)**
- **D-05:** docker-compose = ручной дев-стенд, testcontainers = изолированные эфемерные
  интеграционные тесты. Не объединяем тесты через compose-модуль.
- **D-06:** Дрейфующее знание (имена + cleanup-политики топиков, число партиций, версии образов)
  живёт в **одном Go-модуле**: общая bootstrap-функция на `kadm` + константы. И compose-провижн,
  и тесты зовут **эту же** функцию — единственный источник истины топологии топиков.
- **D-07:** Kafka-образ — `confluentinc/confluent-local` (паритет дев≈тест), KRaft из коробки, один брокер.
- **D-08:** Mongo — `mongo:7` как **single-node replica set** (`--replSet` + `rs.initiate()`);
  `mongo-driver/v2` v2.7 требует MongoDB ≥ 4.4, транзакции требуют RS.

**Provisioning топиков**
- **D-09:** Провижн — тонкий **Go CLI в `cmd/`**, зовёт общую bootstrap-функцию (D-06); дёргается
  `make`-таргетом после подъёма compose; тесты зовут ту же функцию напрямую.
- **D-10:** В Phase 5 заводим **только** `inventory.host.events` + `inventory.host.state`. Список
  агрегатов — **data-driven** (константа/конфиг): добавление `project`/`module`/`location` в
  Phase 6+ = одна строка.
- **D-11:** Число партиций — **параметр** bootstrap'а; дев/тест-дефолт **6** (гоняет sticky-key
  partitioner на >1 партиции). Prod-число — **отложено**.
- **D-12:** Политики топиков (залочено каноном): `*.events` → `cleanup.policy=delete` (длинный
  retention, immutable-история фактов); `*.state` → `cleanup.policy=compact` +
  `delete.retention.ms ≥ 24h`; Kafka message key = внутренний `ID` (НЕ FQDN/INV/MAC).

**Миграция mongo-driver/v2 + интеграционные тесты**
- **D-13:** `internal/` пуст, кода на v1 нет → «миграция» = свап `go.mod` (`mongo-driver v1.17.9`
  → `mongo-driver/v2 v2.7.0`, удалить v1) + доказать v2 рабочим подключением к Mongo RS
  интеграционным тестом.
- **D-14:** Глубина скаффолда — **тонкий connection-helper**: Mongo client-фабрика (RS-aware,
  `writeconcern.Majority`) + health-ping. Эталон «как подключаемся» для Phase 6/7.
  **Без** repository/UoW/Outbox — это Phase 7.
- **D-15:** Интеграционные тесты изолированы **build-tag `integration`** + `make test-integration`.
  `go test ./...` и pre-push гоняют **только unit** (быстро, без контейнеров); testcontainers-тесты
  — за тегом / в CI. Критично, т.к. после D-02 pre-push включает inventory.

### Claude's Discretion

- **DOC-02 (SC5):** заменить audit-рецепт в `knowledge/build.md` на команду с exit 0. Кандидаты:
  `go vet ./...` (предпочтительно) / `go build ./cmd` / `go build -o /dev/null ./...`. Финальную
  форму executor верифицирует эмпирически (SC5 требует реального exit 0); привести рецепты к
  новому workspace-канону (D-01).
- **mockery (SVC-06):** провод `.mockery.yaml` + `make generate-mocks`, доказать smoke одним
  throwaway/example-интерфейсом (реальных доменных портов ещё нет — появятся в Phase 6/7).
- Конкретная структура Makefile-таргетов, healthcheck'и compose, имена переменных/констант,
  лэйаут пакетов хелпера и теста — на усмотрение планнера/executor'а в рамках решений выше.

### Deferred Ideas (OUT OF SCOPE)

- **Prod-число партиций Kafka** (под ~150k парк) — ops-решение; фиксируется при первом консьюмере.
  В Phase 5 — только configurable-параметр с дев-дефолтом.
- **Schema Registry** — вводится с первым консьюмером (BACKWARD_TRANSITIVE для protobuf, SEED-002).
  Вне v3.0.
- **`kprom` / Prometheus-метрики продюсера** — дешёвый плагин franz-go, но наблюдаемость relay
  откладываем до Phase 8+.
- **Redpanda-модуль testcontainers** — fallback только если тестам понадобится SASL/SCRAM. Для
  plaintext-дева не нужен.
- **CI** (полноценный pipeline) — за рамками Phase 5; enforcement через локальный lefthook.
  D-01/D-02/D-15 заложены так, чтобы будущий CI лёг без переделки.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| **SVC-05** | Персистентность — MongoDB через mongo-driver **v2** (миграция с v1 до написания репозиториев) | § Standard Stack (mongo-driver/v2 v2.7.0), § Code Examples / Mongo connection-helper (v2 `Connect` без `ctx`, `writeconcern.Majority`, RS-aware URI), § Pitfall «RS host resolution / directConnection» |
| **SVC-06** | Тесты — Ginkgo v2 + Gomega + mockery; интеграционные через testcontainers (Kafka KRaft + Mongo single-node RS) | § Code Examples / testcontainers smoke (kafka+mongodb v0.43.0 Run/Brokers/ConnectionString), § mockery v3 `.mockery.yaml`, § Validation Architecture (build-tag `integration`) |
| **SVC-07** | Dev-инфра — docker-compose (Kafka KRaft + Mongo RS) + bootstrap провижна топиков (`*.events`/`*.state` с нужной cleanup-policy) | § Code Examples / kadm bootstrap (CreateTopics + cleanup.policy/delete.retention.ms config map), § docker-compose (confluent-local + mongo:7 RS healthcheck), § Architecture Patterns (single-source топология) |
| **DOC-02** | Починить `build.md` audit-рецепт (carry-over из v1.0) | § DOC-02 (эмпирически: `go vet ./...` exit 0; `go build ./cmd` падает) |

> SC1 (inventory собирается с mongo-driver/v2, v1 удалён, build/vet зелёные) маппится на SVC-05
> + D-01 (workspace-build); SC2/SC3 → SVC-07; SC4 → SVC-06; SC5 → DOC-02.
</phase_requirements>

## Project Constraints (from CLAUDE.md / AGENTS.md / knowledge)

Каноны имеют ту же силу, что локированные решения. Планнер обязан их соблюдать:

- **Язык:** комментарии в коде и доменная терминология — **на русском**; имена идентификаторов
  (типы, функции, пакеты) — **на английском**; комментарии **в тестах** — **на английском**
  (`knowledge/style.md`, `knowledge/testing.md`). [CITED: knowledge/style.md]
- **Типизированные ID:** `type HostID string` вместо голого `string` для идентификаторов
  агрегатов. [CITED: knowledge/style.md]
- **Ошибки:** sentinel + `%w`-обёртка (`errorlint` — hook). [CITED: knowledge/style.md]
- **Тесты:** Ginkgo v2 (`RegisterFailHandler(Fail)` + `RunSpecs`), dot-imports
  `. "github.com/onsi/ginkgo/v2"` / `. "github.com/onsi/gomega"`; `*_test.go` рядом с кодом;
  Gomega-ассерты (не голый `t.Fatalf`); mockery через `GinkgoT()` (не `t`), `testify/mock` —
  обычный (не dot) импорт. [CITED: knowledge/testing.md]
- **Слои:** канон `domain / usecases / query / repositories / api / cron` + `app` (composition
  root) + `cmd` (main). `domain` наружу не импортирует (depguard dormant). Connection-helper
  (D-14) ляжет в `repositories`/infra-пакет, но БЕЗ repo/UoW-реализаций.
  [CITED: knowledge/architecture.md]
- **MUST NOT:** возрождать CQRS-шину / `TxManager`/`tx.go` (depguard biting на `pkg/mediatr`).
  [CITED: knowledge/architecture.md]
- **No-phantom:** документировать только реально проверенные команды; buf/proto **не**
  активируется в Phase 5 (нет `.proto`). [CITED: knowledge/boundaries.md, knowledge/build.md]
- **Стале-файлы неавторитетны:** корневые `README.md`, **`Makefile`** (корневой `go.mod` —
  «rotten leftover», go 1.23.6, вне go.work — НЕ трогать, НЕ добавлять `tool`-блок),
  `docker-compose.yml` (если есть) — источник истины только `knowledge/` + `go.work`.
  [VERIFIED: чтение корневого go.mod — `module github.com/gwall-e`, `go 1.23.6`]
- **Пиннинг тулинга:** новые dev-инструменты (mockery) пиннятся версией по паттерну корневого
  `Makefile` (`*_VERSION` + `go install ...@version`). [CITED: Makefile]

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Сборка/линт/тест workspace | Build/Tooling (`go.work`, `lefthook.yml`, `knowledge/build.md`) | — | D-01/D-02: inventory — полноправный член workspace; рецепты — канон в build.md |
| Топология топиков (имена, политики, партиции) | Shared Go-модуль топологии (константы + `kadm` bootstrap) | — | D-06: единственный источник истины, зовут и CLI, и тесты |
| Провижн топиков (запуск) | `cmd/` CLI → make-таргет (для compose) | Integration test (зовёт функцию напрямую) | D-09: тонкий CLI оборачивает bootstrap-функцию |
| Локальный дев-стенд (Kafka+Mongo) | docker-compose (ручной) | — | D-05: compose = ручной стенд, отдельно от тестов |
| Эфемерная инфра для тестов | testcontainers-go (kafka+mongodb модули) | — | D-05/D-15: изолированные интеграционные тесты за build-tag |
| Подключение к Mongo (RS-aware) | `repositories`/infra connection-helper | — | D-14: тонкий helper (client-фабрика + ping), без UoW |
| Мокинг портов | mockery (`.mockery.yaml` + codegen) | — | SVC-06: smoke на throwaway-интерфейсе (доменных портов ещё нет) |
| Audit build-рецепт (DOC-02) | Build/Tooling (`knowledge/build.md`) | — | carry-over долг v1.0; рецепт с exit 0 |

## Standard Stack

> Все версии верифицированы по Go module proxy (`proxy.golang.org/.../@latest`) в этой сессии
> (2026-06-27). Это **дополнение** к зафиксированному в v1.0 стеку (Go 1.24.6, Ginkgo v2.23.4 /
> Gomega v1.38.0 в `pkg/go.mod`, `google/uuid v1.6.0` уже в inventory).

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `go.mongodb.org/mongo-driver/v2` | **v2.7.0** (2026-06-17) | Mongo-клиент (connection-helper; в Phase 6/7 — repo/UoW) | v1 deprecated в 2026; v2 — единственная активная ветка. Замена `v1.17.9` в `services/inventory/go.mod`. [VERIFIED: go proxy `@latest`=v2.7.0] |
| `github.com/twmb/franz-go` | **v1.21.4** (2026-06-24) | `pkg/kgo` — Kafka-клиент (smoke-ping в Phase 5; produce в relay Phase 8) | pure-Go (нет CGO), идемпотентный продюсер по умолчанию. [VERIFIED: go proxy `@latest`=v1.21.4] |
| `github.com/twmb/franz-go/pkg/kadm` | **v1.18.0** (2026-04-21) | Админ-операции: `CreateTopics` с `cleanup.policy`/`delete.retention.ms`/партициями | Тот же экосистемный клиент, что kgo; bootstrap топиков декларативно из кода. [VERIFIED: go proxy subdir `@latest`=v1.18.0] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/testcontainers/testcontainers-go` | **v0.43.0** (2026-06-19) | Эфемерные контейнеры в Ginkgo-suite | Только интеграционные тесты (build-tag `integration`), не unit. [VERIFIED: go proxy `@latest`=v0.43.0] |
| `…/testcontainers-go/modules/kafka` | **v0.43.0** | Kafka KRaft контейнер (`confluentinc/confluent-local`) | `kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", ...)`. [VERIFIED: go proxy subdir тег `modules/kafka/v0.43.0`] |
| `…/testcontainers-go/modules/mongodb` | **v0.43.0** | Mongo single-node RS контейнер | `mongodb.Run(ctx, "mongo:7", mongodb.WithReplicaSet("rs0"))`. [VERIFIED: go proxy subdir тег `modules/mongodb/v0.43.0`] |
| `github.com/vektra/mockery/v3` | **v3.7.1** (2026-06-11) | Codegen моков портов | `.mockery.yaml` + `make generate-mocks`; smoke на throwaway-интерфейсе. [VERIFIED: go proxy `v3/@latest`=v3.7.1]. См. § Open Question о v2 vs v3. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `mongo:7` single-node RS | `mongo:6` | testcontainers-doc дефолт для модуля — `mongo:6`; но D-08 залочил `mongo:7` (паритет дев), v2.7 требует ≥4.4 — `mongo:7` ОК [CITED: golang.testcontainers.org/modules/mongodb] |
| `confluentinc/confluent-local` | `apache/kafka` (нативный KRaft) | D-07 залочил confluent-local (паритет с дефолтом kafka-модуля testcontainers); apache/kafka не выбирать без причины |
| testcontainers kafka-module | Redpanda-module | Только если тестам понадобится SASL/SCRAM (известные проблемы SASL у kafka-модуля). Plaintext-дев — kafka-модуля достаточно (Deferred) |
| `mockery/v3` | `mockery/v2` (v2.53.6) | v3 — текущая major (yaml-config, `.SrcPackageName`/`.SrcPackagePath`, mocks рядом с кодом по умолчанию). `knowledge/testing.md` уже описывает v3-expecter-API → **выбрать v3** |

**Installation** (из каталога `services/inventory`; **с D-01 `go.work` активен** —
`GOWORK=off` больше НЕ нужен, в отличие от устаревшего STACK.md §Installation):
```bash
# core
go get go.mongodb.org/mongo-driver/v2@v2.7.0        # замена v1.17.9 (удалить v1 из go.mod)
go get github.com/twmb/franz-go@v1.21.4
go get github.com/twmb/franz-go/pkg/kadm@v1.18.0
# dev/test (для тестового кода — попадут в require при первом импорте)
go get github.com/testcontainers/testcontainers-go@v0.43.0
go get github.com/testcontainers/testcontainers-go/modules/kafka@v0.43.0
go get github.com/testcontainers/testcontainers-go/modules/mongodb@v0.43.0
# mockery — пиннится в Makefile (НЕ корневой go.mod tool-блок — он rotten):
#   go install github.com/vektra/mockery/v3@v3.7.1
go mod tidy
```

> **Внимание (D-13):** удалить старую `go.mongodb.org/mongo-driver v1.17.9` из `require`.
> v1-зависимости (`xdg-go/*`, `montanaflynn/stats`, `golang/snappy`) — indirect от v1-драйвера;
> после свапа `go mod tidy` их пересоберёт под v2. Пакеты `primitive`/`gridfs`/`bsonrw` в v2
> удалены — но в inventory кода на v1 нет (`internal/` пуст), поэтому import-правок нет.

## Package Legitimacy Audit

> Ecosystem — Go (модули). Seam `package-legitimacy check` поддерживает только npm/pypi/crates —
> поэтому верификация проведена **напрямую через Go module proxy** (`proxy.golang.org/<mod>/@latest`),
> что для Go является авторитетным источником (тот же proxy, что использует `go get`).

| Package | Registry | Latest (proxy) | Source Repo | Verdict | Disposition |
|---------|----------|----------------|-------------|---------|-------------|
| `go.mongodb.org/mongo-driver/v2` | Go proxy | v2.7.0 (2026-06-17) | github.com/mongodb/mongo-go-driver | OK | Approved |
| `github.com/twmb/franz-go` | Go proxy | v1.21.4 (2026-06-24) | github.com/twmb/franz-go | OK | Approved |
| `github.com/twmb/franz-go/pkg/kadm` | Go proxy | v1.18.0 (2026-04-21) | github.com/twmb/franz-go (subdir) | OK | Approved |
| `github.com/testcontainers/testcontainers-go` | Go proxy | v0.43.0 (2026-06-19) | github.com/testcontainers/testcontainers-go | OK | Approved |
| `…/modules/kafka` | Go proxy | v0.43.0 | testcontainers-go (subdir tag) | OK | Approved |
| `…/modules/mongodb` | Go proxy | v0.43.0 | testcontainers-go (subdir tag) | OK | Approved |
| `github.com/vektra/mockery/v3` | Go proxy | v3.7.1 (2026-06-11) | github.com/vektra/mockery | OK | Approved |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

Все пакеты — зрелые, с официальными репозиториями, версии из того же proxy, что `go get`.
Provenance: имена/версии получены из STACK.md (внутренний ресёрч) **и** перепроверены против
go-proxy в этой сессии → `[VERIFIED: go proxy]`.

## Architecture Patterns

### System Architecture Diagram

```text
                    ┌─────────────────────────────────────────────┐
                    │  shared topology module (D-06)               │
                    │  • const: aggregates=[host], suffixes=       │
                    │      {events:cleanup=delete, state:compact}  │
                    │  • Bootstrap(ctx, admCl, cfg{partitions=6})  │
                    │      → kadm.CreateTopics(...)                 │
                    └───────────────┬───────────────┬──────────────┘
                       зовёт напрямую│               │зовёт напрямую
                                     │               │
         ┌───────────────────────────▼──┐      ┌─────▼────────────────────────┐
manual → │ cmd/ CLI (D-09)               │      │ integration test (build-tag  │
dev      │  parse flags → Bootstrap(...) │      │  `integration`, D-15)        │
         │  ← make topics / make dev-up  │      │  testcontainers kafka+mongo  │
         └───────────────┬──────────────┘      │  → Bootstrap(...) → assert   │
                         │ KAFKA_BROKERS env    └───────┬──────────────┬───────┘
              ┌──────────▼──────────┐                   │              │
              │ docker-compose (D-05,D-07,D-08)         │ ephemeral    │ ephemeral
              │  ┌─────────────────┐  ┌──────────────┐  │ Kafka        │ Mongo RS
              │  │ confluent-local │  │ mongo:7      │  │ (Brokers())  │ (ConnString())
              │  │ KRaft, 1 broker │  │ --replSet rs0│  │              │
              │  │ :9092           │  │ healthcheck: │  ▼              ▼
              │  └─────────────────┘  │ rs.initiate  │ kgo.NewClient   mongo.Connect (v2)
              │                       └──────────────┘ .Ping(ctx)      helper.Connect→Ping
              └─────────────────────────────────────┘ (smoke SC4)     (smoke SC1/SC4)
```

Поток для основного use-case Phase 5 (валидация SC2-SC4): `Bootstrap` — единая функция,
вызываемая из CLI (для ручного compose-стенда) и из теста (для эфемерных контейнеров);
ни тест не зависит от compose, ни compose от теста — общий только модуль топологии (D-05/D-06).

### Recommended Project Structure
```text
services/inventory/
├── cmd/
│   ├── main.go              # существующий стаб (func main(){return}) — заменить/дополнить
│   └── topics/ (или флаг)   # тонкий CLI провижна (D-09): parse env/flags → Bootstrap(...)
├── internal/
│   ├── kafka/topology/      # D-06: const топиков + cleanup-политики + Bootstrap(kadm) — single source
│   └── mongo/ (в repositories-слое) # D-14: connection-helper (Connect+Ping), БЕЗ repo/UoW
├── go.mod                   # mongo-driver/v2, franz-go, kadm (+ test-deps testcontainers)
└── <pkg>_integration_test.go  # build-tag `integration` (D-15): testcontainers smoke
docker-compose.yml           # корень или infra/: confluent-local + mongo:7 RS (D-05/07/08)
.mockery.yaml                # SVC-06: throwaway-интерфейс → smoke
Makefile (корневой)          # + dev-up/topics/test-integration/generate-mocks + MOCKERY_VERSION
lefthook.yml                 # D-02: снять исключение inventory из pre-push; добавить test-inventory (unit only)
go.work                      # D-01: inventory остаётся (уже есть — verified)
```

> Точное расположение пакетов (`internal/kafka/topology` vs `pkg`-level) — на усмотрение
> планнера в рамках D-06/D-14. Помни: `internal/` сейчас пуст (stale-леса снесены) — это чистый
> лист, не «починка лесов» (boundaries.md).

### Pattern 1: Single-source топология топиков (D-06)
**What:** Один Go-модуль владеет ИМЕНАМИ топиков, cleanup-политиками, числом партиций; экспортит
`Bootstrap(ctx, admCl, cfg)`. И CLI, и тесты зовут ЕЁ — не дублируют конфиг.
**When to use:** всегда для топологии; добавление агрегата (Phase 6+) = одна строка в списке
агрегатов (D-10 data-driven).
**Example:** см. § Code Examples / kadm bootstrap.

### Pattern 2: Build-tag изоляция интеграционных тестов (D-15)
**What:** Файлы интеграционных тестов начинаются с `//go:build integration`. `go test ./...`
(и pre-push) их НЕ компилируют → unit-прогон быстрый, без Docker. `make test-integration` =
`go test -tags=integration ./...`.
**When to use:** любой тест, поднимающий testcontainers/реальную инфру. Критично после D-02
(pre-push теперь включает inventory — нельзя, чтобы он тянул Docker).

### Pattern 3: RS-aware connection-helper (D-14)
**What:** Тонкая фабрика `Connect(ctx, uri) (*mongo.Client, error)` с `writeconcern.Majority()` +
health-`Ping`. Эталон «как подключаемся» для Phase 6/7. БЕЗ репозиториев/UoW/транзакций.
**When to use:** Phase 5 — доказать v2 рабочим коннектом к RS; дальше переиспользуется.

### Anti-Patterns to Avoid
- **Дубль конфига топиков в CLI и тестах** — нарушает D-06; дрейф имён/политик. Вместо —
  общий модуль.
- **`GOWORK=off` в новых рецептах** — отменён D-01; всё через workspace-build. Старый STACK.md
  §Installation с `GOWORK=off` — **устарел** под D-01, не копировать.
- **Реализовать UoW/repository в helper'е** — это Phase 7, не лезть (D-14).
- **Tombstone/null-ключ при provision** — в Phase 5 топики только создаются (пустые); ключи/
  tombstone — relay Phase 8 (Pitfall 2/3 не активны здесь, но топология должна их допускать).
- **Compose-stand как зависимость теста** — D-05 разводит пути; тест поднимает свои контейнеры.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Поднять Kafka/Mongo для теста | Свой docker-exec/healthcheck-поллинг в Go | `testcontainers-go` modules kafka/mongodb | Управление жизненным циклом, ожидание готовности, random-порты, Terminate — из коробки |
| Создать топик с политиками | Ручной Kafka-admin-протокол / shell `kafka-topics.sh` | `kadm.CreateTopics(ctx, parts, rf, configs, topics...)` | Типобезопасный config-map, единый клиент с kgo |
| Single-node RS healthcheck | Свой sleep+retry скрипт инициализации | compose healthcheck `try{rs.status()}catch{rs.initiate()}` | Идемпотентно, self-contained в compose, ретраит до готовности |
| Mongo client с RS-discovery | Ручной парсинг topology/seedlist | `options.Client().ApplyURI(uri)` + `directConnection` для локального RS | Драйвер делает discovery; для single-node localhost нужен `directConnection=true` (см. Pitfall) |
| Моки портов | Ручные fake-структуры | mockery codegen (`.mockery.yaml`) | Ручной фейк дрейфует от интерфейса; canon `knowledge/testing.md` |

**Key insight:** Phase 5 — это «обвязка», где почти всё уже решено зрелыми библиотеками; ценность
— в правильной **склейке** (single-source топология, build-tag изоляция, RS-aware URI), а не в
своём коде инфраструктуры.

## Common Pitfalls

### Pitfall 1: Single-node RS host resolution / `directConnection`
**What goes wrong:** Клиент к локальному single-node RS виснет/падает
`Could not find host matching read preference { mode: nearest }`, потому что Mongo advertise'ит
в RS-конфиге хост (`localhost`/`127.0.0.1`/имя контейнера), который не резолвится одинаково
изнутри контейнера и с хоста.
**Why it happens:** Драйвер по умолчанию делает RS topology discovery и идёт на advertise'нутый
хост, а не на адрес из URI.
**How to avoid:**
- **compose-стенд (ручной):** в `rs.initiate` member-host = `127.0.0.1:27017` (или
  `host.docker.internal` + `extra_hosts: host-gateway`), а клиент использует URI
  `mongodb://localhost:27017/?replicaSet=rs0&directConnection=true`. `directConnection=true`
  обходит discovery — идёт прямо на адрес URI.
- **testcontainers:** модуль `mongodb.WithReplicaSet("rs0")` сам конфигурит RS и
  `ConnectionString(ctx)` возвращает корректный URI — **доверять ему**, не хардкодить хост.
**Warning signs:** тест/коннект виснет на `Ping`; ошибка про read preference; работает изнутри
контейнера, но не с хоста. [VERIFIED: WebSearch — несколько независимых guides] [CITED: golang.testcontainers.org/modules/mongodb]

### Pitfall 2: `go build ./...` падает на `package main` в `cmd/`
**What goes wrong:** `cd services/audit && go build ./...` → `build output "cmd" already exists
and is a directory` (exit 1). `go build` без `-o` для `package main` пытается записать бинарь с
именем пакета-директории (`cmd`), а директория `cmd/` уже существует. **`go build ./cmd` тоже
падает** тем же сообщением (вопреки STATE.md).
**Why it happens:** Имя выходного бинаря по умолчанию = имя директории `main`-пакета; коллизия с
самой директорией.
**How to avoid (DOC-02):** Использовать `go vet ./...` (exit 0, без артефакта — каноническая
проверка) ИЛИ `go build -o /dev/null ./...`. **НЕ** `go build ./cmd`.
**Warning signs:** stray-бинарь `cmd` в рабочем дереве; CI/hook падает на build-шаге audit.
[VERIFIED: эмпирически воспроизведено в этой сессии — go 1.24.6 darwin/arm64]

### Pitfall 3: WIP-инвариант inventory ломает весь workspace-build (D-03)
**What goes wrong:** После D-01/D-02 любая ошибка компиляции в inventory (свап на v2,
import-конфликты) валит общий `go build ./...` / pre-push для ВСЕХ модулей.
**Why it happens:** Осознанная цена входа inventory в go.work — теперь он «эталонный», должен
компилироваться на каждом коммите (D-03).
**How to avoid:** Свап mongo-driver делать атомарно (go.mod + tidy + build зелёный в одном шаге);
интеграционные тесты — за build-tag `integration` (D-15), чтобы pre-push не тянул Docker/контейнеры.
**Warning signs:** pre-push красный на pkg/audit/analytics из-за inventory; «зелёный локально,
красный в хуке».

### Pitfall 4: Старый STACK.md §Installation с `GOWORK=off` устарел под D-01
**What goes wrong:** Скопировать `GOWORK=off go get ...` из `.planning/research/STACK.md` (он
писался до D-01, когда inventory был вне workspace).
**Why it happens:** STACK.md §Installation отражает прежнее (отменённое) решение.
**How to avoid:** В Phase 5 inventory — член go.work (D-01); `go get`/`go build`/`go vet` — без
`GOWORK=off`. Каноны build.md/structure.md переписать (D-04).
**Warning signs:** `GOWORK=off` в новых рецептах/Makefile-таргетах.

### Pitfall 5: mockery v2 vs v3 несовместимость конфига
**What goes wrong:** Взять синтаксис `.mockery.yaml` из v2-гайда — в v3 переименованы
`.PackageName`→`.SrcPackageName`, `.PackagePath`→`.SrcPackagePath`; по умолчанию моки кладутся
рядом с кодом (не в `mocks/`); один файл на пакет (не на мок).
**Why it happens:** Большинство гайдов в сети — v2.
**How to avoid:** Использовать v3-config (см. § Code Examples / mockery); `knowledge/testing.md`
уже описывает v3-expecter-API (`mock.EXPECT()`, `NewMockX(GinkgoT())`) → согласовано с v3.
**Warning signs:** mockery ругается на неизвестные ключи; шаблон с `.PackagePath` не подставляется.

## Code Examples

> Все сниппеты — целевой вид на основе verbatim-сигнатур из официальных источников
> (pkg.go.dev / исходники тегов). Доменный плейсхолдер избегаем; имена — иллюстративные.

### kadm bootstrap топиков (SVC-07, D-06/D-10/D-11/D-12)
```go
// Source: pkg.go.dev/github.com/twmb/franz-go/pkg/kadm — CreateTopics signature [VERIFIED]
//   func NewClient(cl *kgo.Client) *Client
//   func (cl *Client) CreateTopics(ctx, partitions int32, replicationFactor int16,
//                                  configs map[string]*string, topics ...string) (CreateTopicResponses, error)
//   func StringPtr(s string) *string

// топология — единственный источник истины (D-06); агрегаты data-driven (D-10)
var aggregates = []string{"host"} // Phase 6+: append "project","module","location" — одна строка

func Bootstrap(ctx context.Context, adm *kadm.Client, partitions int32) error {
    rf := int16(1) // single broker (D-07)
    for _, agg := range aggregates {
        // *.events — cleanup=delete, immutable-история фактов (D-12)
        eventsCfg := map[string]*string{"cleanup.policy": kadm.StringPtr("delete")}
        if _, err := adm.CreateTopics(ctx, partitions, rf, eventsCfg,
            "inventory."+agg+".events"); err != nil {
            return fmt.Errorf("создать топик %s.events: %w", agg, err)
        }
        // *.state — cleanup=compact + delete.retention.ms >= 24h (D-12)
        stateCfg := map[string]*string{
            "cleanup.policy":      kadm.StringPtr("compact"),
            "delete.retention.ms": kadm.StringPtr("86400000"), // 24h
        }
        if _, err := adm.CreateTopics(ctx, partitions, rf, stateCfg,
            "inventory."+agg+".state"); err != nil {
            return fmt.Errorf("создать топик %s.state: %w", agg, err)
        }
    }
    return nil
}
// adm := kadm.NewClient(kgoClient)  // kgoClient — kgo.NewClient(kgo.SeedBrokers(brokers...))
```

### Mongo connection-helper v2 (SVC-05, D-14)
```go
// Source: pkg.go.dev/go.mongodb.org/mongo-driver/v2/mongo — v2 API [VERIFIED]
//   ВАЖНО v2: mongo.Connect(opts ...*options.ClientOptions) (*Client, error) — БЕЗ ctx (отличие от v1)
//   writeconcern.Majority() *WriteConcern ; Client.Ping(ctx, *readpref.ReadPref) error
import (
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    "go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

// Connect — RS-aware фабрика клиента + health-ping (D-14). Без UoW/repo (Phase 7).
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
    cl, err := mongo.Connect(
        options.Client().
            ApplyURI(uri). // локальный single-node RS: ...?replicaSet=rs0&directConnection=true
            SetWriteConcern(writeconcern.Majority()),
    )
    if err != nil {
        return nil, fmt.Errorf("подключение к Mongo: %w", err)
    }
    if err := cl.Ping(ctx, nil); err != nil { // health-ping (readpref nil = primary)
        _ = cl.Disconnect(ctx)
        return nil, fmt.Errorf("ping Mongo: %w", err)
    }
    return cl, nil
}
```

### testcontainers smoke-тест (SVC-06, D-15) — Ginkgo + build-tag
```go
//go:build integration
// Source: golang.testcontainers.org/modules/{kafka,mongodb} v0.43.0 + исходники тегов [VERIFIED]
//   kafka.Run(ctx, img, opts...) (*KafkaContainer, error); (kc).Brokers(ctx) ([]string, error)
//   kafka.WithClusterID(string); mongodb.Run(ctx, img, opts...); mongodb.WithReplicaSet(replSetName string)
//   (c).ConnectionString(ctx) (string, error); testcontainers.TerminateContainer(c)
package inventory_test

import (
    "context"
    "testing"

    . "github.com/onsi/ginkgo/v2" // dot-import (knowledge/testing.md)
    . "github.com/onsi/gomega"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/kafka"
    "github.com/testcontainers/testcontainers-go/modules/mongodb"
    "github.com/twmb/franz-go/pkg/kgo"
)

func TestInventoryIntegration(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Inventory Integration Suite")
}

var _ = Describe("dev infra smoke", func() {
    var ctx context.Context
    BeforeEach(func() { ctx = context.Background() })

    It("connects to ephemeral Kafka (KRaft) and provisions topics", func() {
        kc, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0",
            kafka.WithClusterID("test-cluster"))
        DeferCleanup(func() { _ = testcontainers.TerminateContainer(kc) })
        Expect(err).ToNot(HaveOccurred())

        brokers, err := kc.Brokers(ctx)
        Expect(err).ToNot(HaveOccurred())

        cl, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
        Expect(err).ToNot(HaveOccurred())
        DeferCleanup(cl.Close)
        Expect(cl.Ping(ctx)).To(Succeed()) // smoke connect

        // adm := kadm.NewClient(cl); Expect(Bootstrap(ctx, adm, 6)).To(Succeed())
    })

    It("connects to ephemeral Mongo single-node replica set", func() {
        mc, err := mongodb.Run(ctx, "mongo:7", mongodb.WithReplicaSet("rs0"))
        DeferCleanup(func() { _ = testcontainers.TerminateContainer(mc) })
        Expect(err).ToNot(HaveOccurred())

        uri, err := mc.ConnectionString(ctx)
        Expect(err).ToNot(HaveOccurred())

        client, err := Connect(ctx, uri) // helper из примера выше
        Expect(err).ToNot(HaveOccurred())
        DeferCleanup(func() { _ = client.Disconnect(ctx) })
    })
})
```

### docker-compose: confluent-local + mongo:7 single-node RS (SVC-07, D-05/07/08)
```yaml
# Source: WebSearch (несколько independent guides) + confluent-local docs [CITED/ASSUMED]
# Ручной дев-стенд (D-05). Точные env confluent-local — executor верифицирует при `docker compose up`.
services:
  kafka:
    image: confluentinc/confluent-local:7.5.0   # KRaft из коробки, 1 broker (D-07)
    ports: ["9092:9092"]
    # confluent-local auto-конфигурит KRaft; доп. KAFKA_* при необходимости
  mongo:
    image: mongo:7                               # (D-08)
    command: ["--replSet", "rs0", "--bind_ip_all", "--port", "27017"]
    ports: ["27017:27017"]
    healthcheck:                                  # idempotent rs.initiate (Pitfall 1)
      test: >
        echo "try { rs.status() } catch (e) {
          rs.initiate({_id:'rs0',members:[{_id:0,host:'127.0.0.1:27017'}]}) }" |
        mongosh --port 27017 --quiet
      interval: 5s
      timeout: 30s
      retries: 30
# клиент к стенду: mongodb://localhost:27017/?replicaSet=rs0&directConnection=true
```

### mockery v3 `.mockery.yaml` (SVC-06)
```yaml
# Source: vektra.github.io/mockery v3 docs + pkg.go.dev/github.com/vektra/mockery/v3/config [CITED]
# v3: моки рядом с кодом по умолчанию; .SrcPackageName/.SrcPackagePath (НЕ .PackageName); template: testify
all: false
formatter: goimports
template: testify
dir: "{{.InterfaceDir}}/mocks"
filename: "{{.InterfaceName}}.go"
pkgname: "mocks"
packages:
  github.com/gwall-e/services/inventory/internal/<example-pkg>:  # throwaway-интерфейс (SVC-06)
    interfaces:
      ExampleProvisioner: {}   # пример-порт; реальные доменные порты — Phase 6/7
# make generate-mocks → `mockery` (бинарь пиннут в Makefile: go install .../mockery/v3@v3.7.1)
```

## Runtime State Inventory

> Phase 5 — частично rename/refactor (D-01/D-02/D-04: отмена `GOWORK=off`, снятие исключения
> inventory). State-инвентаризация строки `GOWORK=off` и связанных конфигов:

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — БД/датастораны в Phase 5 не несут строк, подлежащих переименованию (топики только создаются пустыми). | none |
| Live service config | None — нет внешних UI-сервисов (n8n/Datadog/Tailscale) в проекте; локальный стенд эфемерный. | none |
| OS-registered state | git-хуки lefthook в `.git/hooks/` (локальны, не коммитятся). **`pre-push`** содержит исключение inventory (комменты+отсутствие команды) — править `lefthook.yml` в репо, затем `lefthook install` переустановит хуки (одноразово, локально). | code edit (`lefthook.yml`) + ручной `lefthook install` |
| Secrets/env vars | None — секретов/`.env` нет; `KAFKA_BROKERS`-подобные env — новые (для CLI), не переименование. | none |
| Build artifacts | `services/inventory/go.sum` (54 строки) под v1-mongo-driver — станет stale после свапа на v2; `go mod tidy` пересоберёт. v1-indirect-deps (`xdg-go/*`, `montanaflynn/stats`, `golang/snappy`) уйдут. | `go mod tidy` после свапа |

**Канон-файлы с `GOWORK=off` (D-04 — править как код):**
- `knowledge/build.md` — §«inventory — WIP, `GOWORK=off`» (строки ~68-81), §pre-push исключение
  (~44-47), §audit-рецепт DOC-02 (~62-63). [VERIFIED: чтение build.md]
- `knowledge/structure.md` — §«inventory — вне workspace, WIP» (строки ~36-48). [VERIFIED]
- `knowledge/boundaries.md` — карта владения «исключение inventory из pre-push» (строка ~72);
  свериться/обновить. [VERIFIED]
- `lefthook.yml` — комменты D-01/D-03 про GOWORK=off и `lint-inventory` с `GOWORK=off` (строки
  ~33-47), pre-push без inventory (~49-57). [VERIFIED]
- `.planning/ROADMAP.md §Phase 5` SC1 — убрать «с `GOWORK=off`» (D-04). [не прочитан — указан в CONTEXT]

**Nothing found in category:** Stored data, Live service config, Secrets/env vars — None
(verified чтением структуры проекта: нет БД-state, нет внешних сервисов, нет `.env`).

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `mongo.Connect(ctx, opts)` (v1) | `mongo.Connect(opts ...)` — **без ctx** (v2) | mongo-driver v2 (2024+) | Сигнатура helper'а без `ctx` в Connect; ctx — в операциях (Ping/Disconnect) [VERIFIED] |
| `primitive.ObjectID`, `bsonrw`, `gridfs` | `bson` пакет реорганизован; эти пакеты удалены/перемещены в v2 | v2 | В inventory кода на v1 нет → import-правок нет (D-13) [CITED] |
| `inventory` вне go.work, `GOWORK=off` | inventory — член go.work, общий build | Phase 5 (D-01) | Все рецепты без `GOWORK=off`; STACK.md §Installation устарел |
| mockery v2 (`.PackageName`, `mocks/` dir) | mockery v3 (`.SrcPackageName`, моки рядом с кодом, 1 файл/пакет) | mockery v3.0 (2025) | v3-config обязателен; v2-гайды несовместимы [CITED] |
| testcontainers `RunContainer(...)` | `Run(ctx, img, opts...)` | testcontainers-go v0.3x | `RunContainer` deprecated; использовать `Run` [VERIFIED] |
| `mongodb.WithReplicaSet()` (без арга, старые версии) | `mongodb.WithReplicaSet(replSetName string)` (v0.43.0) | testcontainers v0.43.0 | Передавать имя RS: `WithReplicaSet("rs0")` [VERIFIED: исходник тега v0.43.0] |

**Deprecated/outdated:**
- `GOWORK=off`-рецепты для inventory — отменены D-01.
- `go build ./cmd` как DOC-02 fix — **не работает** (падает тем же сообщением).
- STACK.md §Installation (`GOWORK=off go get ...`) — устарел под D-01.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `confluentinc/confluent-local:7.5.0` стартует KRaft без доп. `KAFKA_*` env в compose | Code Examples / compose | LOW — confluent-local дизайнится auto-config KRaft; executor верифицирует `docker compose up`. testcontainers-модуль использует тот же образ — паритет (D-07) |
| A2 | Точные env-переменные confluent-local для compose (advertised listeners на :9092) | Code Examples / compose | MEDIUM — может потребовать `KAFKA_ADVERTISED_LISTENERS`; SC2 требует реального подъёма → executor верифицирует эмпирически |
| A3 | `mongo:7` совместим с `mongodb.WithReplicaSet` модуля v0.43.0 (модуль-дефолт `mongo:6`) | Standard Stack / Alternatives | LOW — v2.7 требует ≥4.4; mongo:7 новее mongo:6; модуль принимает любой образ через `Run(ctx, img, ...)` |
| A4 | mockery v3 (не v2) — правильный выбор | Standard Stack | LOW — `knowledge/testing.md` описывает v3-expecter-API; но v2.53.6 тоже поддерживается → см. Open Question Q1 |
| A5 | compose member-host `127.0.0.1` + клиентский `directConnection=true` достаточно для дев-стенда (vs `host.docker.internal`) | Pitfalls / compose | MEDIUM — зависит от того, бежит ли клиент на хосте; для app-на-хосте `directConnection` работает; для контейнер-к-контейнеру может понадобиться `host.docker.internal` |
| A6 | `delete.retention.ms=86400000` (24h) — приемлемый дев-дефолт для `*.state` | Code Examples / kadm | LOW — D-12 залочил «≥24h»; точное prod-значение отложено (Deferred) |

## Open Questions

1. **mockery v2 vs v3?**
   - Что знаем: v3.7.1 — текущая major; v2.53.6 — последняя v2. `knowledge/testing.md` описывает
     v3-стиль (expecter API, yaml-config). Оба генерят testify-style моки с `mock.EXPECT()`.
   - Что неясно: команда могла иметь предпочтение к v2 (стабильность) — но канон уже под v3.
   - Рекомендация: **v3.7.1** (согласовано с testing.md); пиннить в Makefile. Smoke на
     throwaway-интерфейсе докажет работоспособность.

2. **Где живёт docker-compose.yml и Go-модуль топологии — корень репо или infra/ / internal/?**
   - Что знаем: корневой `docker-compose.yml`/`Makefile` помечены «стале/неавторитетны»
     (boundaries.md) — но это про СУЩЕСТВУЮЩИЕ стале-файлы; новый compose в Phase 5 авторитетен.
   - Что неясно: положить новый compose в корень (перезаписав стале) или в `infra/`/`deploy/`.
   - Рекомендация: планнер решает; если перезаписываем корневой стале-compose — убрать его из
     «неавторитетных» в boundaries.md (или явно отметить как актуальный). Go-модуль топологии —
     `services/inventory/internal/kafka/topology` (D-06 говорит «один Go-модуль», в рамках inventory).

3. **`docker compose` доступность в среде разработчика/CI.**
   - Что знаем: в этой sandbox-сессии `docker --version`=28.4.0 ОК, но `docker compose`
     subcommand недоступен (`unknown command`) и `docker-compose` не найден — это ограничение
     sandbox, не машины разработчика.
   - Что неясно: гарантирован ли compose-плагин на dev-машинах.
   - Рекомендация: SC2 («`docker compose up` поднимает…») верифицируется разработчиком на реальной
     машине; в Makefile-таргете `dev-up` использовать `docker compose` (v2 plugin). См. § Environment Availability.

## Environment Availability

| Dependency | Required By | Available (this session) | Version | Fallback |
|------------|------------|--------------------------|---------|----------|
| Go | весь build | ✓ | 1.24.6 darwin/arm64 | — |
| Docker engine | testcontainers, compose | ✓ | 28.4.0 | — |
| `docker compose` (v2 plugin) | SC2 dev-стенд (`make dev-up`) | ✗ (sandbox) | — | верифицировать на dev-машине; ограничение sandbox |
| testcontainers (нужен живой Docker daemon) | интеграционные тесты | ✓ (daemon есть) | — | — |
| Go module proxy | `go get` стека | ✓ (версии подтверждены) | — | — |

**Missing dependencies with no fallback:** None для самого кода.
**Missing dependencies with fallback:** `docker compose` subcommand недоступен в текущей sandbox —
это **не** блокер для разработчика (Docker engine присутствует); SC2 верифицируется на реальной
машине. testcontainers использует Docker daemon напрямую (не compose) → интеграционные тесты не
зависят от наличия compose-плагина.

## DOC-02 — эмпирическое решение

Воспроизведено в этой сессии (go 1.24.6 darwin/arm64, `cd services/audit`):

| Команда | Результат | Артефакт |
|---------|-----------|----------|
| `go build ./...` | **exit 1** — `build output "cmd" already exists and is a directory` | — |
| `go build ./cmd` | **exit 1** — то же сообщение (вопреки STATE.md!) | — |
| `go vet ./...` | **exit 0** | нет (чистая проверка) ✅ |
| `go build -o /dev/null ./...` | **exit 0** | нет ✅ |

**Рекомендация DOC-02:** заменить рецепт на **`go vet ./...`** (каноническая валидация без
бинаря; уже идиома проекта). С D-01 audit — член workspace, рецепт приводится к workspace-форме:
из корня `go vet ./...` покрывает все модули, либо per-module `cd services/audit && go vet ./...`.
Финальную форму executor верифицирует (SC5 требует реального exit 0). [VERIFIED: эмпирически]

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Ginkgo v2.23.4 + Gomega v1.38.0 (pinned в `pkg/go.mod`) [VERIFIED: grep] |
| Config file | none — `RegisterFailHandler(Fail)` + `RunSpecs` в `TestXxx` (knowledge/testing.md) |
| Quick run (unit) | `go test ./...` (без build-tag — integration-тесты не компилируются, D-15) |
| Integration run | `go test -tags=integration ./...` → `make test-integration` (D-15) |
| Mocks | `mockery` (v3.7.1) → `make generate-mocks` |

### Phase Requirements / Success Criteria → Validation Map
| SC / Req | Behavior | Test Type | Validation Command / Observable | Exists? |
|----------|----------|-----------|----------------------------------|---------|
| SC1 / SVC-05 | inventory собирается с mongo-driver/v2, v1 удалён | build (unit-gate) | `go build ./...` + `go vet ./...` из корня workspace зелёные; `grep mongo-driver` в go.mod показывает только `/v2`; нет `v1.17.9` | ❌ Wave 0 |
| SC2 / SVC-07 | `docker compose up` поднимает Kafka KRaft + Mongo RS (транзакции доступны) | manual compose smoke | `make dev-up`; `mongosh ...?directConnection=true` → `rs.status().ok==1`; Kafka доступен на :9092 | ❌ Wave 0 (manual) |
| SC3 / SVC-07 | Bootstrap провижнит `*.events`(delete) + `*.state`(compact) | integration | `Bootstrap(ctx, adm, 6)` → assert `kadm.ListTopics`/`DescribeTopicConfigs` показывает cleanup.policy=delete/compact, partitions=6 | ❌ Wave 0 |
| SC4 / SVC-06 | testcontainers (Kafka KRaft + Mongo RS) стартует и подключается; Ginkgo/Gomega/mockery smoke | integration + unit | `go test -tags=integration ./...` зелёный (kgo.Ping + mongo.Ping проходят); `make generate-mocks` генерит мок; unit-suite с моком зелёный | ❌ Wave 0 |
| SC5 / DOC-02 | build.md audit-рецепт проходит (exit 0) | manual/CI | новый рецепт (`go vet ./...`) → `echo $?` == 0; build.md обновлён | ✅ команда verified, ❌ build.md правка |
| D-01 | inventory в go.work, workspace-build | build-gate | `go.work` содержит `./services/inventory` (verified); общий `go build ./...` зелёный | ✅ go.work, ❌ build green |
| D-02 | pre-push включает inventory (unit only) | hook | `lefthook.yml` pre-push гоняет inventory unit-тесты; integration за тегом не запускается | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go vet ./...` + `go build ./...` из корня (быстро, без Docker).
- **Per wave merge:** `go test ./...` (unit) + `make test-integration` (если Docker доступен).
- **Phase gate:** все 5 SC TRUE; `make dev-up` ручной smoke (SC2); full unit+integration зелёные
  до `/gsd-verify-work`.

### Wave 0 Gaps
- [ ] `services/inventory/go.mod` — свап на mongo-driver/v2, `go mod tidy` (SC1)
- [ ] `internal/kafka/topology/` — const + `Bootstrap` на kadm (SC3) + unit-тест констант
- [ ] connection-helper (Connect+Ping) в repositories/infra (SC4)
- [ ] `<pkg>_integration_test.go` с `//go:build integration` — testcontainers smoke (SC3/SC4)
- [ ] `.mockery.yaml` + throwaway-интерфейс + сгенерированный мок + unit-spec с моком (SC4)
- [ ] `docker-compose.yml` (confluent-local + mongo:7 RS healthcheck) (SC2)
- [ ] `cmd/` CLI провижна → make-таргет (SC3 manual path)
- [ ] Makefile-таргеты: `dev-up`, `topics`, `test-integration`, `generate-mocks` + `MOCKERY_VERSION` pin
- [ ] `lefthook.yml` — снять исключение inventory из pre-push, добавить inventory unit-тест (D-02)
- [ ] Каноны build.md/structure.md/boundaries.md + ROADMAP SC1 — переписать под D-01/D-02/DOC-02 (D-04)

## Security Domain

> `security_enforcement: true`, ASVS level 1. Phase 5 — dev-инфра без обработки пользовательского
> ввода/аутентификации/прав. Применимость ASVS-категорий минимальна.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | Нет аутентификации в Phase 5 (identity/interceptor — SVC-08, Phase 7) |
| V3 Session Management | no | Нет сессий |
| V4 Access Control | no | Нет прав (Access — отдельный домен, out of scope) |
| V5 Input Validation | no | Нет пользовательского ввода (CLI-флаги — dev-only, доверенный оператор) |
| V6 Cryptography | no | Дев-стенд plaintext (Kafka PLAINTEXT, Mongo без auth); НЕ для prod |
| V14 Configuration | partial | Версии стека пиннуты (Makefile/go.mod); образы — конкретные теги (confluent-local:7.5.0, mongo:7), не `:latest` |

### Known Threat Patterns for Go dev-infra

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Slopsquatting Go-модуля | Tampering | Версии верифицированы по go-proxy (§ Package Legitimacy); `go.sum` фиксирует хеши |
| Plaintext Kafka/Mongo в дев-стенде утекает в prod | Information Disclosure | Дев-стенд явно помечен «не для prod»; auth/TLS — отдельная prod-задача (вне Phase 5) |
| `:latest`-образы → недетерминизм/supply-chain | Tampering | Конкретные теги образов (D-07/D-08); пиннинг тулинга в Makefile |

**Вывод:** security-поверхность Phase 5 — supply-chain (пиннинг версий/образов) и явная пометка
дев-стенда как небезопасного-для-prod. Аутентификация/права/крипто — out of scope (другие фазы/домены).

## Sources

### Primary (HIGH confidence)
- Go module proxy (`proxy.golang.org/<mod>/@latest`) — версии franz-go v1.21.4, pkg/kadm v1.18.0,
  mongo-driver/v2 v2.7.0, testcontainers-go v0.43.0 (+modules kafka/mongodb), mockery v3.7.1/v2.53.6
  — verified этой сессией
- pkg.go.dev `github.com/twmb/franz-go/pkg/kadm` — `CreateTopics(ctx, partitions int32, rf int16,
  configs map[string]*string, topics...)`, `NewClient`, `StringPtr`
- pkg.go.dev `github.com/twmb/franz-go/pkg/kgo` — `NewClient`, `SeedBrokers`, `Ping(ctx)`, `Close()`,
  idempotent по умолчанию
- pkg.go.dev `go.mongodb.org/mongo-driver/v2/mongo` — `Connect(opts...)` (без ctx), `options.Client()`,
  `writeconcern.Majority()`, `Ping(ctx, *readpref)`, `Disconnect(ctx)`
- golang.testcontainers.org/modules/kafka + исходник тега `modules/kafka/v0.43.0` —
  `Run`, `Brokers(ctx) ([]string,error)`, `WithClusterID`, default `confluentinc/confluent-local:7.5.0`
- golang.testcontainers.org/modules/mongodb + исходник тега `modules/mongodb/v0.43.0` —
  `Run`, `WithReplicaSet(replSetName string)`, `ConnectionString(ctx)`
- Эмпирическое воспроизведение DOC-02 (go 1.24.6) — `go vet ./...` exit 0; `go build ./cmd` падает
- Чтение репозитория: `go.work`, `go.mod` (inventory/root), `lefthook.yml`, `Makefile`, knowledge/*.md
- vektra.github.io/mockery (v3) + pkg.go.dev/github.com/vektra/mockery/v3/config — yaml-config,
  `.SrcPackageName`/`.SrcPackagePath`, testify-template

### Secondary (MEDIUM confidence)
- WebSearch (Workleap/Anthony Simmon, DEV, GinkCode guides) — docker-compose single-node Mongo RS
  healthcheck `try{rs.status()}catch{rs.initiate()}`, host resolution, `directConnection=true`
- WebSearch (mockery v3 migration, Jaeger/cloudbeat `.mockery.yaml`) — v3-config структура

### Tertiary (LOW confidence / ASSUMED)
- Точные `KAFKA_*` env confluent-local для compose advertised listeners (A1/A2) — executor верифицирует

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — все версии + сигнатуры verified по go-proxy/pkg.go.dev/исходникам тегов
- Architecture (single-source топология, build-tag изоляция, RS-helper): HIGH — следует из
  залоченных D-06/D-14/D-15 + verified API
- DOC-02: HIGH — эмпирически воспроизведено в этой сессии
- docker-compose Mongo RS healthcheck: MEDIUM — паттерн из нескольких independent guides; точные
  confluent-local env — ASSUMED (executor верифицирует, SC2 требует реального подъёма)
- mockery v2/v3: MEDIUM — выбор v3 согласован с testing.md, но требует подтверждения команды

**Research date:** 2026-06-27
**Valid until:** 2026-07-27 (стабильные библиотеки; стек-версии могут получить патчи — перепроверить
go-proxy перед `go get`, если планирование задержится)
