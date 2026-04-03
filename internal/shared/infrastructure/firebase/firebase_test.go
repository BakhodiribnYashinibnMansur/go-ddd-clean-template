package firebase

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/circuitbreaker"
	"gct/internal/shared/infrastructure/logger"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// ---------------------------------------------------------------------------
// selectCBAndClient
// ---------------------------------------------------------------------------

func newTestFirebase() *Firebase {
	log := logger.Noop()
	mobileCB := circuitbreaker.New("mobile-test", circuitbreaker.Config{
		FailureThreshold: 5,
		Timeout:          30 * time.Second,
	})
	webCB := circuitbreaker.New("web-test", circuitbreaker.Config{
		FailureThreshold: 5,
		Timeout:          30 * time.Second,
	})
	return &Firebase{
		logger:   log,
		mobileCB: mobileCB,
		webCB:    webCB,
		// MobileClient and WebClient are nil — we only test the switch logic,
		// not actual FCM sends.
	}
}

func TestSelectCBAndClient_Client(t *testing.T) {
	fb := newTestFirebase()
	cb, client, ok := fb.selectCBAndClient(FCM_TYPE_CLIENT)
	if !ok {
		t.Fatal("expected ok=true for FCM_TYPE_CLIENT")
	}
	if cb != fb.mobileCB {
		t.Error("expected mobileCB for FCM_TYPE_CLIENT")
	}
	if client != fb.MobileClient {
		t.Error("expected MobileClient for FCM_TYPE_CLIENT")
	}
}

func TestSelectCBAndClient_Admin(t *testing.T) {
	fb := newTestFirebase()
	cb, client, ok := fb.selectCBAndClient(FCM_TYPE_ADMIN)
	if !ok {
		t.Fatal("expected ok=true for FCM_TYPE_ADMIN")
	}
	if cb != fb.webCB {
		t.Error("expected webCB for FCM_TYPE_ADMIN")
	}
	if client != fb.WebClient {
		t.Error("expected WebClient for FCM_TYPE_ADMIN")
	}
}

func TestSelectCBAndClient_Craftsman(t *testing.T) {
	fb := newTestFirebase()
	cb, client, ok := fb.selectCBAndClient(FCM_TYPE_CRAFTSMAN)
	if !ok {
		t.Fatal("expected ok=true for FCM_TYPE_CRAFTSMAN")
	}
	if cb != fb.webCB {
		t.Error("expected webCB for FCM_TYPE_CRAFTSMAN")
	}
	if client != fb.WebClient {
		t.Error("expected WebClient for FCM_TYPE_CRAFTSMAN")
	}
}

func TestSelectCBAndClient_Unknown(t *testing.T) {
	fb := newTestFirebase()
	cb, client, ok := fb.selectCBAndClient("UNKNOWN")
	if ok {
		t.Fatal("expected ok=false for unknown FCM type")
	}
	if cb != nil {
		t.Error("expected nil breaker for unknown FCM type")
	}
	if client != nil {
		t.Error("expected nil client for unknown FCM type")
	}
}

// ---------------------------------------------------------------------------
// FCMFlusher — construction and lifecycle
// ---------------------------------------------------------------------------

func TestNewFCMFlusher(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	sendAlert := func(msg string) error { return nil }
	f := NewFCMFlusher(rdb, sendAlert, logger.Noop())
	if f == nil {
		t.Fatal("expected non-nil FCMFlusher")
	}
}

func TestFCMFlusher_Stop_NilCancel(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	f := NewFCMFlusher(rdb, nil, logger.Noop())
	// Stop without Start — must not panic.
	f.Stop()
}

func TestFCMFlusher_StartAndStop(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	f := NewFCMFlusher(rdb, func(string) error { return nil }, logger.Noop())
	ctx := context.Background()
	f.Start(ctx)

	// Give the goroutine a moment to start.
	time.Sleep(10 * time.Millisecond)

	// Stop must not panic and must cancel the internal context.
	f.Stop()
}

// ---------------------------------------------------------------------------
// FCMFlusher.flush
// ---------------------------------------------------------------------------

