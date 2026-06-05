package mediatr

import (
	"context"
	"log"
	"time"
)

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

type dispatchNameKey struct{}

func DispatchNameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(dispatchNameKey{}).(string); ok {
		return v
	}
	return ""
}
