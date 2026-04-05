package reqlog

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// DefaultRedisKey is the Redis list key used to buffer incoming request log
// entries before they are flushed to PostgreSQL.
const DefaultRedisKey = "http_request_logs:buffer"

const (
	defaultChanSize    = 2048
	defaultBatchSize   = 100
	defaultPushTimeout = 500 * time.Millisecond
	defaultFlushEvery  = 100 * time.Millisecond
)

// RedisSink is a fire-and-forget async sink. Push() places the entry onto an
// in-memory channel and returns immediately — the hot request path is never
// blocked by Redis. A background worker batches entries and LPushes them to
// Redis via a single pipeline, minimising round-trips.
//
// If the in-memory buffer is full (Redis is slow or down), new entries are
// dropped silently. This is the correct behaviour: we prefer to drop telemetry
// rather than degrade user-facing latency.
type RedisSink struct {
	rdb    redis.Cmdable
	key    string
	ch     chan Entry
	stopCh chan struct{}
	doneCh chan struct{}
	once   sync.Once

	// metrics (read via Stats)
	dropped uint64
	dropMu  sync.Mutex
}

// NewRedisSink constructs and starts an async Redis sink. Call Stop to flush
// the buffer on shutdown.
func NewRedisSink(rdb redis.Cmdable, key string) *RedisSink {
	if key == "" {
		key = DefaultRedisKey
	}
	s := &RedisSink{
		rdb:    rdb,
		key:    key,
		ch:     make(chan Entry, defaultChanSize),
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
	go s.run()
	return s
}

// Push enqueues an entry. Non-blocking: if the in-memory buffer is full, the
// entry is dropped and the dropped counter is incremented.
func (s *RedisSink) Push(e Entry) {
	if s == nil || s.rdb == nil {
		return
	}
	select {
	case s.ch <- e:
	default:
		s.dropMu.Lock()
		s.dropped++
		s.dropMu.Unlock()
		incDropped()
	}
}

// Stop signals the worker to drain the buffer and exit. Safe to call multiple
// times. Blocks until drain completes.
func (s *RedisSink) Stop() {
	s.once.Do(func() {
		close(s.stopCh)
	})
	<-s.doneCh
}

// Dropped returns the total number of entries dropped due to buffer pressure
// since construction.
func (s *RedisSink) Dropped() uint64 {
	s.dropMu.Lock()
	defer s.dropMu.Unlock()
	return s.dropped
}

func (s *RedisSink) run() {
	defer close(s.doneCh)

	batch := make([]Entry, 0, defaultBatchSize)
	ticker := time.NewTicker(defaultFlushEvery)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		s.writeBatch(batch)
		batch = batch[:0]
	}

	for {
		select {
		case e := <-s.ch:
			batch = append(batch, e)
			if len(batch) >= defaultBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-s.stopCh:
			// Drain remaining entries then exit.
			for {
				select {
				case e := <-s.ch:
					batch = append(batch, e)
					if len(batch) >= defaultBatchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

func (s *RedisSink) writeBatch(batch []Entry) {
	defer func() {
		// Guard the worker from unexpected panics (nil map, etc.) so the
		// goroutine survives transient Redis or encoding weirdness.
		_ = recover()
	}()

	// Runs on the background worker goroutine started in NewRedisSink — no
	// caller context exists, and the writer outlives every individual request.
	ctx, cancel := context.WithTimeout(context.Background(), defaultPushTimeout)
	defer cancel()

	pipe := s.rdb.Pipeline()
	for i := range batch {
		data, err := json.Marshal(&batch[i])
		if err != nil {
			continue
		}
		pipe.LPush(ctx, s.key, data)
	}
	_, _ = pipe.Exec(ctx)
}
