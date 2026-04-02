package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type State string

const (
	StateClosed   State = "CLOSED"
	StateOpen     State = "OPEN"
	StateHalfOpen State = "HALF_OPEN"
)

type Config struct {
	FailureThreshold int
	Timeout          time.Duration
}

type Breaker struct {
	name     string
	cfg      Config
	mu       sync.Mutex
	state    State
	failures int
	lastFail time.Time
}

func New(name string, cfg Config) *Breaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &Breaker{
		name:  name,
		cfg:   cfg,
		state: StateClosed,
	}
}

func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == StateOpen && time.Since(b.lastFail) > b.cfg.Timeout {
		b.state = StateHalfOpen
	}
	return b.state
}

func (b *Breaker) Name() string { return b.name }

func (b *Breaker) Execute(fn func() error) error {
	b.mu.Lock()
	if b.state == StateOpen {
		if time.Since(b.lastFail) > b.cfg.Timeout {
			b.state = StateHalfOpen
		} else {
			b.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	currentState := b.state
	b.mu.Unlock()

	err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.failures++
		b.lastFail = time.Now()
		if currentState == StateHalfOpen {
			b.state = StateOpen
		} else if b.failures >= b.cfg.FailureThreshold {
			b.state = StateOpen
		}
		return err
	}

	b.failures = 0
	b.state = StateClosed
	return nil
}

// ExecuteWithFallback runs fn through the circuit breaker.
// If the circuit is open, it calls fallback instead of returning ErrCircuitOpen.
func (b *Breaker) ExecuteWithFallback(fn func() error, fallback func() error) error {
	err := b.Execute(fn)
	if errors.Is(err, ErrCircuitOpen) && fallback != nil {
		return fallback()
	}
	return err
}
