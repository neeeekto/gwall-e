# Architecture Research

**Domain:** Conventions knowledge base for AI agents/team (PRIMARY) + canonical Go DDD/hexagonal service it documents (SECONDARY) + Go microservices HaaS/DC-management platform (BACKGROUND)
**Researched:** 2026-06-17
**Confidence:** HIGH (knowledge-base layout + canonical service — grounded in locked project decisions); MEDIUM (Layer-3 platform topology, future epics)

> **Два слоя, один документ.** Слой 1 (PRIMARY) — *архитектура самой базы знаний*: какие rule-доки существуют, как они связаны и в каком порядке их писать. Слой 2 (SECONDARY) — *эталонная архитектура сервиса, которую эти доки описывают*: DDD + гексагон (порты и адаптеры), **без CQRS-шины**. Слой 3 (BACKGROUND) — топология много-сервисной платформы для будущих эпиков. Первый milestone строит Слой 1; Слой 2 — это *контент*, который он документирует; Слой 3 — вне скоупа, но формирует словарь глоссария/архитектуры.
>
> ⚠️ **Важно:** старый код в git HEAD (CQRS-диспетчер `pkg/mediatr`, `CommandDispatcher`/`QueryDispatcher`, `TxManager`/`tx.go`) **снесён и невалиден**. Этот документ описывает целевую архитектуру, зафиксированную в PROJECT.md → Key Decisions, а не мёртвый код.

## Standard Architecture

### System Overview — Layer 2 (эталонный сервис)

Гексагон (порты и адаптеры) + DDD, БЕЗ CQRS-шины/диспетчера. Одно правило зависимостей: **всё указывает внутрь, на `domain`; `domain` не импортирует ничего нашего.** Inbound-адаптеры (gRPC, cron) зовут use case'ы напрямую.

```
┌──────────────────────────────────────────────────────────────────┐
│                     INBOUND ADAPTERS                               │
│  ┌──────────────────┐                ┌──────────────────┐         │
│  │  api/  (gRPC)     │                │  cron/  (jobs)   │         │
│  │  proto-хендлеры   │                │  периодич. задачи│         │
│  └────────┬─────────┘                └────────┬─────────┘         │
│           │  зовут Execute() напрямую          │                   │
├───────────┴────────────────────────────────────┴──────────────────┤
│                     APPLICATION                                    │
│  ┌────────────────────────┐      ┌────────────────────────┐       │
│  │  usecases/ (write)      │      │  query/ (read)          │       │
│  │  1 use case = 1 struct  │      │  query-сервисы → DTO    │       │
│  │  Execute(ctx,in)(out,e) │      │  читают Mongo напрямую  │       │
│  └───────────┬────────────┘      └───────────┬────────────┘       │
│              │ uow.Do(...) + repo              │ прямой read       │
├──────────────┼──────────────────────────────────┼─────────────────┤
│              ▼            DOMAIN                  │                 │
│  ┌──────────────────────────────────────────────┐│                 │
│  │ агрегаты (private поля, фабрики NewX, инвар.) ││                 │
│  │ value objects, типизированные ID             ││                 │
│  │ доменные события (RecordEvent / PullEvents)   ││                 │
│  │ ПОРТЫ (интерфейсы): Repository, UnitOfWork,   ││                 │
│  │                     EventPublisher            ││                 │
│  └──────────────────────────────────────────────┘│                 │
├───────────────────────────────────────────────────┼────────────────┤
│                     OUTBOUND ADAPTERS              ▼                │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │ repositories/ : Mongo-реализации портов + UnitOfWork +        │ │
│  │                 transactional outbox (запись событий в tx)    │ │
│  │ + relay: читает outbox и публикует через EventPublisher       │ │
│  └──────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘

app/  = composition root (ручной DI, сборка зависимостей)
cmd/  = main, запуск процесса
```

### Component Responsibilities — Layer 2

