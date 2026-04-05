package metrics

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// BusinessMetrics provides a thread-safe registry for domain event counters.
// Bounded contexts receive this via constructor injection.
type BusinessMetrics struct {
	meter    metric.Meter
	counters map[string]metric.Int64Counter
	mu       sync.RWMutex
}

// NewBusinessMetrics creates a new business metrics registry.
func NewBusinessMetrics(serviceName string) *BusinessMetrics {
	return &BusinessMetrics{
		meter:    otel.Meter(serviceName + "/business"),
		counters: make(map[string]metric.Int64Counter),
	}
}

// Inc increments a named business counter by 1.
func (b *BusinessMetrics) Inc(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if b == nil {
		return
	}

	b.mu.RLock()
	counter, ok := b.counters[name]
	b.mu.RUnlock()

	if !ok {
		counter = b.getOrCreate(name)
	}

	counter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// getOrCreate returns an existing counter or creates a new one.
func (b *BusinessMetrics) getOrCreate(name string) metric.Int64Counter {
	b.mu.Lock()
	defer b.mu.Unlock()

	if counter, ok := b.counters[name]; ok {
		return counter
	}

	counter, _ := b.meter.Int64Counter("business_"+name,
		metric.WithDescription("Business metric: "+name),
	)
	b.counters[name] = counter
	return counter
}
