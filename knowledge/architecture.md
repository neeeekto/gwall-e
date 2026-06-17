# Архитектура gwall-e

Канон **целевой архитектуры** gwall-e: **DDD + гексагональная архитектура БЕЗ
CQRS-шины**. Здесь — инварианты и *почему* (правила слоёв, направление импортов,
транзакционная граница, события); пошаговые рецепты «как добавить use case / query /
агрегат / репозиторий» — в [patterns.md](patterns.md), который **ссылается сюда**
за правилами, а не копирует их (pointer-over-copy — [authoring.md](authoring.md)).

Все правила следуют стандарту [authoring.md](authoring.md) (MUST/SHOULD/WON'T, парность
«запрет → do», pointer-over-copy). Целевая внутренняя раскладка слоёв — канон **здесь**;
уровень модулей/workspace — в [structure.md](structure.md) (обратная ссылка уже есть там).

> Все Go-сниппеты ниже — **целевой вид / иллюстрация — НЕ из компилируемого файла**
> репозитория (no-phantom — [authoring.md](authoring.md), [boundaries.md](boundaries.md)).
> Нейтральный плейсхолдер `Order` фиксируется в [style.md](style.md) и переиспользуется
> здесь и в `patterns.md` (НЕ домен gwall-e: ни host, ни VM, ни owner). Сниппеты НЕ
> ссылаются на реальные пути вида `services/.../order.go` — таких файлов нет.

## Слои и направление импортов

Гексагон строится вокруг `domain`: зависимости **всегда направлены внутрь**, к домену.

- **MUST** держать направление импортов строго внутрь, на `domain`. `domain` импортирует
  что-либо наружу (use cases, инфраструктуру, транспорт) — **WON'T**, потому что это
  выворачивает гексагон и тянет инфраструктуру в ядро; вместо этого — `domain` объявляет
  **порты** (интерфейсы), а реализации живут в адаптерах снаружи. ⟶ planned: CI-gated
  Phase 4 (depguard)
- **MUST** объявлять в `domain` только агрегаты, value objects, доменные события и
  **порты**; конкретные реализации (Mongo, gRPC) — **WON'T** держать в домене; вместо
  этого — в `repositories`/`api`. ⟶ planned: CI-gated Phase 4 (depguard)

Направление зависимостей (стрелка = «может импортировать»):

```text
            api ───────┐
       repositories ───┤
            query ─────┼──────►  domain   (ядро; наружу не импортирует)
             cron ─────┘
                                    ▲
   app  = composition root ─────────┘  (ручной DI: связывает порты с реализациями)
   cmd  = main             ─────────┘  (точка входа: вызывает app)
```

| Слой           | Роль                                                          | Импортирует    |
| -------------- | ------------------------------------------------------------ | -------------- |
| `domain`       | Агрегаты, VO, доменные события, **порты** (интерфейсы)       | — (никого)     |
| `usecases`     | Write-side: 1 use case = struct + `Execute` (interactor)     | `domain`       |
| `query`        | Read-side: query-lite, читает хранилище напрямую в DTO       | `domain` (DTO) |
| `repositories` | Mongo-реализации портов + `UnitOfWork`                       | `domain`       |
| `api`          | gRPC-адаптеры (inbound): зовут use case напрямую             | `usecases`     |
| `cron`         | Джобы (inbound): зовут use case / relay                      | `usecases`     |
| `app`          | Composition root: ручной DI, связывает порты с реализациями  | все слои       |
| `cmd`          | `main`: точка входа, поднимает `app`                          | `app`          |

> Корректность направления (`domain` не импортирует наружу) — Manual-Only ревью
> ([03-VALIDATION.md](../.planning/phases/03-conventions-architecture-docs/03-VALIDATION.md),
> D-06): обязательный гейт.

## Write-side: use case через `Execute`

- **MUST** реализовывать один write-сценарий как **1 use case = 1 struct** с методом
  `Execute(ctx, in) (out, error)` (паттерн interactor). Складывать несколько сценариев в
  один «сервис»-божество или вводить диспетчер команд — **WON'T**, потому что это
  возрождает CQRS-шину и размывает границы; вместо этого — отдельная struct на сценарий,
  inbound-адаптер зовёт её напрямую. ⟶ convention-only (review-enforced)
- **MUST** принимать в `Execute` **доменный** вход, а не транспортный DTO; маппинг
  DTO→домен — на edge, в хендлере (канон — [style.md](style.md), не дублируется здесь).