| Компонент | За что отвечает | Реализация |
|-----------|-----------------|------------|
| `domain/` | Агрегаты, VO, типизированные ID, доменные события, **порты** (`Repository`, `UnitOfWork`, `EventPublisher`). Инварианты в фабриках `NewX(...) (*X, error)`. | Чистый Go, без внешних зависимостей; встраивает `AggregateRoot` для накопления событий |
| `usecases/` | Write-side. **1 use case = 1 struct** + `Execute(ctx, in) (out, error)`, по файлу на use case. Маппинг DTO→домен, оркестрация, обёртка записи в `uow.Do`. | Конструктор `NewXxxUseCase(repo, uow, ...)`; без диспетчера |
| `query/` | Read-side. Query-сервисы читают Mongo **напрямую в DTO**, минуя агрегаты/репозитории write-side. | `type ListHostsQuery struct{...}` + `Execute(ctx, in) ([]DTO, error)` |
| `repositories/` | Outbound: Mongo-реализации портов `Repository`; реализация `UnitOfWork` (сессия/транзакция в ctx); запись событий в outbox внутри транзакции. | mongo-driver v2; транзакция «едет» в `ctx` |
| `api/` | Inbound: gRPC-адаптеры (proto). Декод запроса → вызов `usecase.Execute` / `query.Execute` → энкод ответа. | grpc-go + go-grpc-middleware/v2 + protovalidate |
| `cron/` | Inbound: периодические джобы; зовут конкретный use case напрямую. | Планировщик + ссылки на use case'ы |
| `app/` | Composition root: ручной DI, проводка всех слоёв, env-конфиг. | Конструкторы + явная сборка в одном месте |
| `cmd/` | `main`: запуск процесса, graceful shutdown. | Тонкий entrypoint |

### System Overview — Layer 1 (структура базы знаний)

База знаний — это граф документов с порядком авторинга. Тонкие точки входа (`CLAUDE.md`/`AGENTS.md`) ссылаются на per-topic файлы в `knowledge/` (progressive disclosure).

```
CLAUDE.md / AGENTS.md  (тонкий индекс-ссылки, MUST/SHOULD/WON'T)
        │
        ├──→ knowledge/README.md         # индекс/навигация
        ├──→ knowledge/glossary.md        # ubiquitous language (RU термины) ← prereq для DDD-доков
        ├──→ knowledge/structure.md       # go.work, модули, раскладка
        ├──→ knowledge/build.md           # сборка/тесты, GOWORK=off, cd pkg
        ├──→ knowledge/testing.md         # Ginkgo + Gomega
        ├──→ knowledge/style.md           # стиль, ошибки, язык (RU комментарии)
        ├──→ knowledge/architecture.md    # DDD+гексагон, слои/порты, UoW, outbox
        ├──→ knowledge/git.md             # git-конвенции
        └──→ knowledge/boundaries.md      # «do not»: не чинить леса, не верить stale README
```

## Recommended Project Structure

```
internal/
├── domain/            # агрегаты, VO, типизированные ID, доменные события, ПОРТЫ
│   ├── host.go        # пример агрегата
│   ├── events.go      # доменные события
│   ├── repository.go  # порт Repository (интерфейс)
│   ├── unit_of_work.go# порт UnitOfWork (интерфейс)
│   └── publisher.go   # порт EventPublisher (интерфейс)
├── usecases/          # write-side: 1 файл = 1 use case (struct + Execute)
│   ├── register_host.go
│   └── provision_vm.go
├── query/             # read-side: query-сервисы → DTO
│   ├── list_hosts.go
│   └── get_project.go
├── repositories/      # infra: Mongo-реализации портов + UnitOfWork + outbox
│   ├── host_repo.go
│   ├── unit_of_work.go
│   └── outbox.go
├── api/               # inbound: gRPC (proto) адаптеры
│   └── grpc/
└── cron/              # inbound: джобы
app/                   # composition root (ручной DI)
cmd/                   # main
```

### Structure Rationale

- **`domain/` владеет портами:** интерфейсы `Repository`/`UnitOfWork`/`EventPublisher` объявлены в домене (DDD-стандарт), реализации — в `repositories/`. Зависимость всегда внутрь.
- **`usecases/` (interactor):** один сценарий — одна структура; легко тестировать, легко навигировать ИИ, нет «магии» диспетчера.
- **`query/` отдельно от `usecases/`:** read-side не тащит доменную логику; чтение оптимизируется независимо (CQRS-lite — разделение reads/writes без шины).
- **`app/` отделён от `cmd/`:** вся проводка зависимостей собрана в одном месте, `main` остаётся тонким.

