package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setup(t *testing.T) (*AuthLimiter, *miniredis.Miniredis) {
	t.Helper()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	limiter := New(client, DefaultConfig())
	return limiter, s
}

func TestIPRateLimit(t *testing.T) {
	limiter, _ := setup(t)
	ctx := context.Background()

	// 10 failures should be allowed.
	for i := 0; i < 10; i++ {
		if err := limiter.RecordFailedIP(ctx, "1.2.3.4"); err != nil {
			t.Fatalf("RecordFailedIP #%d: %v", i+1, err)
		}
	}

	// 11th attempt should be blocked.
	if err := limiter.CheckIP(ctx, "1.2.3.4"); !errors.Is(err, ErrIPRateLimited) {
		t.Fatalf("expected ErrIPRateLimited, got %v", err)
	}

	// A different IP should still pass.
	if err := limiter.CheckIP(ctx, "5.6.7.8"); err != nil {
		t.Fatalf("different IP should pass: %v", err)
	}
}

func TestIPRateLimitWindowExpiry(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	limiter := New(client, Config{
		IPLimit:     2,
		IPWindow:    10 * time.Second,
		UserLimit:   5,
		UserWindow:  15 * time.Minute,
		LockoutBase: 30 * time.Minute,
	})
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		_ = limiter.RecordFailedIP(ctx, "10.0.0.1")
	}
	if err := limiter.CheckIP(ctx, "10.0.0.1"); !errors.Is(err, ErrIPRateLimited) {
		t.Fatal("expected rate limit")
	}

	// Advance past the window.
	s.FastForward(11 * time.Second)

	if err := limiter.CheckIP(ctx, "10.0.0.1"); err != nil {
		t.Fatalf("expected pass after window expiry, got %v", err)
	}
}

func TestUserLockout(t *testing.T) {
	limiter, _ := setup(t)
	ctx := context.Background()

	// 5 failures should trigger lockout.
	for i := 0; i < 5; i++ {
		if err := limiter.RecordFailedUser(ctx, "alice@example.com"); err != nil {
			t.Fatalf("RecordFailedUser #%d: %v", i+1, err)
		}
	}

	if err := limiter.CheckUser(ctx, "alice@example.com"); !errors.Is(err, ErrAccountLocked) {
		t.Fatalf("expected ErrAccountLocked, got %v", err)
	}
}

func TestUserLockoutExpiry(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	limiter := New(client, Config{
		IPLimit:     10,
		IPWindow:    60 * time.Second,
		UserLimit:   3,
		UserWindow:  5 * time.Minute,
		LockoutBase: 10 * time.Second, // short for testing
	})
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_ = limiter.RecordFailedUser(ctx, "bob")
	}
	if err := limiter.CheckUser(ctx, "bob"); !errors.Is(err, ErrAccountLocked) {
		t.Fatal("expected lockout")
	}

	// Advance past lockout.
	s.FastForward(11 * time.Second)

	if err := limiter.CheckUser(ctx, "bob"); err != nil {
		t.Fatalf("expected pass after lockout expiry, got %v", err)
	}
}

func TestResetUserClearsState(t *testing.T) {
	limiter, _ := setup(t)
	ctx := context.Background()

	// Trigger lockout.
	for i := 0; i < 5; i++ {
		_ = limiter.RecordFailedUser(ctx, "carol")
	}
	if err := limiter.CheckUser(ctx, "carol"); !errors.Is(err, ErrAccountLocked) {
		t.Fatal("expected lockout")
	}

	// Successful login resets everything.
	if err := limiter.ResetUser(ctx, "carol"); err != nil {
		t.Fatalf("ResetUser: %v", err)
	}

	if err := limiter.CheckUser(ctx, "carol"); err != nil {
		t.Fatalf("expected pass after reset, got %v", err)
	}
}

