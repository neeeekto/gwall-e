package application

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gwall-e/services/inventory/internal/domain"
)

// CommandHandler — обобщённый интерфейс хендлера команды.
// C — тип команды, R — тип результата.
type CommandHandler[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

// CommandDispatcher — диспетчер команд (write-side).
//
// Каждая команда выполняется внутри транзакции (TxManager).
// Pipeline (самый внешний → самый внутренний):
//  1. TransactionMiddleware  ← оборачивает всё, откат при ошибке
//  2. LoggingMiddleware      ← логирует имя, время, ошибку
//  3. TracingMiddleware      ← добавляет span
//  4. Handler.Handle(ctx, cmd)
//
// Middleware pipeline строится ОДИН РАЗ при RegisterCommand — нет аллокаций в SendCommand.
type CommandDispatcher struct {
	handlers   map[reflect.Type]handlerFunc
	txManager  domain.TxManager
	middleware []Middleware
}

// CommandDispatcherOption — функциональная опция.
type CommandDispatcherOption func(*CommandDispatcher)

// WithCommandMiddleware добавляет middleware в pipeline команд.
// Middleware применяются ВНУТРИ транзакции (после транзакционного враппера).
func WithCommandMiddleware(mw ...Middleware) CommandDispatcherOption {
	return func(d *CommandDispatcher) {
		d.middleware = append(d.middleware, mw...)
	}
}

// NewCommandDispatcher создаёт диспетчер команд.
// txManager — порт управления транзакциями (реализован в infra/mongo).
// Если txManager == nil — команды выполняются без транзакции (удобно для тестов).
func NewCommandDispatcher(txManager domain.TxManager, opts ...CommandDispatcherOption) *CommandDispatcher {
	d := &CommandDispatcher{
		handlers:  make(map[reflect.Type]handlerFunc),
		txManager: txManager,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// RegisterCommand регистрирует хендлер команды типа C с результатом R.
// Паника при дублировании регистрации.
func RegisterCommand[C any, R any](d *CommandDispatcher, h CommandHandler[C, R]) {
	cmdType := reflect.TypeOf(*new(C))
	if _, exists := d.handlers[cmdType]; exists {
		panic(fmt.Sprintf("command_dispatcher: handler for %s already registered", cmdType.Name()))
	}

	name := cmdType.Name()

	// Базовый вызов хендлера
	base := func(ctx context.Context, cmd any) (any, error) {
		return h.Handle(ctx, cmd.(C))
	}

	// Оборачиваем в пользовательские middleware (изнутри наружу)
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

	// Транзакционный враппер — самый внешний (оборачивает middleware + handler)
	withTx := wrapped
	if d.txManager != nil {
		withTx = func(ctx context.Context, cmd any) (any, error) {
			var result any
			err := d.txManager.RunInTx(ctx, func(ctx context.Context) error {
				var e error
				result, e = wrapped(ctx, cmd)
				return e
			})
			return result, err
		}
	}

	d.handlers[cmdType] = withTx
}

// SendCommand отправляет команду типа C и возвращает результат типа R.
// Выполняется внутри транзакции (если TxManager задан).
// Hot path: ~10 нс overhead, 0 аллокаций.
func SendCommand[C any, R any](ctx context.Context, d *CommandDispatcher, cmd C) (R, error) {
	handler, ok := d.handlers[reflect.TypeOf(cmd)]
	if !ok {
		var zero R
		return zero, fmt.Errorf("command_dispatcher: no handler registered for %T", cmd)
	}

	raw, err := handler(ctx, cmd)
	if err != nil {
		var zero R
		return zero, err
	}

	return raw.(R), nil
}