## Architectural Patterns

### Pattern 1: Use case как interactor (write-side) — Layer 2

**What:** Каждый write-сценарий — отдельная структура с зависимостями-полями и единственным методом `Execute`.
**When to use:** Все команды (изменение состояния).
**Trade-offs:** + Изоляция, тестируемость, дружелюбность к ИИ. − Больше файлов (приемлемо).

```go
type RegisterHostUseCase struct {
    hosts domain.HostRepository
    uow   domain.UnitOfWork
}

func NewRegisterHostUseCase(hosts domain.HostRepository, uow domain.UnitOfWork) *RegisterHostUseCase {
    return &RegisterHostUseCase{hosts: hosts, uow: uow}
}

func (uc *RegisterHostUseCase) Execute(ctx context.Context, in RegisterHostInput) (RegisterHostOutput, error) {
    // маппинг DTO → домен, проверка инвариантов в фабрике
    host, err := domain.NewHost(in.FQDN, in.RackID)
    if err != nil {
        return RegisterHostOutput{}, err
    }
    // вся запись — в одной транзакции
    err = uc.uow.Do(ctx, func(ctx context.Context) error {
        return uc.hosts.Save(ctx, host) // Save пишет агрегат + события в outbox в этой же tx
    })
    if err != nil {
        return RegisterHostOutput{}, fmt.Errorf("register host: %w", err)
    }
    return RegisterHostOutput{ID: host.ID()}, nil
}
```

### Pattern 2: Read-side query-сервисы (CQRS-lite) — Layer 2

**What:** Чтение идёт мимо агрегатов — query-сервис читает Mongo напрямую и собирает DTO.
**When to use:** Все запросы (list/get/search).
**Trade-offs:** + Быстро, без гидрации агрегатов; read-модель эволюционирует отдельно. − Дублирование формы данных (осознанное).

```go
type ListHostsQuery struct {
    db *mongo.Database
}

func (q *ListHostsQuery) Execute(ctx context.Context, in ListHostsInput) ([]HostDTO, error) {
    // прямой запрос в коллекцию → []HostDTO, без domain.Host
}
```

### Pattern 3: UnitOfWork как порт, транзакция в ctx — Layer 2

**What:** Порт `UnitOfWork` в `domain/`, Mongo-реализация в `repositories/` открывает сессию/транзакцию и кладёт её в `ctx`; репозитории достают транзакцию из `ctx`. **`TxManager` старого кода удалён — это его замена.**
**When to use:** Любая запись, затрагивающая >1 операцию/документ.
**Trade-offs:** + Единая, не-протекающая обработка транзакций; use case владеет границей. − Репозитории обязаны уважать транзакцию из ctx.

```go
// domain/unit_of_work.go
type UnitOfWork interface {
    Do(ctx context.Context, fn func(ctx context.Context) error) error
}
```

### Pattern 4: Доменные события через transactional outbox — Layer 2

**What:** Агрегат накапливает события (`RecordEvent`), отдаёт через `PullEvents()`. Репозиторий **внутри той же UnitOfWork-транзакции** пишет и агрегат, и сериализованные события в коллекцию `outbox`. Отдельный **relay** читает outbox и публикует через `EventPublisher` (at-least-once).
**When to use:** Везде, где есть доменные события и нужна согласованность (core value проекта).
**Trade-offs:** + Нет dual-write (коммит и публикация атомарны на уровне записи в БД); надёжная доставка. − Нужен relay-процесс и идемпотентные обработчики; события доставляются «хотя бы один раз».

```
usecase.Execute
  └─ uow.Do(ctx, fn):
        BEGIN tx
          repo.Save(agg)            // upsert агрегата
          outbox.Append(agg.PullEvents())  // события в той же tx
        COMMIT
  (relay, асинхронно) outbox → EventPublisher.Publish → mark published
```

## Data Flow

### Request Flow (write)

```
gRPC request
    ↓
api/grpc хендлер  (decode + protovalidate)
    ↓  Execute(ctx, input)
usecases/RegisterHostUseCase
    ↓  domain.NewHost(...) → инварианты
    ↓  uow.Do(ctx, fn)
repositories  (Mongo tx): Save(агрегат) + Append(события в outbox)
    ↓  COMMIT
gRPC reply ← encode(output)

(асинхронно) relay: outbox → EventPublisher
```

