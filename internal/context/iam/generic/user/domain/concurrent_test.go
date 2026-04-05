package domain_test

import (
	"sync"
	"testing"

	domain "gct/internal/context/iam/generic/user/domain"
)

func TestUser_ConcurrentAddSession(t *testing.T) {
	t.Parallel()

	// Aggregate roots are NOT required to be goroutine-safe: the standard
	// DDD contract is that a single transaction owns the aggregate for the
	// duration of a use case. This test verifies the application-level
	// pattern of serializing writes via an external mutex, which is what
	// real callers must do.
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u, _ := domain.NewUser(phone, pw)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = make(chan error, 20)
	)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			_, err := u.AddSession(domain.DeviceMobile, "1.1.1.1", "Agent", "gct-client")
			mu.Unlock()
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	var successes, maxReached int
	for err := range results {
		if err == nil {
			successes++
		} else {
			maxReached++
		}
	}

	// With serialized access, exactly maxSessions (50 by default) sessions
	// can be added, but we only attempt 20 — so all 20 must succeed.
	if successes != 20 || maxReached != 0 {
		t.Fatalf("expected 20 successes and 0 errors, got successes=%d errors=%d", successes, maxReached)
	}
}

func TestUser_ConcurrentEvents(t *testing.T) {
	t.Parallel()

	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u, _ := domain.NewUser(phone, pw)
	u.ClearEvents()
	u.Approve()

	// Concurrent reads of events: callers must serialize access to the
	// aggregate; this test verifies the guarded pattern works.
	var (
		wg sync.WaitGroup
		mu sync.RWMutex
	)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.RLock()
			events := u.Events()
			_ = len(events)
			mu.RUnlock()
		}()
	}
	wg.Wait()
}

func TestNewPasswordFromRaw_Concurrent(t *testing.T) {
	t.Parallel()

	// bcrypt should be safe for concurrent use
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pw, err := domain.NewPasswordFromRaw("ConcurrentP@ss1")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if err := pw.Compare("ConcurrentP@ss1"); err != nil {
				t.Errorf("compare failed: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestSignInService_Concurrent(t *testing.T) {
	t.Parallel()

	svc := &domain.SignInService{}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			phone := mustPhone(t, "+998901234567")
			pw := mustPassword(t, "SecureP@ss1")
			u, _ := domain.NewUser(phone, pw)
			u.Approve()
			_, _ = svc.SignIn(u, "SecureP@ss1", domain.DeviceDesktop, "10.0.0.1", "Agent", "gct-client")
		}()
	}
	wg.Wait()
}
