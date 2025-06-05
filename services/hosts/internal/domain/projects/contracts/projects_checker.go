package contracts

import "context"

type ProjectChecker interface {
	CheckIdUnique(ctx context.Context, id string) (bool, error)
}
