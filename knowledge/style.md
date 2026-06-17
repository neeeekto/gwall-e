# Стиль кода и язык gwall-e

Канон **языка кода/комментариев** и **проект-специфичного** Go-стиля gwall-e:
типизированные ID, sentinel-vs-обёрнутые ошибки (`%w`), маппинг DTO→домен внутри
хендлера. Это **единственное место** правила про язык комментариев (карта владения —
[boundaries.md](boundaries.md)); остальные доки ссылаются сюда, а не копируют.

Все правила следуют стандарту [authoring.md](authoring.md) (MUST/SHOULD/WON'T, парность
«запрет → do», pointer-over-copy). Формат каждого правила — **«правило + плохо/хорошо»**:
короткий тег силы + мини-пример. Механизируемые правила несут пред-пометку будущего
enforcement-статуса (Phase 4 переключит её на фактический, а не ретрофитит с нуля).

**Общий Go-стиль НЕ дублируется.** Формат, нейминг, идиомы общего уровня — это
[Effective Go](https://go.dev/doc/effective_go) и автоформат **gofumpt** (planned: hook
Phase 4); повторять их здесь — **WON'T**, потому что общий стиль уже задан индустрией и
тулингом. Здесь — только то, что специфично для gwall-e.

> Все Go-сниппеты ниже — **иллюстрация** на нейтральном плейсхолдере `Order` (НЕ домен
> gwall-e: ни host, ни VM, ни owner). Это целевой вид правила, **не** код из
> компилируемого файла репозитория (no-phantom — [authoring.md](authoring.md)).
> Плейсхолдер `Order` фиксируется здесь и переиспользуется в `architecture.md`/`patterns.md`.

## Язык кода и комментариев

Это **канон языка** для всего репозитория — единственное место правила (карта владения,
[boundaries.md](boundaries.md)). Другие доки (например `testing.md` про комментарии в
тестах) **MUST** ссылаться сюда, а не повторять правило.

- **MUST** писать комментарии в коде и доменную терминологию **на русском** — это часть
  единого ubiquitous language. ⟶ convention-only (review-enforced)
- **MUST** давать имена идентификаторов (типы, функции, переменные, пакеты) **на
  английском**. Русские имена идентификаторов — **WON'T**, потому что ломают идиоматику Go
  и инструментарий; вместо этого — английское имя + русский комментарий. ⟶ convention-only
  (review-enforced)
- **MUST** писать комментарии в тестах **на английском** (тесты — техническая, не доменная
  область). Полная конвенция тестов — `testing.md` (planned, Phase 3), здесь — только
  правило языка. ⟶ convention-only (review-enforced)

```go
// плохо: имя идентификатора на русском, комментарий смешан
func ПолучитьЗаказ(id string) (*Order, error) // GetOrder но русским именем

// хорошо: имя — английское, доменный комментарий — русский
// Получает заказ по идентификатору; возвращает ErrOrderNotFound, если не найден.
func GetOrder(id OrderID) (*Order, error)
```

## Типизированные ID

- **MUST** использовать типизированный ID (`type OrderID string`) вместо «голой» строки
  для идентификаторов агрегатов/сущностей. Голый `string` для ID — **WON'T**, потому что
  компилятор не отличит один ID от другого и легко перепутать аргументы; вместо этого —
  именованный тип, который компилятор проверяет. ⟶ planned: CI-gated Phase 4 (linter)

```go
// плохо: «голая» строка — легко перепутать с любым другим string
func GetOrder(id string) (*Order, error)

// хорошо: типизированный ID — компилятор ловит подмену
type OrderID string
func GetOrder(id OrderID) (*Order, error)
```

## Sentinel vs обёрнутые ошибки

- **MUST** объявлять предсказуемые доменные ошибки как **sentinel**
  (`var ErrOrderNotFound = errors.New(...)`), чтобы вызывающий код сравнивал их через
  `errors.Is`. ⟶ convention-only (review-enforced)
- **MUST** оборачивать ошибки через `%w` (`fmt.Errorf("...: %w", err)`), сохраняя цепочку.
  Глотать причину через `%v` — **WON'T**, потому что `errors.Is`/`errors.As` теряют
  sentinel и теряется контекст; вместо этого — `%w`, который сохраняет и текст, и цепочку.
  ⟶ planned: CI-gated Phase 4 (linter, напр. errorlint)

```go
var ErrOrderNotFound = errors.New("order not found")

// плохо: %v рвёт цепочку — errors.Is(err, ErrOrderNotFound) вернёт false
return fmt.Errorf("get order %s: %v", id, ErrOrderNotFound)

// хорошо: %w сохраняет sentinel в цепочке — errors.Is работает
return fmt.Errorf("get order %s: %w", id, ErrOrderNotFound)
```

## Маппинг DTO→домен внутри хендлера

- **MUST** декодить и валидировать транспорт (gRPC/HTTP DTO) на **edge** и маппить DTO→домен
  **внутри хендлера**: домен не знает о транспорте. Тянуть транспортные DTO/теги в домен
  или принимать DTO в use case — **WON'T**, потому что это связывает домен с транспортом и
  ломает границы гексагона; вместо этого — хендлер собирает доменный вход из DTO, а use
  case принимает доменные типы. ⟶ planned: CI-gated Phase 4 (depguard)

```go
// плохо: транспортный DTO протекает в use case — домен знает про gRPC
func (uc *RegisterOrderUseCase) Execute(ctx context.Context, req *pb.RegisterOrderRequest) ...

// хорошо: хендлер маппит DTO→домен, use case принимает доменный вход
func (h *OrderHandler) Register(ctx context.Context, req *pb.RegisterOrderRequest) (*pb.RegisterOrderResponse, error) {
    in := RegisterOrderInput{SKU: req.GetSku(), Qty: int(req.GetQty())} // DTO→домен на edge
    out, err := h.uc.Execute(ctx, in)
    if err != nil {
        return nil, err
    }
    return &pb.RegisterOrderResponse{Id: string(out.ID)}, nil // домен→DTO обратно на edge
}
```

Архитектурные правила слоёв (направление импортов, `UnitOfWork`, outbox) — канон в
`architecture.md` (planned, Phase 3); здесь — только стиль кода уровня файла, не
архитектура. Команды сборки/тестов — канон в [build.md](build.md). Дублировать их здесь —
**WON'T**; вместо этого — ссылка.
