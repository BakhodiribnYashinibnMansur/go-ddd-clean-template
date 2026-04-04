package httpclient

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/logstore"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const tableName = "external_api_logs"

// Flusher drains the Redis buffer and bulk-inserts entries into
// external_api_logs via COPY FROM on a fixed interval. Mirrors the pattern used
// by the logger.Flusher for app_logs.
type Flusher struct {
	rdb       redis.Cmdable
	pool      *pgxpool.Pool
	key       string
	batchSize int64
	interval  time.Duration
	retention int // days; 0 = no cleanup
	log       logger.Log

	stopOnce  sync.Once
	stopCh    chan struct{}
	redisDown atomic.Bool
}

// FlusherConfig configures the persistence loop.
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

// cleanup drops expired monthly partitions (O(1) heap-free operation) and
// pre-creates upcoming ones so the write path never fails for lack of a
// partition.
func (f *Flusher) cleanup() {
	defer func() {
		if r := recover(); r != nil {
			f.log.Errorf("external_api_logs cleanup: panic recovered — %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := logstore.EnsureFuture(ctx, f.pool, tableName, 2); err != nil {
		f.log.Warnf("external_api_logs cleanup: ensure-future failed — %v", err)
	}

	if f.retention <= 0 {
		return
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -f.retention)
	dropped, err := logstore.DropOlderThan(ctx, f.pool, tableName, cutoff)
	if err != nil {
		f.log.Warnf("external_api_logs cleanup: drop partitions failed — %v", err)
		return
	}
	if dropped > 0 {
		f.log.Infof("external_api_logs cleanup: dropped %d partition(s) older than %s",
			dropped, cutoff.Format("2006-01"))
	}
}

func (f *Flusher) flush() {
	defer func() {
		if r := recover(); r != nil {
			f.log.Errorf("external_api_logs flusher: panic recovered — %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// PEEK — only trim AFTER COPY FROM succeeds, so PG failures never lose data.
	raw, err := f.rdb.LRange(ctx, f.key, -f.batchSize, -1).Result()
	if err != nil {
		if !f.redisDown.Swap(true) {
			f.log.Warnf("external_api_logs flusher: redis unavailable — %v", err)
		}
		return
	}
	if f.redisDown.Swap(false) {
		f.log.Infof("external_api_logs flusher: redis recovered")
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
	if len(entries) == 0 {
		// All entries malformed — trim to avoid an infinite retry loop.
		_ = f.rdb.LTrim(ctx, f.key, 0, -int64(len(raw))-1).Err()
		incFlushPoisoned()
		return
	}

	columns := []string{
		"api_name", "operation",
		"request_method", "request_url", "request_headers", "request_body", "request_body_size",
		"response_status", "response_headers", "response_body", "response_body_size",
		"error_text", "duration_ms",
		"request_id", "user_id", "session_id", "ip_address",
		"created_at",
	}

	rows := make([][]any, len(entries))
	for i, e := range entries {
		ts := e.Timestamp
		if ts.IsZero() {
			ts = time.Now().UTC()
		}
		rows[i] = []any{
			e.APIName, nullIfEmpty(e.Operation),
			e.RequestMethod, e.RequestURL, nullIfEmpty(e.RequestHeaders), nullIfEmpty(e.RequestBody), e.RequestBodySize,
			e.ResponseStatus, nullIfEmpty(e.ResponseHeaders), nullIfEmpty(e.ResponseBody), e.ResponseBodySize,
			nullIfEmpty(e.ErrorText), e.DurationMs,
			nullIfEmpty(e.RequestID), nullIfEmpty(e.UserID), nullIfEmpty(e.SessionID), nullIfEmpty(e.IPAddress),
			ts,
		}
	}

	if _, err := f.pool.CopyFrom(ctx,
		pgx.Identifier{tableName},
		columns,
		pgx.CopyFromRows(rows),
	); err != nil {
		// Do NOT trim — batch stays in Redis and retries next tick.
		incFlushFailed(len(entries))
		f.log.Errorc(ctx, "external_api_logs flusher: COPY FROM failed, batch will retry",
			"error", err,
			"batch_size", len(entries),
		)
		return
	}

	if err := f.rdb.LTrim(ctx, f.key, 0, -int64(len(raw))-1).Err(); err != nil {
		incFlushDupRisk()
		f.log.Warnc(ctx, "external_api_logs flusher: LTrim failed, next batch may contain duplicates",
			"error", err,
			"batch_size", len(entries),
		)
	}

	incFlushed(len(entries))
	f.log.Debugf("external_api_logs flusher: persisted %d entries", len(entries))
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
