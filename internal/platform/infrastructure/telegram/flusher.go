package telegram

import (
	"context"
	"encoding/json"
	"time"

	"gct/internal/platform/infrastructure/circuitbreaker"
	"gct/internal/platform/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// TelegramFlusher periodically retries pending Telegram messages.
type TelegramFlusher struct {
	rdb    redis.Cmdable
	client *Client
	log    logger.Log
	cancel context.CancelFunc
}

// NewTelegramFlusher creates a new flusher.
func NewTelegramFlusher(rdb redis.Cmdable, client *Client, log logger.Log) *TelegramFlusher {
	return &TelegramFlusher{
		rdb:    rdb,
		client: client,
		log:    log,
	}
}

// Start begins the periodic flush loop.
func (f *TelegramFlusher) Start(ctx context.Context) {
	ctx, f.cancel = context.WithCancel(ctx)
	go f.run(ctx)
}

// Stop cancels the flusher.
func (f *TelegramFlusher) Stop() {
	if f.cancel != nil {
		f.cancel()
	}
}

func (f *TelegramFlusher) run(ctx context.Context) {
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

func (f *TelegramFlusher) flush(ctx context.Context) {
	if f.rdb == nil {
		return
	}

	// Check if circuit is closed or half-open (can try sending)
	state := f.client.telegramCB.State()
	if state == circuitbreaker.StateOpen {
		return
	}

	// Read batch from Redis
	entries, err := f.rdb.LRange(ctx, redisTelegramPendingKey, 0, 49).Result()
	if err != nil || len(entries) == 0 {
		return
	}

	// Trim the read entries
	f.rdb.LTrim(ctx, redisTelegramPendingKey, int64(len(entries)), -1)

	for _, raw := range entries {
		var entry telegramPendingEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			f.log.Warnw("TelegramFlusher: failed to unmarshal entry", "error", err)
			continue
		}

		if err := f.client.doSend(MessageType(entry.MessageType), entry.Text); err != nil {
			// Re-queue on failure
			f.rdb.RPush(ctx, redisTelegramPendingKey, raw)
			f.log.Warnw("TelegramFlusher: re-queued failed message", "error", err)
			return // stop flushing on first failure
		}
	}

	f.log.Infow("TelegramFlusher: flushed pending messages", "count", len(entries))
}
