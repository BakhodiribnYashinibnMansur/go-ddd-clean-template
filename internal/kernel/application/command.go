package application

import "context"

// CommandHandler handles a command of type C.
type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}
