package pgxutil

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const appTracerName = "app"

// AppSpan starts an OTel span for an application-layer (command/query) operation.
// Usage:
//
//	ctx, end := pgxutil.AppSpan(ctx, "CreateUserHandler.Handle")
//	defer func() { end(err) }()
func AppSpan(ctx context.Context, name string) (context.Context, func(error)) {
	ctx, span := otel.Tracer(appTracerName).Start(ctx, name)
	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}
