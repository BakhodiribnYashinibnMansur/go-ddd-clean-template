package logger

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap/zapcore"
)

// logEntry represents a single log record buffered in Redis.
type logEntry struct {
	Level     string `json:"level"`
	Message   string `json:"msg"`
	Caller    string `json:"caller"`
	Operation string `json:"operation,omitempty"`
	Entity    string `json:"entity,omitempty"`
	EntityID  string `json:"entity_id,omitempty"`
	ErrorText string `json:"error,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Extra     string `json:"extra,omitempty"`
	Timestamp string `json:"ts"`
}

// PersistConfig configures log persistence to Redis + PostgreSQL.
type PersistConfig struct {
	Level     string // minimum level to persist (debug, info, warn, error)
	RedisKey  string // Redis list key
	BatchSize    int // max entries per flush
	Interval     time.Duration
	RetentionDay int // days to keep logs in PostgreSQL (0 = no cleanup)
}

// ── Redis Sink (zap Core wrapper) ──────────────────────────────────────────

// RedisSink is a zapcore.Core that intercepts log entries above the configured
// level and pushes them to a Redis list for later flushing to PostgreSQL.
// It degrades gracefully: if Redis is unavailable, writes are silently dropped.
type RedisSink struct {
	rdb      redis.Cmdable
	key      string
	minLevel zapcore.Level
	pushTimeout time.Duration
}

// NewRedisSink creates a zap core that pushes log entries to Redis.
func NewRedisSink(rdb redis.Cmdable, cfg PersistConfig) zapcore.Core {
	return &RedisSink{
		rdb:         rdb,
		key:         cfg.RedisKey,
		minLevel:    parseLevel(cfg.Level),
		pushTimeout: 100 * time.Millisecond,
	}
}

func (s *RedisSink) Enabled(lvl zapcore.Level) bool {
	return lvl >= s.minLevel
}

func (s *RedisSink) With(fields []zapcore.Field) zapcore.Core {
	return s // fields are captured per-entry, not globally
}

func (s *RedisSink) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if s.Enabled(ent.Level) {
		return ce.AddCore(ent, s)
	}
	return ce
}

func (s *RedisSink) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	entry := logEntry{
		Level:     ent.Level.String(),
		Message:   ent.Message,
		Caller:    ent.Caller.TrimmedPath(),
		Timestamp: ent.Time.UTC().Format(time.RFC3339Nano),
	}

	// Extract known structured fields, collect the rest as extra
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}

	// Pull out known fields from the encoded map
	for k, v := range enc.Fields {
		str := fieldToString(v)
		switch k {
		case "operation":
			entry.Operation = str
		case "entity":
			entry.Entity = str
		case "entity_id":
			entry.EntityID = str
		case "error":
			entry.ErrorText = str
		case "meta_data":
			// meta_data contains request_id, user_id, etc.
			if m, ok := v.(map[string]any); ok {
				if rid, ok := m["request_id"].(string); ok {
					entry.RequestID = rid
				}
				if uid, ok := m["user_id"]; ok {
					entry.UserID = fieldToString(uid)
				}
				if sid, ok := m["session_id"].(string); ok {
					entry.SessionID = sid
				}
				if ip, ok := m["ip_address"].(string); ok {
					entry.IPAddress = ip
				}
			}
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return nil // don't block logging on marshal errors
	}

	// Fire-and-forget LPUSH with short timeout — silent fail if Redis is down.
	// Invoked from the zapcore Write path: no caller ctx is available and we
	// must not inherit request cancellation or the log entry would be lost.
	ctx, cancel := context.WithTimeout(context.Background(), s.pushTimeout)
	defer cancel()
	_ = s.rdb.LPush(ctx, s.key, data).Err()
	return nil
}

func (s *RedisSink) Sync() error { return nil }

// ── PostgreSQL Flusher ─────────────────────────────────────────────────────

// Flusher reads log entries from Redis and bulk-inserts them into PostgreSQL
// using COPY FROM at a configurable interval.
// When Redis is unavailable, flushes are silently skipped — the error is logged
// once on transition to unhealthy and once on recovery, avoiding log spam.
type Flusher struct {
	rdb       redis.Cmdable
	pool      *pgxpool.Pool
	cfg       PersistConfig
	stopOnce  sync.Once
	stopCh    chan struct{}
	log       Log
	redisDown atomic.Bool // true while Redis is unreachable
}

// NewFlusher creates a new log flusher.
func NewFlusher(rdb redis.Cmdable, pool *pgxpool.Pool, cfg PersistConfig, l Log) *Flusher {
	return &Flusher{
		rdb:    rdb,
		pool:   pool,
		cfg:    cfg,
		stopCh: make(chan struct{}),
		log:    l,
	}
}

// Start begins the periodic flush loop in a background goroutine.
func (f *Flusher) Start() {
	go f.run()
}

// Stop gracefully stops the flusher and performs a final flush.
func (f *Flusher) Stop() {
	f.stopOnce.Do(func() {
		close(f.stopCh)
		// Final flush
		f.flush()
	})
}

func (f *Flusher) run() {
	flushTicker := time.NewTicker(f.cfg.Interval)
	defer flushTicker.Stop()

	// Cleanup old logs once per hour
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

func (f *Flusher) cleanup() {
	if f.cfg.RetentionDay <= 0 {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			f.log.Errorf("log cleanup: panic recovered — %v", r)
		}
	}()

	// Runs on the flusher's background goroutine — no caller context exists.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cutoff := time.Now().AddDate(0, 0, -f.cfg.RetentionDay)
	tag, err := f.pool.Exec(ctx, "DELETE FROM app_logs WHERE created_at < $1", cutoff)
	if err != nil {
		f.log.Warnf("log cleanup: failed — %v", err)
		return
	}
	if tag.RowsAffected() > 0 {
		f.log.Infof("log cleanup: deleted %d old entries (before %s)", tag.RowsAffected(), cutoff.Format("2006-01-02"))
	}
}

func (f *Flusher) flush() {
	defer func() {
		if r := recover(); r != nil {
			f.log.Errorf("log flusher: panic recovered — %v", r)
		}
	}()

	// Runs on the flusher's background goroutine — no caller context exists.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	batchSize := int64(f.cfg.BatchSize)
	if batchSize <= 0 {
		batchSize = 1000
	}

	// Atomically pop up to batchSize entries from Redis
	pipe := f.rdb.Pipeline()
	lrangeCmd := pipe.LRange(ctx, f.cfg.RedisKey, -batchSize, -1)
	pipe.LTrim(ctx, f.cfg.RedisKey, 0, -batchSize-1)
	_, err := pipe.Exec(ctx)
	if err != nil {
		// Log only on transition healthy → unhealthy to avoid spam.
		if !f.redisDown.Swap(true) {
			f.log.Warnf("log flusher: redis unavailable, persistence degraded — %v", err)
		}
		return
	}
	// Log recovery once on transition unhealthy → healthy.
	if f.redisDown.Swap(false) {
		f.log.Infof("log flusher: redis recovered, persistence resumed")
	}

	raw := lrangeCmd.Val()
	if len(raw) == 0 {
		return
	}

	// Parse entries
	entries := make([]logEntry, 0, len(raw))
	for _, r := range raw {
		var e logEntry
		if err := json.Unmarshal([]byte(r), &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	if len(entries) == 0 {
		return
	}

	// Bulk insert via COPY FROM
	columns := []string{
		"level", "message", "caller",
		"operation", "entity", "entity_id", "error_text",
		"request_id", "user_id", "session_id", "ip_address",
		"extra", "created_at",
	}

	rows := make([][]any, len(entries))
	for i, e := range entries {
		ts, _ := time.Parse(time.RFC3339Nano, e.Timestamp)
		if ts.IsZero() {
			ts = time.Now().UTC()
		}
		rows[i] = []any{
			e.Level, e.Message, nullIfEmpty(e.Caller),
			nullIfEmpty(e.Operation), nullIfEmpty(e.Entity), nullIfEmpty(e.EntityID), nullIfEmpty(e.ErrorText),
			nullIfEmpty(e.RequestID), nullIfEmpty(e.UserID), nullIfEmpty(e.SessionID), nullIfEmpty(e.IPAddress),
			nullIfEmpty(e.Extra), ts,
		}
	}

	_, err = f.pool.CopyFrom(ctx,
		pgx.Identifier{"app_logs"},
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		f.log.Errorc(ctx, "log flusher: COPY FROM failed",
			"error", err,
			"batch_size", len(entries),
		)
		return
	}

	f.log.Debugf("log flusher: persisted %d entries", len(entries))
}

// ── Helpers ────────────────────────────────────────────────────────────────

func fieldToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case error:
		return val.Error()
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// parseLevel is already defined in logger.go — this variant handles persist level strings.
func parsePersistLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.WarnLevel
	}
}