func TestExponentialBackoff(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	cfg := Config{
		IPLimit:     10,
		IPWindow:    60 * time.Second,
		UserLimit:   2,
		UserWindow:  5 * time.Minute,
		LockoutBase: 10 * time.Second,
	}
	limiter := New(client, cfg)
	ctx := context.Background()

	// First lockout: 10s (base * 1).
	for i := 0; i < 2; i++ {
		_ = limiter.RecordFailedUser(ctx, "dave")
	}
	if err := limiter.CheckUser(ctx, "dave"); !errors.Is(err, ErrAccountLocked) {
		t.Fatal("expected first lockout")
	}

	// Advance past first lockout (10s).
	s.FastForward(11 * time.Second)
	if err := limiter.CheckUser(ctx, "dave"); err != nil {
		t.Fatalf("expected unlock after first lockout, got %v", err)
	}

	// Second lockout: 20s (base * 2).
	for i := 0; i < 2; i++ {
		_ = limiter.RecordFailedUser(ctx, "dave")
	}
	if err := limiter.CheckUser(ctx, "dave"); !errors.Is(err, ErrAccountLocked) {
		t.Fatal("expected second lockout")
	}

	// After 11s the second lockout (20s) should still be active.
	s.FastForward(11 * time.Second)
	if err := limiter.CheckUser(ctx, "dave"); !errors.Is(err, ErrAccountLocked) {
		t.Fatal("expected second lockout still active at 11s")
	}

	// After another 10s (total 21s) the second lockout should have expired.
	s.FastForward(10 * time.Second)
	if err := limiter.CheckUser(ctx, "dave"); err != nil {
		t.Fatalf("expected unlock after second lockout, got %v", err)
	}

	// Third lockout: 40s (base * 4).
	for i := 0; i < 2; i++ {
		_ = limiter.RecordFailedUser(ctx, "dave")
	}
	if err := limiter.CheckUser(ctx, "dave"); !errors.Is(err, ErrAccountLocked) {
		t.Fatal("expected third lockout")
	}

	// Verify it lasts longer than 20s.
	s.FastForward(21 * time.Second)
	if err := limiter.CheckUser(ctx, "dave"); !errors.Is(err, ErrAccountLocked) {
		t.Fatal("expected third lockout still active at 21s")
	}

	s.FastForward(20 * time.Second)
	if err := limiter.CheckUser(ctx, "dave"); err != nil {
		t.Fatalf("expected unlock after third lockout, got %v", err)
	}
}

func TestNoopLimiter(t *testing.T) {
	var noop NoopLimiter
	ctx := context.Background()

	if err := noop.CheckIP(ctx, "1.2.3.4"); err != nil {
		t.Fatalf("NoopLimiter.CheckIP should return nil, got %v", err)
	}
	if err := noop.CheckUser(ctx, "alice"); err != nil {
		t.Fatalf("NoopLimiter.CheckUser should return nil, got %v", err)
	}
	if err := noop.RecordFailedIP(ctx, "1.2.3.4"); err != nil {
		t.Fatalf("NoopLimiter.RecordFailedIP should return nil, got %v", err)
	}
	if err := noop.RecordFailedUser(ctx, "alice"); err != nil {
		t.Fatalf("NoopLimiter.RecordFailedUser should return nil, got %v", err)
	}
	if err := noop.ResetUser(ctx, "alice"); err != nil {
		t.Fatalf("NoopLimiter.ResetUser should return nil, got %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.IPLimit != 10 {
		t.Errorf("IPLimit = %d, want 10", cfg.IPLimit)
	}
	if cfg.IPWindow != 60*time.Second {
		t.Errorf("IPWindow = %v, want 60s", cfg.IPWindow)
	}
	if cfg.UserLimit != 5 {
		t.Errorf("UserLimit = %d, want 5", cfg.UserLimit)
	}
	if cfg.UserWindow != 15*time.Minute {
		t.Errorf("UserWindow = %v, want 15m", cfg.UserWindow)
	}
	if cfg.LockoutBase != 30*time.Minute {
		t.Errorf("LockoutBase = %v, want 30m", cfg.LockoutBase)
	}
}
