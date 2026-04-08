package metrics

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// BusinessMetrics provides a thread-safe registry for domain event counters,
// histograms and gauges. Bounded contexts receive this via constructor injection.
type BusinessMetrics struct {
	meter      metric.Meter
	counters   map[string]metric.Int64Counter
	histograms map[string]metric.Float64Histogram
	gauges     map[string]metric.Int64UpDownCounter
	mu         sync.RWMutex
}

// NewBusinessMetrics creates a new business metrics registry.
func NewBusinessMetrics(serviceName string) *BusinessMetrics {
	return &BusinessMetrics{
		meter:      otel.Meter(serviceName + "/business"),
		counters:   make(map[string]metric.Int64Counter),
		histograms: make(map[string]metric.Float64Histogram),
		gauges:     make(map[string]metric.Int64UpDownCounter),
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

// Observe records a float64 value in a named business histogram.
func (b *BusinessMetrics) Observe(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) {
	if b == nil {
		return
	}

	b.mu.RLock()
	hist, ok := b.histograms[name]
	b.mu.RUnlock()

	if !ok {
		hist = b.getOrCreateHistogram(name)
	}

	hist.Record(ctx, value, metric.WithAttributes(attrs...))
}

// Gauge adjusts a named business gauge by delta (positive = up, negative = down).
func (b *BusinessMetrics) Gauge(ctx context.Context, name string, delta int64, attrs ...attribute.KeyValue) {
	if b == nil {
		return
	}

	b.mu.RLock()
	gauge, ok := b.gauges[name]
	b.mu.RUnlock()

	if !ok {
		gauge = b.getOrCreateGauge(name)
	}

	gauge.Add(ctx, delta, metric.WithAttributes(attrs...))
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

func (b *BusinessMetrics) getOrCreateHistogram(name string) metric.Float64Histogram {
	b.mu.Lock()
	defer b.mu.Unlock()

	if hist, ok := b.histograms[name]; ok {
		return hist
	}

	hist, _ := b.meter.Float64Histogram("business_"+name,
		metric.WithDescription("Business histogram: "+name),
	)
	b.histograms[name] = hist
	return hist
}

func (b *BusinessMetrics) getOrCreateGauge(name string) metric.Int64UpDownCounter {
	b.mu.Lock()
	defer b.mu.Unlock()

	if gauge, ok := b.gauges[name]; ok {
		return gauge
	}

	gauge, _ := b.meter.Int64UpDownCounter("business_"+name,
		metric.WithDescription("Business gauge: "+name),
	)
	b.gauges[name] = gauge
	return gauge
}
