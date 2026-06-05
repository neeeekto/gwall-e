# Общие библиотеки (`pkg/`)

Переиспользуемые модули. Каждый — отдельный Go-модуль `github.com/gwall-e/...`, покрыт тестами. В составе `go.work`.

- **`pkg/http`** — устойчивый HTTP-клиент: `go-retryablehttp` + `gobreaker` (circuit breaker) + middlewares/transports. Покрыт тестами.
- **`pkg/mediatr`** — обобщённый mediator/dispatcher с middleware (`LoggingMiddleware`, `TracingMiddleware`; trace-имя кладётся в context). Используй как основу при реализации dispatcher-слоя в сервисах.

> Планируется `pkg/logger` — общий пакет-логгер (в работе). Добавь сюда описание, когда появится.
