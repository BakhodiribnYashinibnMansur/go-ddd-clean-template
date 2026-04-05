package reqlog

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// blockingRedis pretends to be a redis.Cmdable but blocks on LPush until
// released, which lets us verify buffer-full drop behaviour.
type blockingRedis struct {
	redis.Cmdable
	gate  chan struct{}
	count atomic.Int64
}

func (b *blockingRedis) Pipeline() redis.Pipeliner {
	return &blockingPipeline{parent: b}
}

type blockingPipeline struct {
	redis.Pipeliner
	parent *blockingRedis
	cmds   int
}

func (p *blockingPipeline) LPush(_ context.Context, _ string, _ ...any) *redis.IntCmd {
	p.cmds++
	return redis.NewIntCmd(context.Background())
}

func (p *blockingPipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	select {
	case <-p.parent.gate:
	case <-ctx.Done():
	}
	p.parent.count.Add(int64(p.cmds))
	return nil, nil
}

func TestRedisSink_DropsOnBufferFull(t *testing.T) {
	br := &blockingRedis{gate: make(chan struct{})}
	s := &RedisSink{
		rdb:    br,
		key:    "x",
		ch:     make(chan Entry, 4), // tiny buffer
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
	go s.run()

	// Fill buffer + overflow
	for i := 0; i < 100; i++ {
		s.Push(Entry{Method: "GET"})
	}
	// Buffer was 4 — expect lots of drops.
	if s.Dropped() < 50 {
		t.Fatalf("expected drops (buffer is tiny + worker is blocked), got %d", s.Dropped())
	}

	close(br.gate)
	// Let worker drain.
	time.Sleep(50 * time.Millisecond)
	s.once.Do(func() { close(s.stopCh) })
	<-s.doneCh
}

func TestRedisSink_PushZeroAllocOnStoppedSink(t *testing.T) {
	var s *RedisSink
	s.Push(Entry{}) // must be no-op, not panic
}

func TestRedisSink_Stop_DrainsBuffer(t *testing.T) {
	// A fake redis that records what was LPush'd.
	rec := &recordingRedis{}
	s := NewRedisSink(rec, "k")
	for i := 0; i < 10; i++ {
		s.Push(Entry{Method: "POST"})
	}
	s.Stop()
	if rec.count.Load() < 10 {
		t.Fatalf("drain incomplete: pushed=%d got=%d", 10, rec.count.Load())
	}
}

type recordingRedis struct {
	redis.Cmdable
	mu    sync.Mutex
	count atomic.Int64
}

func (r *recordingRedis) Pipeline() redis.Pipeliner { return &recordingPipeline{parent: r} }

type recordingPipeline struct {
	redis.Pipeliner
	parent *recordingRedis
	cmds   int
}

func (p *recordingPipeline) LPush(_ context.Context, _ string, _ ...any) *redis.IntCmd {
	p.cmds++
	return redis.NewIntCmd(context.Background())
}
func (p *recordingPipeline) Exec(_ context.Context) ([]redis.Cmder, error) {
	p.parent.count.Add(int64(p.cmds))
	return nil, nil
}
