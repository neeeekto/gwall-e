# gwall-e

Gwall-E — платформа управления серверами в дата-центрах. Позволяет: видеть состояние хостов, инвентаризировать их, выписывать на них права, выполнять операции над хостами (в т.ч. массовые действия) и автолечение.

Технологии: Go-микросервисы (бэкенд) + React/Nx (фронтенд). Архитектурный подход: DDD + CQRS + гексагональная архитектура.

> 📒 **Memory Bank**: в корне есть командный банк знаний [`memory-bank/`](memory-bank/README.md) — версионируемые правила проекта (тесты, стиль и т.п.). Прочти его в начале работы и **следуй правилам оттуда**. Например, тесты: Ginkgo + Gomega, комментарии на английском ([`memory-bank/testing.md`](memory-bank/testing.md)).

> ⚠️ `README.md`, `Makefile` и `docker-compose.yml` в корне — **устаревшие**. Не доверяй им; ориентируйся на этот файл и memory-bank.

## Раскладка репозитория

Мульти-модульный Go workspace (`go.work`, Go 1.24.6); каждый сервис/пакет — отдельный модуль `github.com/gwall-e/...`. Подробности — в memory-bank:

- [`memory-bank/structure.md`](memory-bank/structure.md) — полная раскладка репозитория и workspace;
- [`memory-bank/libraries.md`](memory-bank/libraries.md) — общие библиотеки `pkg/`;
- [`memory-bank/agents.md`](memory-bank/agents.md) — агенты `agents/`.

Ключевое: `services/inventory/` — эталонная реализация (DDD/CQRS), WIP; `go.work` включает только `./pkg`, `./services/analytics`, `./services/audit`, **`inventory` НЕ в workspace** (намеренно).

## Сборка и тесты

```bash
# pkg (в составе workspace)
cd pkg && go test ./...

# inventory — НЕ в go.work, собирать с GOWORK=off из каталога модуля
cd services/inventory && GOWORK=off go build ./...
cd services/inventory && GOWORK=off go test ./...
cd services/inventory && GOWORK=off go vet ./...

# фронтенд
cd web && npx nx serve dashboard     # dev
cd web && npx nx build dashboard     # prod
```

> ⚠️ **Текущий build `inventory` сломан**: `cmd/main.go` и HTTP-хендлеры импортируют пакет `internal/application` (`CommandDispatcher`, `QueryDispatcher`, `RegisterCommand`, middleware), которого **ещё нет** — есть только подпакеты `application/commands` и `application/queries`. Это известная незавершённая работа, а не баг для случайной починки. Много хендлеров в `main.go` намеренно сконструированы с `nil`-зависимостями и пометками `// TODO`.

## Архитектура (на примере `services/inventory`)

Гексагон + CQRS. Поток: **inbound adapter → dispatcher (+middleware) → handler → domain → outbound adapter (port)**.

```
internal/
  domain/         # Агрегаты, value objects, доменные события, интерфейсы репозиториев (порты), sentinel-ошибки
  application/
    commands/     # Write-side use cases: XxxCommand (DTO) + XxxHandler.Handle(ctx, cmd)
    queries/      # Read-side use cases (через read-models, без транзакций)
    (root)        # CommandDispatcher/QueryDispatcher + middleware + Tx — TODO, ещё не реализован
  infra/
    mongo/        # Outbound: репозитории и read-models (MongoDB)
    bot/          # Outbound: HTTP-клиент внешнего сервиса
    events/       # Outbound: EventPublisher
  api/
    http/         # Inbound: chi-роутер, хендлеры. Команды → CommandDispatcher, запросы → QueryDispatcher
    cron/         # Inbound: периодические джобы (вызывают конкретный handler напрямую)
cmd/main.go       # Composition root: ручной DI, проводка всех слоёв, env-конфиг через getEnv()
```

## Конвенции кода

- **Комментарии и доменная терминология — на русском.** Сохраняй этот стиль в Go-коде.
- **Агрегаты**: приватные поля; создание через фабрику `NewX(...) (*X, error)` с проверкой инвариантов; встраивают `AggregateRoot` для накопления доменных событий; события забираются через `PullEvents()` **после** успешного `Save`.
- **Типизированные ID**: `type HostID string`, `type ProjectID string` и т.п.
- **Ошибки**: доменные правила — sentinel-ошибки (`domain.ErrHostAlreadyExists`); инфраструктурные/прикладные — оборачивать через `fmt.Errorf("...: %w", err)`.
- **Команда/запрос**: входной DTO (`XxxCommand`/`XxxQuery`) + хендлер с конструктором `NewXxxHandler(deps...)` и методом `Handle(ctx, in) (Result, error)`.
- **Маппинг DTO → домен** делается внутри хендлера, домен не знает о транспортных типах.
- Конфигурация — через переменные окружения с дефолтами (`getEnv`, `parseDuration` в `main.go`).

## Общие пакеты (`pkg/`)

См. [`memory-bank/libraries.md`](memory-bank/libraries.md).

## Заметки для агентов

- Перед правкой `inventory` помни про `GOWORK=off`, иначе `go` ругается, что каталог не в workspace.
- Не «чини» массово `nil`-зависимости и `TODO` в `main.go` — это строительные леса.
- При добавлении сервиса в общий build — добавь его модуль в `go.work` (`use`).
- Git remote: `origin` → `github.com/neeeekto/gwall-e`. Основная ветка — `main`.
