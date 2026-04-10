package circuitbreaker_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"gct/internal/kernel/infrastructure/circuitbreaker"

	"pgregory.net/rapid"
)

var errTest = errors.New("test error")

func TestBreaker_Property_InitialState(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		threshold := rapid.IntRange(1, 10).Draw(t, "threshold")
		b := circuitbreaker.New("test", circuitbreaker.Config{
			FailureThreshold: threshold,
			Timeout:          time.Millisecond,
		})
		if b.State() != circuitbreaker.StateClosed {
			t.Fatalf("initial state = %q, want CLOSED", b.State())
		}
	})
}

func TestBreaker_Property_ThresholdTransition(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		threshold := rapid.IntRange(1, 10).Draw(t, "threshold")
		b := circuitbreaker.New("test", circuitbreaker.Config{
			FailureThreshold: threshold,
			Timeout:          time.Second, // long timeout so it stays open
		})

		// exactly threshold failures should open the circuit
		for i := 0; i < threshold; i++ {
			_ = b.Execute(func() error { return errTest })
		}
		if b.State() != circuitbreaker.StateOpen {
			t.Fatalf("state after %d failures = %q, want OPEN", threshold, b.State())
		}
	})
}

func TestBreaker_Property_OpenRejects(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		threshold := rapid.IntRange(1, 5).Draw(t, "threshold")
		b := circuitbreaker.New("test", circuitbreaker.Config{
			FailureThreshold: threshold,
			Timeout:          time.Second,
		})

		// open the circuit
		for i := 0; i < threshold; i++ {
			_ = b.Execute(func() error { return errTest })
		}

		// subsequent calls should be rejected
		nCalls := rapid.IntRange(1, 10).Draw(t, "nCalls")
		for i := 0; i < nCalls; i++ {
			err := b.Execute(func() error { return nil })
			if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
				t.Fatalf("call %d in OPEN state: err = %v, want ErrCircuitOpen", i, err)
			}
		}
	})
}

func TestBreaker_Property_FSMValidity(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		threshold := rapid.IntRange(1, 5).Draw(t, "threshold")
		b := circuitbreaker.New("test", circuitbreaker.Config{
			FailureThreshold: threshold,
			Timeout:          time.Millisecond,
		})

		validStates := map[circuitbreaker.State]bool{
			circuitbreaker.StateClosed:   true,
			circuitbreaker.StateOpen:     true,
			circuitbreaker.StateHalfOpen: true,
		}

		nActions := rapid.IntRange(1, 30).Draw(t, "nActions")
		for i := 0; i < nActions; i++ {
			action := rapid.SampledFrom([]string{"success", "failure", "wait"}).Draw(t, fmt.Sprintf("action%d", i))
			switch action {
			case "success":
				_ = b.Execute(func() error { return nil })
			case "failure":
				_ = b.Execute(func() error { return errTest })
			case "wait":
				time.Sleep(2 * time.Millisecond)
			}
			state := b.State()
			if !validStates[state] {
				t.Fatalf("invalid state %q after action %q", state, action)
			}
		}
	})
}

func TestBreaker_Property_SuccessResetsToClosedFromHalfOpen(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		threshold := rapid.IntRange(1, 5).Draw(t, "threshold")
		b := circuitbreaker.New("test", circuitbreaker.Config{
			FailureThreshold: threshold,
			Timeout:          time.Millisecond,
		})

		// open the circuit
		for i := 0; i < threshold; i++ {
			_ = b.Execute(func() error { return errTest })
		}

		// wait for timeout to transition to half-open
		time.Sleep(2 * time.Millisecond)

		// success should close the circuit
		err := b.Execute(func() error { return nil })
		if err != nil {
			t.Fatalf("success in half-open failed: %v", err)
		}
		if b.State() != circuitbreaker.StateClosed {
			t.Fatalf("state after success in half-open = %q, want CLOSED", b.State())
		}
	})
}
