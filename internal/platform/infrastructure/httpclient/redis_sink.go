package httpclient

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// DefaultRedisKey is the Redis list key used to buffer external api log
// entries before they are flushed to PostgreSQL.
const DefaultRedisKey = "external_api_logs:buffer"

const (
	defaultChanSize    = 1024
	defaultBatchSize   = 100
	defaultPushTimeout = 500 * time.Millisecond
	defaultFlushEvery  = 100 * time.Millisecond
)

// RedisSink is a fire-and-forget async sink. Push() enqueues onto an
// in-memory channel and returns instantly; a worker goroutine batches entries
// and LPushes them to Redis via a single pipeline.
//
// If the in-memory buffer is full (Redis slow/down), entries are dropped. The
// caller's request latency is never impacted by the logging path.
type RedisSink struct {
	rdb    redis.Cmdable
	key    string
	ch     chan Entry
	stopCh chan struct{}
	doneCh chan struct{}
	once   sync.Once

	dropped uint64
	dropMu  sync.Mutex
}

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

// Push enqueues an entry non-blocking. If the buffer is full the entry is
// dropped and Dropped() is incremented.
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

// Stop drains the in-memory buffer then returns. Safe to call multiple times.
func (s *RedisSink) Stop() {
	s.once.Do(func() { close(s.stopCh) })
	<-s.doneCh
}

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
	defer func() { _ = recover() }()

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
