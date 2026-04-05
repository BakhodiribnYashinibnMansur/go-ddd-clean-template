package postgres

import (
	"context"
	"strings"
	"time"

	"gct/internal/platform/infrastructure/logger"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type queryCtxKey struct{}

type queryCtxData struct {
	start     time.Time
	sql       string
	argsCount int
}

// MetricsTracer wraps a pgx.QueryTracer to record query duration metrics and log slow queries.
// It also delegates BatchTracer, CopyFromTracer, PrepareTracer, and ConnectTracer
// to the inner tracer so that Jaeger traces are preserved for all operation types.
type MetricsTracer struct {
	inner         pgx.QueryTracer
	queryDuration metric.Float64Histogram
	slowThreshold time.Duration
	logger        logger.Log
}

// Compile-time interface checks.
var (
	_ pgx.QueryTracer    = (*MetricsTracer)(nil)
	_ pgx.BatchTracer    = (*MetricsTracer)(nil)
	_ pgx.CopyFromTracer = (*MetricsTracer)(nil)
	_ pgx.PrepareTracer  = (*MetricsTracer)(nil)
	_ pgx.ConnectTracer  = (*MetricsTracer)(nil)
)

// NewMetricsTracer creates a composite tracer that delegates to inner and adds metrics + slow query logging.
func NewMetricsTracer(inner pgx.QueryTracer, l logger.Log, slowThreshold time.Duration) *MetricsTracer {
	meter := otel.Meter("db/postgres")

	queryDuration, _ := meter.Float64Histogram("db_query_duration_seconds",
		metric.WithDescription("Duration of database queries in seconds"),
		metric.WithExplicitBucketBoundaries(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5),
	)

	return &MetricsTracer{
		inner:         inner,
		queryDuration: queryDuration,
		slowThreshold: slowThreshold,
		logger:        l,
	}
}

// ─── QueryTracer ────────────────────────────────────────────────────────────────

// TraceQueryStart delegates to the inner tracer and stores start time + SQL in context.
func (t *MetricsTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx = context.WithValue(ctx, queryCtxKey{}, &queryCtxData{
		start:     time.Now(),
		sql:       data.SQL,
		argsCount: len(data.Args),
	})
	if t.inner != nil {
		ctx = t.inner.TraceQueryStart(ctx, conn, data)
	}
	return ctx
}

// TraceQueryEnd records the query duration metric and logs slow queries.
func (t *MetricsTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	if t.inner != nil {
		t.inner.TraceQueryEnd(ctx, conn, data)
	}

	qd, ok := ctx.Value(queryCtxKey{}).(*queryCtxData)
	if !ok || qd == nil {
		return
	}

	duration := time.Since(qd.start)
	operation := extractOperation(qd.sql)

	t.queryDuration.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attribute.String("operation", operation)),
	)

	if duration >= t.slowThreshold {
		sql := qd.sql
		if len(sql) > 200 {
			sql = sql[:200] + "..."
		}

		t.logger.Warnc(ctx, "Slow query detected",
			"sql", sql,
			"duration", duration.String(),
			"duration_ms", duration.Milliseconds(),
			"operation", operation,
			"args_count", qd.argsCount,
		)
	}
}

// ─── BatchTracer ────────────────────────────────────────────────────────────────

func (t *MetricsTracer) TraceBatchStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	if bt, ok := t.inner.(pgx.BatchTracer); ok {
		return bt.TraceBatchStart(ctx, conn, data)
	}
	return ctx
}

func (t *MetricsTracer) TraceBatchQuery(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchQueryData) {
	if bt, ok := t.inner.(pgx.BatchTracer); ok {
		bt.TraceBatchQuery(ctx, conn, data)
	}
}

func (t *MetricsTracer) TraceBatchEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchEndData) {
	if bt, ok := t.inner.(pgx.BatchTracer); ok {
		bt.TraceBatchEnd(ctx, conn, data)
	}
}

// ─── CopyFromTracer ─────────────────────────────────────────────────────────────

func (t *MetricsTracer) TraceCopyFromStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromStartData) context.Context {
	if ct, ok := t.inner.(pgx.CopyFromTracer); ok {
		return ct.TraceCopyFromStart(ctx, conn, data)
	}
	return ctx
}

func (t *MetricsTracer) TraceCopyFromEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
	if ct, ok := t.inner.(pgx.CopyFromTracer); ok {
		ct.TraceCopyFromEnd(ctx, conn, data)
	}
}

// ─── PrepareTracer ──────────────────────────────────────────────────────────────

func (t *MetricsTracer) TracePrepareStart(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareStartData) context.Context {
	if pt, ok := t.inner.(pgx.PrepareTracer); ok {
		return pt.TracePrepareStart(ctx, conn, data)
	}
	return ctx
}

func (t *MetricsTracer) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData) {
	if pt, ok := t.inner.(pgx.PrepareTracer); ok {
		pt.TracePrepareEnd(ctx, conn, data)
	}
}

// ─── ConnectTracer ──────────────────────────────────────────────────────────────

func (t *MetricsTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	if ct, ok := t.inner.(pgx.ConnectTracer); ok {
		return ct.TraceConnectStart(ctx, data)
	}
	return ctx
}

func (t *MetricsTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	if ct, ok := t.inner.(pgx.ConnectTracer); ok {
		ct.TraceConnectEnd(ctx, data)
	}
}

// ─── Helpers ────────────────────────────────────────────────────────────────────

// extractOperation parses the SQL operation type from the query prefix.
func extractOperation(sql string) string {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return "unknown"
	}

	end := strings.IndexByte(sql, ' ')
	if end == -1 {
		end = len(sql)
	}

	switch strings.ToUpper(sql[:end]) {
	case "SELECT":
		return "SELECT"
	case "INSERT":
		return "INSERT"
	case "UPDATE":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	case "WITH":
		return "WITH"
	case "BEGIN", "COMMIT", "ROLLBACK":
		return "TX"
	default:
		return "OTHER"
	}
}