func TestFCMFlusher_Flush_EmptyRedis(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	var alerts []string
	sendAlert := func(msg string) error { alerts = append(alerts, msg); return nil }

	f := NewFCMFlusher(rdb, sendAlert, logger.Noop())
	f.flush(context.Background())

	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestFCMFlusher_Flush_WithEntries(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	// Push one fallback entry into Redis.
	entry := fcmFallbackEntry{
		Token:     "tok-1",
		Title:     "Test Title",
		Body:      "Test Body",
		Data:      map[string]string{"key": "val"},
		FCMType:   FCM_TYPE_CLIENT,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	raw, _ := json.Marshal(entry)
	if err := rdb.LPush(context.Background(), redisFCMFallbackKey, raw).Err(); err != nil {
		t.Fatal(err)
	}

	var alerts []string
	sendAlert := func(msg string) error { alerts = append(alerts, msg); return nil }

	f := NewFCMFlusher(rdb, sendAlert, logger.Noop())
	f.flush(context.Background())

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}

	// The Redis list should be cleared after flush.
	remaining, _ := rdb.LLen(context.Background(), redisFCMFallbackKey).Result()
	if remaining != 0 {
		t.Fatalf("expected Redis list to be empty after flush, got %d entries", remaining)
	}
}

func TestFCMFlusher_Flush_RedisError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	var alerts []string
	sendAlert := func(msg string) error { alerts = append(alerts, msg); return nil }

	f := NewFCMFlusher(rdb, sendAlert, logger.Noop())

	// Close miniredis to simulate Redis being unavailable.
	mr.Close()

	f.flush(context.Background())

	// No alert should be sent — the counter should be incremented instead.
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts when Redis is down, got %d", len(alerts))
	}
	if got := f.inMemCounter.Load(); got != 1 {
		t.Fatalf("expected inMemCounter=1, got %d", got)
	}
}

func TestFCMFlusher_Flush_GroupsEntries(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	ctx := context.Background()

	// Push three entries: two with same title+body, one different.
	entries := []fcmFallbackEntry{
		{Token: "tok-A", Title: "Same", Body: "Body", FCMType: FCM_TYPE_CLIENT, Timestamp: time.Now().UTC().Format(time.RFC3339)},
		{Token: "tok-B", Title: "Same", Body: "Body", FCMType: FCM_TYPE_CLIENT, Timestamp: time.Now().UTC().Format(time.RFC3339)},
		{Token: "tok-C", Title: "Different", Body: "Other", FCMType: FCM_TYPE_ADMIN, Timestamp: time.Now().UTC().Format(time.RFC3339)},
	}
	for _, e := range entries {
		raw, _ := json.Marshal(e)
		if err := rdb.LPush(ctx, redisFCMFallbackKey, raw).Err(); err != nil {
			t.Fatal(err)
		}
	}

	var alerts []string
	sendAlert := func(msg string) error { alerts = append(alerts, msg); return nil }

	f := NewFCMFlusher(rdb, sendAlert, logger.Noop())
	f.flush(ctx)

	// Two groups: ("Same","Body") and ("Different","Other").
	if len(alerts) != 2 {
		t.Fatalf("expected 2 grouped alerts, got %d", len(alerts))
	}
}

// ---------------------------------------------------------------------------
// Content constants
// ---------------------------------------------------------------------------

func TestContent_StructFields(t *testing.T) {
	c := Content{Title: "T", Body: "B"}
	if c.Title != "T" {
		t.Error("Title field mismatch")
	}
	if c.Body != "B" {
		t.Error("Body field mismatch")
	}
}

func TestContent_OrderNewMap(t *testing.T) {
	if len(OrderNewMap) != 3 {
		t.Fatalf("expected 3 languages in OrderNewMap, got %d", len(OrderNewMap))
	}
	for _, lang := range []string{En, Ru, Uz} {
		c, ok := OrderNewMap[lang]
		if !ok {
			t.Errorf("OrderNewMap missing language %s", lang)
			continue
		}
		if c.Title == "" {
			t.Errorf("OrderNewMap[%s].Title is empty", lang)
		}
		if c.Body == "" {
			t.Errorf("OrderNewMap[%s].Body is empty", lang)
		}
	}
}

func TestContent_LanguageConstants(t *testing.T) {
	if En != "en" {
		t.Errorf("expected En=%q, got %q", "en", En)
	}
	if Ru != "ru" {
		t.Errorf("expected Ru=%q, got %q", "ru", Ru)
	}
	if Uz != "uz" {
		t.Errorf("expected Uz=%q, got %q", "uz", Uz)
	}
}
