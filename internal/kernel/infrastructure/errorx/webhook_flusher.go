package errorx

import (
	"context"
	"time"

	"gct/internal/kernel/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// WebhookFlusher periodically reads from the webhook pending queue and sends via HTTP.
type WebhookFlusher struct {
	reporter *WebhookReporter
	rdb      redis.Cmdable
	log      logger.Log
	cancel   context.CancelFunc
}

// NewWebhookFlusher creates a new webhook flusher.
func NewWebhookFlusher(reporter *WebhookReporter, rdb redis.Cmdable, log logger.Log) *WebhookFlusher {
	return &WebhookFlusher{
		reporter: reporter,
		rdb:      rdb,
		log:      log,
	}
}

// Start begins the periodic flush loop.
func (f *WebhookFlusher) Start(ctx context.Context) {
	ctx, f.cancel = context.WithCancel(ctx)
	go f.run(ctx)
}

// Stop cancels the flusher.
func (f *WebhookFlusher) Stop() {
	if f.cancel != nil {
		f.cancel()
	}
}

func (f *WebhookFlusher) run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f.flush(ctx)
		}
	}
}

func (f *WebhookFlusher) flush(ctx context.Context) {
	if f.rdb == nil {
		return
	}

	entries, err := f.rdb.LRange(ctx, redisWebhookPendingKey, 0, 49).Result()
	if err != nil || len(entries) == 0 {
		return
	}

	f.rdb.LTrim(ctx, redisWebhookPendingKey, int64(len(entries)), -1)

	for _, raw := range entries {
		err := f.reporter.webhookCB.Execute(func() error {
			return f.reporter.doPost([]byte(raw))
		})
		if err != nil {
			// Re-queue on failure
			f.rdb.RPush(ctx, redisWebhookPendingKey, raw)
			f.log.Warnw("WebhookFlusher: re-queued failed entry", "error", err)
			return // stop on first failure
		}
	}

	f.log.Infow("WebhookFlusher: flushed pending webhooks", "count", len(entries))
}
