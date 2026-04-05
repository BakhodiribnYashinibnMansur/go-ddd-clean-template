// Package metrics provides OpenTelemetry-based metrics with Prometheus exporter.
package metrics

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Provider holds the OTel MeterProvider and Prometheus HTTP handler.
type Provider struct {
	mp      *sdkmetric.MeterProvider
	handler http.Handler
}

// NewProvider creates an OTel MeterProvider with Prometheus exporter and starts Go runtime metrics.
func NewProvider(serviceName string) (*Provider, error) {
	reg := promclient.NewRegistry()

	exporter, err := prometheus.New(
		prometheus.WithRegisterer(reg),
		prometheus.WithoutScopeInfo(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
	)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(mp)

	// Start Go runtime metrics collection (goroutines, heap, GC).
	if err := runtime.Start(
		runtime.WithMeterProvider(mp),
		runtime.WithMinimumReadMemStatsInterval(15*time.Second),
	); err != nil {
		return nil, err
	}

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	return &Provider{mp: mp, handler: handler}, nil
}

// Handler returns the HTTP handler for the /metrics endpoint.
func (p *Provider) Handler() http.Handler {
	if p == nil {
		return http.NotFoundHandler()
	}
	return p.handler
}

// Shutdown gracefully shuts down the MeterProvider.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil || p.mp == nil {
		return nil
	}
	return p.mp.Shutdown(ctx)
}
