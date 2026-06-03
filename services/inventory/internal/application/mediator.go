// Package application содержит общие типы для CommandDispatcher и QueryDispatcher.
//
// Архитектура диспетчеров:
//
//	CommandDispatcher — для команд (write-side): транзакция + logging + tracing
//	QueryDispatcher   — для запросов (read-side): только logging + tracing, без транзакции
//
// Общие типы (этот файл):
//   - handlerFunc — внутренний тип хранения скомпилированного хендлера
//   - Middleware   — тип middleware (используется в обоих диспетчерах)
//   - LoggingMiddleware, TracingMiddleware — стандартные реализации
package application

import (
	"context"
	"log"
	"time"
)

// handlerFunc — внутренний тип для хранения скомпилированного хендлера.
// Middleware pipeline встроен при Register — нет аллокаций в Send/hot path.
type handlerFunc func(ctx context.Context, cmd any) (any, error)

// Middleware — функция-обёртка вокруг выполнения хендлера.
// name — имя команды/запроса для логирования/трейсинга.
// next — следующий хендлер в цепочке.
//
// Используется как CommandMiddleware и QueryMiddleware через type alias.
type Middleware = func(ctx context.Context, name string, next func(context.Context) error) error

// ============================================================
// Стандартные middleware (применимы к командам и запросам)
// ============================================================

// LoggingMiddleware логирует имя, длительность и ошибку каждой команды/запроса.
func LoggingMiddleware() Middleware {
	return func(ctx context.Context, name string, next func(context.Context) error) error {
		start := time.Now()
		log.Printf("dispatch: %s started", name)

		err := next(ctx)

		elapsed := time.Since(start)
		if err != nil {
			log.Printf("dispatch: %s failed in %s: %v", name, elapsed, err)
		} else {
			log.Printf("dispatch: %s completed in %s", name, elapsed)
		}
		return err
	}
}

// TracingMiddleware добавляет трейсинг-спан к каждой команде/запросу.
// Заглушка — заменить на opentelemetry в production:
//
//	ctx, span := otel.Tracer("inventory").Start(ctx, "dispatch."+name)
//	defer span.End()
func TracingMiddleware() Middleware {
	return func(ctx context.Context, name string, next func(context.Context) error) error {
		// TODO: otel span
		// ctx, span := otel.Tracer("inventory").Start(ctx, "dispatch."+name)
		// defer span.End()
		ctx = context.WithValue(ctx, dispatchNameKey{}, name)
		err := next(ctx)
		if err != nil {
			// span.RecordError(err)
		}
		return err
	}
}

// dispatchNameKey — ключ для хранения имени команды/запроса в контексте.
type dispatchNameKey struct{}

// DispatchNameFromContext возвращает имя текущей команды/запроса из контекста.
func DispatchNameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(dispatchNameKey{}).(string); ok {
		return v
	}
	return ""
}
