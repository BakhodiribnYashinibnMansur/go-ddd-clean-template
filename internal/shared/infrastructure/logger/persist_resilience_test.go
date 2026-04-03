package logger

import (
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap/zapcore"
)

// TestRedisSink_Write_ContinuesOnRedisFailure verifies that Write returns nil
// (not an error) when Redis is completely down, ensuring graceful degradation.
func TestRedisSink_Write_ContinuesOnRedisFailure(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	sink := &RedisSink{
		rdb:         rdb,
		key:         "test:resilience:logs",
		minLevel:    zapcore.WarnLevel,
		pushTimeout: 200 * time.Millisecond,
	}

	// Shut down Redis before writing.
	mr.Close()

	ent := zapcore.Entry{
		Level:   zapcore.ErrorLevel,
		Message: "this must not fail",
		Time:    time.Now(),
	}

	err := sink.Write(ent, nil)
	if err != nil {
		t.Fatalf("Write() returned error when Redis is down: %v; expected nil (graceful degradation)", err)
	}
}

// TestRedisSink_Write_RecoverAfterRedisRestart verifies that after Redis goes
// down and comes back up, the sink successfully pushes entries again.
func TestRedisSink_Write_RecoverAfterRedisRestart(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	sink := &RedisSink{
		rdb:         rdb,
		key:         "test:resilience:logs",
		minLevel:    zapcore.WarnLevel,
		pushTimeout: 200 * time.Millisecond,
	}

	// Phase 1: Redis is down — write should silently succeed (return nil).
	mr.Close()

	ent := zapcore.Entry{
		Level:   zapcore.ErrorLevel,
		Message: "during outage",
		Time:    time.Now(),
	}
	if err := sink.Write(ent, nil); err != nil {
		t.Fatalf("Write() during outage returned error: %v", err)
	}

	// Phase 2: Restart miniredis on the same address.
	if err := mr.Restart(); err != nil {
		t.Fatalf("failed to restart miniredis: %v", err)
	}

	ent2 := zapcore.Entry{
		Level:   zapcore.ErrorLevel,
		Message: "after recovery",
		Time:    time.Now(),
	}
	if err := sink.Write(ent2, nil); err != nil {
		t.Fatalf("Write() after recovery returned error: %v", err)
	}

	items, err := mr.List(sink.key)
	if err != nil {
		t.Fatalf("failed to read list from miniredis: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item after recovery, got %d", len(items))
	}
}

// TestRedisSink_HighVolume writes 1000 entries rapidly to ensure no panics or
// data races occur under concurrent load.
func TestRedisSink_HighVolume(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	sink := &RedisSink{
		rdb:         rdb,
		key:         "test:highvolume:logs",
		minLevel:    zapcore.DebugLevel,
		pushTimeout: time.Second,
	}

	const numEntries = 1000
	var wg sync.WaitGroup
	wg.Add(numEntries)

	for i := 0; i < numEntries; i++ {
		go func(n int) {
			defer wg.Done()
			ent := zapcore.Entry{
				Level:   zapcore.ErrorLevel,
				Message: "high-volume entry",
				Time:    time.Now(),
			}
			if err := sink.Write(ent, nil); err != nil {
				t.Errorf("Write() #%d returned error: %v", n, err)
			}
		}(i)
	}

	wg.Wait()

	items, err := mr.List(sink.key)
	if err != nil {
		t.Fatalf("failed to read list: %v", err)
	}
	if len(items) != numEntries {
		t.Errorf("expected %d items in Redis, got %d", numEntries, len(items))
	}
}
