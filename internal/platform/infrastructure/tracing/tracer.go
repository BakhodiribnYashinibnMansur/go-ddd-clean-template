package telemetry

import (
	"context"

	"gct/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer initializes OpenTelemetry tracer using Jaeger exporter.
func InitTracer(ctx context.Context, cfg config.Tracing) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(cfg.Endpoint),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
		)),
		// ParentBased follows the upstream sampling decision for distributed
		// traces; the inner sampler is used only for root spans.
		sdktrace.WithSampler(sdktrace.ParentBased(sampler(cfg.SamplerRatio))),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp.Shutdown, nil
}

// sampler returns AlwaysSample for ratios >= 1.0 (avoiding per-span RNG cost)
// and TraceIDRatioBased otherwise. A ratio of 0 disables root-span sampling.
func sampler(ratio float64) sdktrace.Sampler {
	if ratio >= 1.0 {
		return sdktrace.AlwaysSample()
	}
	return sdktrace.TraceIDRatioBased(ratio)
}