```go
// целевой вид / иллюстрация — НЕ из компилируемого файла; плейсхолдер Order
type RegisterOrderUseCase struct {
    uow UnitOfWork // порт транзакционной границы (см. ниже)
}

// Регистрирует заказ: создаёт агрегат фабрикой и сохраняет в одной транзакции.
func (uc *RegisterOrderUseCase) Execute(ctx context.Context, in RegisterOrderInput) (RegisterOrderOutput, error) {
    order, err := NewOrder(in.SKU, in.Qty) // фабрика агрегата держит инварианты
    if err != nil {
        return RegisterOrderOutput{}, err
    }
    if err := uc.uow.Do(ctx, func(ctx context.Context) error {
        // запись агрегата + outbox-событий — внутри той же tx
        return saveOrderAndOutbox(ctx, order)
    }); err != nil {
        return RegisterOrderOutput{}, err
    }
    return RegisterOrderOutput{ID: order.ID()}, nil
}
```

## Транзакционная граница: порт `UnitOfWork`

- **MUST** очерчивать транзакционную границу портом `UnitOfWork` (`Do(ctx, fn)`),
  объявленным в `domain`; реализация (Mongo-транзакция) — в `repositories`. Тянуть в use
  case `*mongo.Session` или ручной commit/rollback — **WON'T**, потому что это связывает
  домен с инфраструктурой; вместо этого — порт `UnitOfWork`, кладущий транзакцию в `ctx`.
  ⟶ convention-only (review-enforced)
- **MUST** писать агрегат и его outbox-события **в одной** транзакции `UnitOfWork`
  (отсюда — нет dual-write, см. ниже).

## Read-side: query-lite

- **MUST** реализовывать чтение как **query-lite** напрямую в DTO, мимо агрегатов, и
  держать его отдельно от write-side. Гонять чтения через агрегаты/репозитории write-side
  — **WON'T**, потому что это раздувает домен ради проекций; вместо этого — query-сервис
  читает хранилище напрямую в read-DTO. ⟶ convention-only (review-enforced)

Это «CQRS-lite»: разделение reads/writes **без** инфраструктуры CQRS-шины.

## Доменные события: фабрики + `PullEvents` + outbox/relay

- **MUST** создавать агрегат **фабрикой** (`NewOrder(...)`), которая держит инварианты;
  агрегат накапливает доменные события и отдаёт их через `PullEvents()` для записи в
  outbox. Поднимать события «руками» в use case мимо агрегата — **WON'T**, потому что
  инварианты и события расходятся с состоянием; вместо этого — события рождаются в
  агрегате, `PullEvents()` сливает их перед сохранением. ⟶ convention-only (review-enforced)
- **MUST** писать доменные события в **transactional outbox** в **той же** транзакции, что
  и агрегат; отдельный async **relay** публикует их и помечает `published`. Публиковать
  событие прямой записью в брокер рядом с записью в БД — **WON'T**, потому что это
  dual-write (запись и публикация могут разойтись); вместо этого — outbox в одной tx +
  relay (at-least-once), что поддерживает core value «согласованность».
  ⟶ convention-only (review-enforced)

```go
// целевой вид / иллюстрация — НЕ из компилируемого файла; плейсхолдер Order
events := order.PullEvents()          // агрегат отдаёт накопленные доменные события
if err := outbox.Append(ctx, events); err != nil { // тот же ctx = та же tx
    return err
}
// relay (отдельный процесс) позже публикует и помечает published
```

## Валидация транспорта на edge

- **SHOULD** валидировать транспорт **на edge** (в `api/`, напр. protovalidate); домен
  **не** валидирует транспортные правила. Тянуть проверку транспортных тегов в домен —
  **WON'T**, потому что это связывает ядро с протоколом; вместо этого — валидация на
  адаптере, домен принимает уже доменный вход. ⟶ convention-only (review-enforced)

> Это конвенция, не описание существующей реализации: платформенные security-контроли
> (ownership-гонки, SSH-права, audit) здесь **не** документируются как действующие
> правила (no-phantom — [boundaries.md](boundaries.md)).

## MUST NOT: возрождать CQRS-шину

Снесённые подсистемы удалены **намеренно** (старый код — леса, не «сломан»):

- **WON'T** возрождать CQRS-диспетчер (`CommandDispatcher` / `QueryDispatcher`),
  `pkg/mediatr` или mediatr-подобную шину — это снятая сложность; вместо диспетчера —
  inbound-адаптер (`api`/`cron`) зовёт нужный use case **напрямую**.
  ⟶ planned: CI-gated Phase 4 (depguard на запрет импорта снесённых пакетов)
- **WON'T** возрождать `TxManager` / `tx.go` как обёртку транзакций — вместо менеджера
  транзакций — порт `UnitOfWork` в `domain` (см. выше).
  ⟶ planned: CI-gated Phase 4 (depguard на запрет импорта снесённых пакетов)
