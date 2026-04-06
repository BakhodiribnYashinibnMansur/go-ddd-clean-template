package audit

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gct/internal/kernel/infrastructure/logger"
)

const (
	channelSize   = 1000
	batchSize     = 50
	flushInterval = 500 * time.Millisecond
)

// PgLogger is a Logger backed by PostgreSQL. It buffers entries in a channel
// and batch-INSERTs them in a background goroutine so callers are never blocked.
type PgLogger struct {
	pool    *pgxpool.Pool
	ch      chan Entry
	dropped atomic.Int64
	l       logger.Log
	done    chan struct{}
}

// NewPgLogger creates a new PgLogger. Call Start to launch the background writer
// and Close to drain and stop it.
func NewPgLogger(pool *pgxpool.Pool, l logger.Log) *PgLogger {
	return &PgLogger{
		pool: pool,
		ch:   make(chan Entry, channelSize),
		l:    l,
		done: make(chan struct{}),
	}
}

// Start launches the background goroutine that drains the channel and writes
// batches to PostgreSQL. It returns immediately.
func (p *PgLogger) Start(ctx context.Context) {
	go p.run(ctx)
}

// Log enqueues an entry for asynchronous persistence. If the internal buffer is
// full the entry is silently dropped and the dropped counter is incremented.
func (p *PgLogger) Log(_ context.Context, entry Entry) {
	select {
	case p.ch <- entry:
	default:
		n := p.dropped.Add(1)
		if n%100 == 1 {
			p.l.Warnw("audit entry dropped (channel full)", "total_dropped", n)
		}
	}
}

// Dropped returns the number of entries dropped because the channel was full.
func (p *PgLogger) Dropped() int64 {
	return p.dropped.Load()
}

// Close signals the background goroutine to drain remaining entries and stop.
// It blocks until draining is complete.
func (p *PgLogger) Close() {
	close(p.ch)
	<-p.done
}

// run is the background loop. It collects entries up to batchSize or
// flushInterval and writes them in a single pgx.Batch round-trip.
func (p *PgLogger) run(ctx context.Context) {
	defer close(p.done)

	buf := make([]Entry, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case e, ok := <-p.ch:
			if !ok {
				// Channel closed — drain remaining and exit.
				p.flush(ctx, buf)
				return
			}
			buf = append(buf, e)
			if len(buf) >= batchSize {
				p.flush(ctx, buf)
				buf = buf[:0]
			}

		case <-ticker.C:
			if len(buf) > 0 {
				p.flush(ctx, buf)
				buf = buf[:0]
			}

		case <-ctx.Done():
			// Context cancelled — drain what we can.
			for {
				select {
				case e, ok := <-p.ch:
					if !ok {
						p.flush(ctx, buf)
						return
					}
					buf = append(buf, e)
				default:
					p.flush(ctx, buf)
					return
				}
			}
		}
	}
}

const insertSQL = `INSERT INTO security_audit_log
	(event, integration_name, user_id, session_id, ip_address, user_agent, metadata, created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())`

func (p *PgLogger) flush(ctx context.Context, entries []Entry) {
	if len(entries) == 0 {
		return
	}

	// Use a background context for the actual write so we can still flush
	// even if the request context is already cancelled.
	writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	batch := &pgx.Batch{}
	for i := range entries {
		e := &entries[i]
		meta, err := json.Marshal(e.Metadata)
		if err != nil {
			meta = []byte("{}")
		}
		batch.Queue(insertSQL, e.Event, nilStr(e.IntegrationName), e.UserID, e.SessionID, nilStr(e.IPAddress), nilStr(e.UserAgent), meta)
	}

	br := p.pool.SendBatch(writeCtx, batch)
	defer func() {
		if err := br.Close(); err != nil {
			p.l.Warnw("audit batch close error", "err", err)
		}
	}()

	for range entries {
		if _, err := br.Exec(); err != nil {
			p.l.Errorw("audit insert failed", "err", err)
		}
	}

	_ = ctx // keep linter happy about named param
}

// nilStr returns nil for empty strings so PostgreSQL stores NULL instead of ''.
func nilStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
