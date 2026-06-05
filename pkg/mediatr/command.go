package mediatr

import (
	"context"
	"fmt"
	"reflect"
)

type CommandHandler[Command any, Result any] interface {
	Handle(ctx context.Context, cmd Command) (Result, error)
}

type CommandDispatcher struct {
	handlers   map[reflect.Type]handlerFunc
	middleware []Middleware
}

type CommandDispatcherOption func(*CommandDispatcher)

func WithCommandMiddleware(mw ...Middleware) CommandDispatcherOption {
	return func(d *CommandDispatcher) {
		d.middleware = append(d.middleware, mw...)
	}
}

func NewCommandDispatcher(opts ...CommandDispatcherOption) *CommandDispatcher {
	d := &CommandDispatcher{
		handlers: make(map[reflect.Type]handlerFunc),
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func RegisterCommand[C any, R any](d *CommandDispatcher, h CommandHandler[C, R]) {
	cmdType := reflect.TypeOf(*new(C))
	if _, exists := d.handlers[cmdType]; exists {
		panic(fmt.Sprintf("command_dispatcher: handler for %s already registered", cmdType.Name()))
	}

	name := cmdType.Name()

	base := func(ctx context.Context, cmd any) (any, error) {
		return h.Handle(ctx, cmd.(C))
	}

	wrapped := base
	for i := len(d.middleware) - 1; i >= 0; i-- {
		mw := d.middleware[i]
		inner := wrapped
		wrapped = func(ctx context.Context, cmd any) (any, error) {
			var result any
			err := mw(ctx, name, func(ctx context.Context) error {
				var e error
				result, e = inner(ctx, cmd)
				return e
			})
			return result, err
		}
	}

	d.handlers[cmdType] = wrapped
}

func SendCommand[Command any, Response any](ctx context.Context, d *CommandDispatcher, cmd Command) (Response, error) {
	handler, ok := d.handlers[reflect.TypeOf(cmd)]
	if !ok {
		var zero Response
		return zero, fmt.Errorf("command_dispatcher: no handler registered for %T", cmd)
	}

	raw, err := handler(ctx, cmd)
	if err != nil {
		var zero Response
		return zero, err
	}

	return raw.(Response), nil
}
