# Рецепты: как добавить слой gwall-e

Этот док — **копируемые пошаговые рецепты** «как добавить use case / query / агрегат /
репозиторий». Здесь — *how-to* (вертикальный срез от struct до wiring), а **не** правила:
за инвариантами, направлением импортов, транзакционной границей и событиями — ссылка на
[architecture.md](architecture.md); за языком кода, типизированными ID, ошибками и маппингом
DTO→домен — ссылка на [style.md](style.md). Правила эти доки держат канонически, а
`patterns.md` их **НЕ дублирует** (pointer-over-copy — [authoring.md](authoring.md)).

Все правила следуют стандарту [authoring.md](authoring.md) (MUST/SHOULD/WON'T, парность
«запрет → do»).

> Все Go-сниппеты ниже — **целевой вид / иллюстрация — НЕ из компилируемого файла**
> репозитория (no-phantom — [authoring.md](authoring.md), [boundaries.md](boundaries.md)).
> Нейтральный плейсхолдер `Order` фиксируется в [style.md](style.md) и переиспользуется
> в [architecture.md](architecture.md) и здесь (НЕ домен gwall-e: ни host, ни VM, ни owner).
> Сниппеты НЕ ссылаются на реальные пути вида `services/.../order.go` — таких файлов нет;
> копируя рецепт, создавайте новый код, а не «дописывайте» несуществующий.

Целевая раскладка слоёв (`domain/usecases/query/repositories/api/cron/app/cmd`) и
направление импортов — канон в [architecture.md](architecture.md) §«Слои и направление
импортов»; рецепты ниже опираются на неё, не переописывая.

## Рецепт 1: добавить use case (write-side)

Полный вертикальный срез write-сценария — от доменного `Execute` до gRPC-адаптера.

- **MUST** реализовывать сценарий как **1 use case = 1 struct + `Execute`** (interactor) —
  правило и *почему* (запрет бога-сервиса/диспетчера команд) см. [architecture.md](architecture.md)
  §«Write-side». Маппинг транспортного DTO→домен — **на edge, в хендлере** (правило —
  [style.md](style.md) §«Маппинг DTO→домен»), use case принимает доменный вход. ⟶ convention-only (review-enforced)

Шаги:

1. **struct + `Execute`** в `usecases`: зависимости — только **порты** из `domain`.
2. **uow.Do(...)**: запись агрегата и его outbox-событий — в одной транзакции (правило
   транзакционной границы см. [architecture.md](architecture.md) §«UnitOfWork»; не тянуть
   `*mongo.Session` в use case).
3. **Composition root** (`app`, ручной DI): собрать use case, связав порты с реализациями.
4. **gRPC-адаптер** в `api`: хендлер маппит DTO→домен, зовёт `Execute` напрямую, маппит
   результат домен→DTO обратно.

```go
// целевой вид / иллюстрация — НЕ из компилируемого файла; плейсхолдер Order

// 1) usecases: struct + Execute, зависимости — порты из domain
type RegisterOrderUseCase struct {
    orders OrderRepository // порт
    uow    UnitOfWork      // порт транзакционной границы
}

func (uc *RegisterOrderUseCase) Execute(ctx context.Context, in RegisterOrderInput) (RegisterOrderOutput, error) {
    order, err := NewOrder(in.SKU, in.Qty) // фабрика держит инварианты (см. Рецепт 3)
    if err != nil {
        return RegisterOrderOutput{}, err
    }
    // 2) одна транзакция: агрегат + outbox-события (правило — architecture.md §UnitOfWork)
    if err := uc.uow.Do(ctx, func(ctx context.Context) error {
        return uc.orders.Save(ctx, order)
    }); err != nil {
        return RegisterOrderOutput{}, fmt.Errorf("register order: %w", err) // %w — style.md
    }
    return RegisterOrderOutput{ID: order.ID()}, nil
}

// 3) app (composition root, ручной DI): связать порты с реализациями
uc := &RegisterOrderUseCase{orders: mongoOrders, uow: mongoUoW}

// 4) api: gRPC-адаптер маппит DTO→домен и зовёт Execute напрямую (без диспетчера)
func (h *OrderHandler) Register(ctx context.Context, req *pb.RegisterOrderRequest) (*pb.RegisterOrderResponse, error) {
    in := RegisterOrderInput{SKU: req.GetSku(), Qty: int(req.GetQty())} // DTO→домен на edge
    out, err := h.uc.Execute(ctx, in)
    if err != nil {
        return nil, err
    }
    return &pb.RegisterOrderResponse{Id: string(out.ID)}, nil
}
```

## Рецепт 2: добавить query (read-side, DTO)

Чтение реализуется как **query-lite** — отдельно от write-side, напрямую в read-DTO мимо
агрегатов (правило и *почему* — [architecture.md](architecture.md) §«Read-side: query-lite»).

- **MUST** держать read-side отдельной struct, читающей хранилище **сразу в DTO**; гонять
  чтение через write-репозитории/агрегаты — **WON'T** (см. architecture.md), вместо этого —
  свой query-сервис. ⟶ convention-only (review-enforced)

Шаги:

