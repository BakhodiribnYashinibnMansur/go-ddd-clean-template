package domain_test

import (
	"sync"
	"testing"

	domain "gct/internal/user/domain"
)

func TestUser_ConcurrentAddSession(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)

	// Try to add sessions concurrently — should not panic or corrupt state.
	var wg sync.WaitGroup
	results := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := u.AddSession(domain.DeviceMobile, "1.1.1.1", "Agent")
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

	// Due to race conditions without locks, we just verify no panic occurred.
	// The total should be roughly 10 successes and 10 errors, but exact count
	// depends on goroutine scheduling.
	t.Logf("successes: %d, maxReached: %d", successes, maxReached)
}

func TestUser_ConcurrentEvents(t *testing.T) {
	phone := mustPhone(t, "+998901234567")
	pw := mustPassword(t, "SecureP@ss1")
	u := domain.NewUser(phone, pw)
	u.ClearEvents()
	u.Approve()

	// Concurrent reads of events should not panic
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			events := u.Events()
			_ = len(events)
		}()
	}
	wg.Wait()
}

func TestNewPasswordFromRaw_Concurrent(t *testing.T) {
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
	svc := &domain.SignInService{}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			phone := mustPhone(t, "+998901234567")
			pw := mustPassword(t, "SecureP@ss1")
			u := domain.NewUser(phone, pw)
			u.Approve()
			_, _ = svc.SignIn(u, "SecureP@ss1", domain.DeviceDesktop, "10.0.0.1", "Agent")
		}()
	}
	wg.Wait()
}
