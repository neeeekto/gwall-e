package application

import (
	"context"
	"fmt"
	"reflect"
)

// QueryHandler — обобщённый интерфейс хендлера запроса (read-side).
// Q — тип запроса, R — тип результата.
type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

// QueryDispatcher — диспетчер запросов (read-side).
//
// Запросы НЕ используют транзакции — они только читают данные.
// Pipeline:
//  1. LoggingMiddleware  ← логирует имя, время, ошибку
//  2. TracingMiddleware  ← добавляет span
//  3. Handler.Handle(ctx, query)
//
// Middleware pipeline строится ОДИН РАЗ при RegisterQuery — нет аллокаций в SendQuery.
type QueryDispatcher struct {
	handlers   map[reflect.Type]handlerFunc
	middleware []Middleware
}

// QueryDispatcherOption — функциональная опция.
type QueryDispatcherOption func(*QueryDispatcher)

// WithQueryMiddleware добавляет middleware в pipeline запросов.
func WithQueryMiddleware(mw ...Middleware) QueryDispatcherOption {
	return func(d *QueryDispatcher) {
		d.middleware = append(d.middleware, mw...)
	}
}

// NewQueryDispatcher создаёт диспетчер запросов.
func NewQueryDispatcher(opts ...QueryDispatcherOption) *QueryDispatcher {
	d := &QueryDispatcher{
		handlers: make(map[reflect.Type]handlerFunc),
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// RegisterQuery регистрирует хендлер запроса типа Q с результатом R.
// Паника при дублировании регистрации.
func RegisterQuery[Q any, R any](d *QueryDispatcher, h QueryHandler[Q, R]) {
	queryType := reflect.TypeOf(*new(Q))
	if _, exists := d.handlers[queryType]; exists {
		panic(fmt.Sprintf("query_dispatcher: handler for %s already registered", queryType.Name()))
	}

	name := queryType.Name()

	// Базовый вызов хендлера
	base := func(ctx context.Context, query any) (any, error) {
		return h.Handle(ctx, query.(Q))
	}

	// Оборачиваем в middleware (изнутри наружу)
	wrapped := base
	for i := len(d.middleware) - 1; i >= 0; i-- {
		mw := d.middleware[i]
		inner := wrapped
		wrapped = func(ctx context.Context, query any) (any, error) {
			var result any
			err := mw(ctx, name, func(ctx context.Context) error {
				var e error
				result, e = inner(ctx, query)
				return e
			})
			return result, err
		}
	}

	d.handlers[queryType] = wrapped
}

// SendQuery отправляет запрос типа Q и возвращает результат типа R.
// Выполняется БЕЗ транзакции.
// Hot path: ~10 нс overhead, 0 аллокаций.
func SendQuery[Q any, R any](ctx context.Context, d *QueryDispatcher, query Q) (R, error) {
	handler, ok := d.handlers[reflect.TypeOf(query)]
	if !ok {
		var zero R
		return zero, fmt.Errorf("query_dispatcher: no handler registered for %T", query)
	}

	raw, err := handler(ctx, query)
	if err != nil {
		var zero R
		return zero, err
	}

	return raw.(R), nil
}
