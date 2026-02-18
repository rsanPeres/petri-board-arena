package port

import "context"

type UnitOfWork interface {
	WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}
