package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"gct/internal/shared/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// FCMFlusher periodically reads from the FCM fallback buffer and sends grouped Telegram alerts.
type FCMFlusher struct {
	rdb          redis.Cmdable
	sendAlert    func(string) error
	log          logger.Log
	cancel       context.CancelFunc
	inMemCounter atomic.Int64
}

// NewFCMFlusher creates a new FCM flusher.
// sendAlert is typically telegramClient.SendInfo.
func NewFCMFlusher(rdb redis.Cmdable, sendAlert func(string) error, log logger.Log) *FCMFlusher {
	return &FCMFlusher{
		rdb:       rdb,
		sendAlert: sendAlert,
		log:       log,
	}
}

// Start begins the periodic flush loop.
func (f *FCMFlusher) Start(ctx context.Context) {
	ctx, f.cancel = context.WithCancel(ctx)
	go f.run(ctx)
}

// Stop cancels the flusher.
func (f *FCMFlusher) Stop() {
	if f.cancel != nil {
		f.cancel()
	}
}

func (f *FCMFlusher) run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
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

type groupKey struct {
	Title string
	Body  string
}

func (f *FCMFlusher) flush(ctx context.Context) {
	if f.rdb == nil || f.sendAlert == nil {
		return
	}

	entries, err := f.rdb.LRange(ctx, redisFCMFallbackKey, 0, -1).Result()
	if err != nil {
		// Redis down - count locally
		f.inMemCounter.Add(1)
		return
	}

	// Check recovery from Redis outage
	if dropped := f.inMemCounter.Swap(0); dropped > 0 {
		_ = f.sendAlert(fmt.Sprintf("FCM fallback: Redis recovered. ~%d flush cycles were missed during outage.", dropped))
	}

	if len(entries) == 0 {
		return
	}

	// Clear the list
	f.rdb.Del(ctx, redisFCMFallbackKey)

	// Group by title+body
	groups := make(map[groupKey][]fcmFallbackEntry)
	for _, raw := range entries {
		var entry fcmFallbackEntry
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			continue
		}
		key := groupKey{Title: entry.Title, Body: entry.Body}
		groups[key] = append(groups[key], entry)
	}

	// Send grouped alerts
	for key, group := range groups {
		tokens := make([]string, 0, len(group))
		for _, e := range group {
			tokens = append(tokens, e.Token)
		}

		text := fmt.Sprintf("FCM Delivery Failure Alert\nTitle: %s\nBody: %s\nAffected tokens: %d\nTokens: %s\nFCM Type: %s",
			key.Title, key.Body, len(tokens),
			strings.Join(tokens, ", "),
			group[0].FCMType)

		if err := f.sendAlert(text); err != nil {
			f.log.Warnw("FCMFlusher: failed to send Telegram alert", "error", err)
		}
	}

	f.log.Infow("FCMFlusher: processed fallback entries", "count", len(entries))
}
