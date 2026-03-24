package application

import "context"

// QueryHandler handles a query of type Q and returns a result of type R.
type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}
