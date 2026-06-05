package mediatr

import "context"

type handlerFunc func(ctx context.Context, cmd any) (any, error)

type Middleware = func(ctx context.Context, name string, next func(context.Context) error) error