1. **struct + метод** в `query`: возвращает read-DTO, не доменный агрегат.
2. **Composition root** (`app`): собрать query с read-зависимостью (коллекция/драйвер).
3. **gRPC-адаптер** в `api`: маппит read-DTO → ответный protobuf.

```go
// целевой вид / иллюстрация — НЕ из компилируемого файла; плейсхолдер Order

// 1) query: читает напрямую в read-DTO, мимо агрегатов
type ListOrdersQuery struct {
    orders OrderReadStore // read-порт; НЕ write-репозиторий
}

func (q *ListOrdersQuery) Execute(ctx context.Context, in ListOrdersInput) ([]OrderView, error) {
    return q.orders.List(ctx, in.Limit) // OrderView — read-DTO, не агрегат Order
}

// 2) app: wiring read-зависимости; 3) api: OrderView → protobuf-ответ на edge
```

## Рецепт 3: добавить aggregate (+ фабрика, `PullEvents`)

Агрегат создаётся **фабрикой** с инвариантами и накапливает доменные события, которые
сливаются через `PullEvents()` перед записью в outbox (правило, *почему* и связка с
outbox/relay — [architecture.md](architecture.md) §«Доменные события»).

- **MUST** создавать агрегат фабрикой `NewOrder(...)`, держащей инварианты, и поднимать
  события **внутри** агрегата; поднимать события «руками» в use case мимо агрегата —
  **WON'T** (см. architecture.md), вместо этого — `PullEvents()` сливает накопленное. ⟶ convention-only (review-enforced)

Шаги (всё в `domain`):

1. **Фабрика** `NewOrder(...)`: валидирует вход и возвращает агрегат с типизированным ID
   (`OrderID` — см. [style.md](style.md) §«Типизированные ID»).
2. **События**: на доменных переходах агрегат записывает событие во внутренний буфер.
3. **`PullEvents()`**: отдаёт накопленные события (репозиторий пишет их в outbox в той же
   транзакции — см. Рецепт 4 и architecture.md §UnitOfWork/outbox).

```go
// целевой вид / иллюстрация — НЕ из компилируемого файла; плейсхолдер Order

// 1) фабрика держит инварианты
func NewOrder(sku string, qty int) (*Order, error) {
    if qty <= 0 {
        return nil, ErrInvalidQty // sentinel — style.md §ошибки
    }
    o := &Order{id: newOrderID(), sku: sku, qty: qty}
    o.record(OrderRegistered{ID: o.id}) // 2) событие копится в агрегате
    return o, nil
}

// 3) репозиторий перед записью сливает события: outbox.Append(order.PullEvents())
func (o *Order) PullEvents() []DomainEvent { /* отдаёт и очищает буфер */ }
```

## Рецепт 4: добавить repository (Mongo-реализация порта + UnitOfWork)

Репозиторий — это **реализация доменного порта** в `repositories` (Mongo); он участвует в
одной транзакции через `UnitOfWork` и пишет агрегат вместе с его outbox-событиями (правило
единой tx и запрет dual-write — [architecture.md](architecture.md) §«UnitOfWork» / §«события»).

- **MUST** реализовывать порт `OrderRepository` (объявленный в `domain`) в `repositories`,
  участвуя в транзакции через `UnitOfWork`, и писать агрегат + outbox в **одной** tx;
  публиковать событие прямой записью в брокер рядом с БД — **WON'T** (dual-write, см.
  architecture.md), вместо этого — outbox в той же tx + relay. ⟶ convention-only (review-enforced)

Шаги:

1. **Реализация порта** в `repositories`: тип реализует интерфейс `OrderRepository` из
   `domain` (домен про Mongo не знает).
2. **Одна tx**: `Save` берёт транзакцию из `ctx` (её кладёт `UnitOfWork.Do`) и пишет агрегат
   и `outbox.Append(order.PullEvents())` в этой же транзакции.
3. **Composition root** (`app`): связать порт `OrderRepository` и `UnitOfWork` с
   Mongo-реализациями и передать их в use case (Рецепт 1).

```go
// целевой вид / иллюстрация — НЕ из компилируемого файла; плейсхолдер Order

// 1) repositories: Mongo-реализация доменного порта OrderRepository
type mongoOrderRepo struct{ /* коллекции/драйвер */ }

func (r *mongoOrderRepo) Save(ctx context.Context, o *Order) error {
    // 2) транзакция уже в ctx (положена UnitOfWork.Do); агрегат + outbox в одной tx
    if err := r.upsert(ctx, o); err != nil {
        return fmt.Errorf("save order %s: %w", o.ID(), err) // %w — style.md
    }
    return r.outbox.Append(ctx, o.PullEvents()) // та же tx — нет dual-write
}

// 3) app: mongoOrders := &mongoOrderRepo{...}; mongoUoW := newMongoUnitOfWork(...)
//          uc := &RegisterOrderUseCase{orders: mongoOrders, uow: mongoUoW}
```

---

Что **WON'T** делать в рецептах: возрождать CQRS-диспетчер / `pkg/mediatr` / `TxManager` —
запрет и *почему* см. [architecture.md](architecture.md) §«MUST NOT: возрождать CQRS-шину»;
inbound-адаптер (`api`/`cron`) зовёт use case напрямую.