### Request Flow (read)

```
gRPC request → api/grpc хендлер → query/ListHostsQuery.Execute
    → прямой read из Mongo → []HostDTO → encode → gRPC reply
```

### Key Data Flows

1. **Команда:** inbound-адаптер → use case (`Execute`) → домен (инварианты) → `uow.Do` → репозиторий (агрегат + outbox) → COMMIT; relay публикует события.
2. **Запрос:** inbound-адаптер → query-сервис → прямой read Mongo → DTO. Без транзакций, без агрегатов.
3. **Doc-зависимости (Layer 1):** `glossary` → `structure`/`build` → `testing`/`style`/`architecture` → `git`/`boundaries`; `README` — индекс; ADR/паттерны — P2.

## Scaling Considerations

| Масштаб | Корректировки архитектуры |
|---------|---------------------------|
| Старт (1 сервис) | Монорепо-workspace; `inventory` вне `go.work`; ручной DI достаточно |
| Несколько сервисов | gRPC-контракты (buf); общие `pkg/`; per-service Mongo база |
| Рост нагрузки чтения | Отдельные read-модели/проекции; outbox-relay масштабируется отдельно |

### Scaling Priorities

1. **Первое узкое место (сервис):** граница транзакции и публикация событий — поэтому сразу фиксируем UnitOfWork + outbox как эталон, иначе каждый use case изобретёт свой способ.
2. **Второе:** read-модели под нагрузкой — отдельные проекции, идемпотентные обработчики, мониторинг лага relay.

## Anti-Patterns

### Anti-Pattern 1: Возрождать CQRS-шину / диспетчер

**What people do:** Тащить обратно `CommandDispatcher`/`QueryDispatcher`/mediatr «для единообразия».
**Why it's wrong:** Сознательно удалено как лишняя сложность; добавляет косвенность без выгоды на текущем масштабе.
**Do this instead:** inbound-адаптеры зовут use case'ы напрямую; разделение read/write — через отдельный пакет `query/`.

### Anti-Pattern 2: Транзакция/публикация событий в репозитории или «после коммита» в use case

**What people do:** Публиковать события из `Save` или сразу после коммита отдельным вызовом.
**Why it's wrong:** Dual-write — коммит может пройти, а публикация упасть (или наоборот), ломая согласованность.
**Do this instead:** transactional outbox внутри UnitOfWork-транзакции + relay.

### Anti-Pattern 3: Инфраструктура протекает в домен

**What people do:** Импорт mongo-driver / типов gRPC в `domain/` или `usecases/`.
**Why it's wrong:** Ломает гексагон; домен становится непереносимым и трудно-тестируемым.
**Do this instead:** Только порты (интерфейсы) в домене; реализации — в `repositories/`/`api/`.

## Integration Points

### External Services

| Сервис | Паттерн интеграции | Заметки |
|--------|--------------------|---------|
| MongoDB | Outbound-адаптер `repositories/` реализует порты + UnitOfWork + outbox | mongo-driver v2; транзакция в ctx; per-service база |
| Внешние сервисы (bot и т.п.) | Outbound-порт + адаптер (gRPC/HTTP клиент) | На этапе фундамента — заглушки/моки |

### Internal Boundaries

| Граница | Коммуникация | Заметки |
|---------|--------------|---------|
| api ↔ usecases/query | Прямой вызов `Execute` с DTO | Inbound-адаптер не трогает домен напрямую |
| usecases ↔ repositories | Через порты домена | Реализация подставляется в `app/` (DI) |
| между сервисами (Layer 3) | gRPC-контракты (buf) | Будущие эпики |

## Sources

- Зафиксированные решения проекта: `.planning/PROJECT.md` → Key Decisions (источник истины для целевой архитектуры)
- `.planning/research/STACK.md`, `.planning/research/FEATURES.md` (этот milestone)
- Hexagonal architecture / Ports & Adapters (Alistair Cockburn); Transactional Outbox (microservices.io)

---
*Architecture research for: AI-conventions knowledge base (primary) + canonical Go DDD/hexagonal service (no CQRS bus) + HaaS platform topology*
*Researched: 2026-06-17*
