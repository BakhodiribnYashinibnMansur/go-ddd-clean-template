package reqlog

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/logstore"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const tableName = "http_request_logs"

// Flusher drains the Redis buffer and bulk-inserts entries into
// http_request_logs via COPY FROM on a fixed interval.
type Flusher struct {
	rdb       redis.Cmdable
	pool      *pgxpool.Pool
	key       string
	batchSize int64
	interval  time.Duration
	retention int
	log       logger.Log

	stopOnce  sync.Once
	stopCh    chan struct{}
	redisDown atomic.Bool
}

type FlusherConfig struct {
	RedisKey     string
	BatchSize    int
	Interval     time.Duration
	RetentionDay int
}

func NewFlusher(rdb redis.Cmdable, pool *pgxpool.Pool, cfg FlusherConfig, l logger.Log) *Flusher {
	if cfg.RedisKey == "" {
		cfg.RedisKey = DefaultRedisKey
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 1000
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Second
	}
	return &Flusher{
		rdb:       rdb,
		pool:      pool,
		key:       cfg.RedisKey,
		batchSize: int64(cfg.BatchSize),
		interval:  cfg.Interval,
		retention: cfg.RetentionDay,
		log:       l,
		stopCh:    make(chan struct{}),
	}
}

func (f *Flusher) Start() { go f.run() }

func (f *Flusher) Stop() {
	f.stopOnce.Do(func() {
		close(f.stopCh)
		f.flush()
	})
}

func (f *Flusher) run() {
	flushTicker := time.NewTicker(f.interval)
	defer flushTicker.Stop()

	cleanupTicker := time.NewTicker(time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-flushTicker.C:
			f.flush()
		case <-cleanupTicker.C:
			f.cleanup()
		case <-f.stopCh:
			return
		}
	}
}

// cleanup expires old monthly partitions and pre-creates upcoming ones.
// Partition DROP is O(1); the previous DELETE approach rewrote the whole heap.
func (f *Flusher) cleanup() {
	defer func() {
		if r := recover(); r != nil {
			f.log.Errorf("http_request_logs cleanup: panic recovered — %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Always keep 2 months of future partitions ready for writes.
	if err := logstore.EnsureFuture(ctx, f.pool, tableName, 2); err != nil {
		f.log.Warnf("http_request_logs cleanup: ensure-future failed — %v", err)
	}

	if f.retention <= 0 {
		return
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -f.retention)
	dropped, err := logstore.DropOlderThan(ctx, f.pool, tableName, cutoff)
	if err != nil {
		f.log.Warnf("http_request_logs cleanup: drop partitions failed — %v", err)
		return
	}
	if dropped > 0 {
		f.log.Infof("http_request_logs cleanup: dropped %d partition(s) older than %s",
			dropped, cutoff.Format("2006-01"))
	}
}

func (f *Flusher) flush() {
	defer func() {
		if r := recover(); r != nil {
			f.log.Errorf("http_request_logs flusher: panic recovered — %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// PEEK — read the oldest N entries without removing them. We only trim
	// after COPY FROM succeeds, so a PostgreSQL failure does NOT lose the
	// batch — next tick will retry the same entries.
	raw, err := f.rdb.LRange(ctx, f.key, -f.batchSize, -1).Result()
	if err != nil {
		if !f.redisDown.Swap(true) {
			f.log.Warnf("http_request_logs flusher: redis unavailable — %v", err)
		}
		return
	}
	if f.redisDown.Swap(false) {
		f.log.Infof("http_request_logs flusher: redis recovered")
	}
	if len(raw) == 0 {
		return
	}

	entries := make([]Entry, 0, len(raw))
	for _, r := range raw {
		var e Entry
		if err := json.Unmarshal([]byte(r), &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	// All entries were malformed — trim them so we don't retry forever.
	if len(entries) == 0 {
		_ = f.rdb.LTrim(ctx, f.key, 0, -int64(len(raw))-1).Err()
		incFlushPoisoned()
		return
	}

	columns := []string{
		"method", "path", "query", "route",
		"request_headers", "request_body", "request_body_size",
		"response_status", "response_headers", "response_body", "response_body_size",
		"duration_ms", "client_ip", "user_agent",
		"request_id", "user_id", "session_id",
		"created_at",
	}

	rows := make([][]any, len(entries))
	for i, e := range entries {
		ts := e.Timestamp
		if ts.IsZero() {
			ts = time.Now().UTC()
		}
		rows[i] = []any{
			e.Method, e.Path, nullIfEmpty(e.Query), nullIfEmpty(e.Route),
			nullIfEmpty(e.RequestHeaders), nullIfEmpty(e.RequestBody), e.RequestBodySize,
			e.ResponseStatus, nullIfEmpty(e.ResponseHeaders), nullIfEmpty(e.ResponseBody), e.ResponseBodySize,
			e.DurationMs, nullIfEmpty(e.ClientIP), nullIfEmpty(e.UserAgent),
			nullIfEmpty(e.RequestID), nullIfEmpty(e.UserID), nullIfEmpty(e.SessionID),
			ts,
		}
	}

	if _, err := f.pool.CopyFrom(ctx,
		pgx.Identifier{tableName},
		columns,
		pgx.CopyFromRows(rows),
	); err != nil {
		// DO NOT trim — the batch stays in Redis and will be retried next tick.
		incFlushFailed(len(entries))
		f.log.Errorc(ctx, "http_request_logs flusher: COPY FROM failed, batch will retry",
			"error", err,
			"batch_size", len(entries),
		)
		return
	}

	// COPY succeeded — only now do we remove the entries from Redis.
	if err := f.rdb.LTrim(ctx, f.key, 0, -int64(len(raw))-1).Err(); err != nil {
		incFlushDupRisk()
		f.log.Warnc(ctx, "http_request_logs flusher: LTrim failed, next batch may contain duplicates",
			"error", err,
			"batch_size", len(entries),
		)
	}

	incFlushed(len(entries))
	f.log.Debugf("http_request_logs flusher: persisted %d entries", len(entries))
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
