package mediatr

import (
	"context"
	"fmt"
	"reflect"
)

type QueryHandler[Query any, Response any] interface {
	Handle(ctx context.Context, query Query) (Response, error)
}

type QueryDispatcher struct {
	handlers   map[reflect.Type]handlerFunc
	middleware []Middleware
}

type QueryDispatcherOption func(*QueryDispatcher)

func WithQueryMiddleware(mw ...Middleware) QueryDispatcherOption {
	return func(d *QueryDispatcher) {
		d.middleware = append(d.middleware, mw...)
	}
}

func NewQueryDispatcher(opts ...QueryDispatcherOption) *QueryDispatcher {
	d := &QueryDispatcher{
		handlers: make(map[reflect.Type]handlerFunc),
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func RegisterQuery[Query any, Response any](d *QueryDispatcher, h QueryHandler[Query, Response]) {
	queryType := reflect.TypeOf(*new(Query))
	if _, exists := d.handlers[queryType]; exists {
		panic(fmt.Sprintf("query_dispatcher: handler for %s already registered", queryType.Name()))
	}

	name := queryType.Name()

	base := func(ctx context.Context, query any) (any, error) {
		return h.Handle(ctx, query.(Query))
	}

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
