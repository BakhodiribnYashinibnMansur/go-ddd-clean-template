package pgxutil

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "repo"

// RepoSpan starts an OTel span for a repository operation.
// Usage:
//
//	ctx, end := pgxutil.RepoSpan(ctx, "UserWriteRepo.Save")
//	defer func() { end(err) }()
func RepoSpan(ctx context.Context, name string) (context.Context, func(error)) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, name)
	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}

// RepoSpanSimple starts an OTel span and returns the span directly.
// Use when error recording is not needed.
func RepoSpanSimple(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, name)
}
