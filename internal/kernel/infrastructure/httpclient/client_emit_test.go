package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// captureSink is a test Sink that records every Entry pushed to it.
type captureSink struct {
	mu      sync.Mutex
	entries []Entry
}

func (s *captureSink) Push(e Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, e)
}

func (s *captureSink) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

func (s *captureSink) last() Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.entries[len(s.entries)-1]
}

// newTestClient wires a Client to a fresh captureSink with the given knobs.
func newTestClient(t *testing.T, slow time.Duration, sampleRate float64) (*Client, *captureSink) {
	t.Helper()
	sink := &captureSink{}
	c := New(Options{
		APIName:           "test",
		Timeout:           5 * time.Second,
		MaxBodyBytes:      1024,
		SlowThreshold:     slow,
		SuccessSampleRate: sampleRate,
	}, sink, nil)
	return c, sink
}

func TestDo_TransportError_Emits(t *testing.T) {
	c, sink := newTestClient(t, time.Second, 0)

	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:1/unreachable", nil)
	_, _, err := c.Do(context.Background(), req, "op")
	if err == nil {
		t.Fatal("expected transport error")
	}
	if sink.count() != 1 {
		t.Fatalf("expected 1 emission on transport error, got %d", sink.count())
	}
	if got := sink.last().Outcome; got != OutcomeError {
		t.Errorf("outcome = %q, want %q", got, OutcomeError)
	}
}

func TestDo_Status500_Emits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"err":"boom"}`))
	}))
	defer srv.Close()

	c, sink := newTestClient(t, time.Second, 0)
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	_, _, err := c.Do(context.Background(), req, "op")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if sink.count() != 1 {
		t.Fatalf("expected 1 emission for status 500, got %d", sink.count())
	}
	if got := sink.last().Outcome; got != OutcomeError {
		t.Errorf("outcome = %q, want %q", got, OutcomeError)
	}
	if got := sink.last().ResponseStatus; got != 500 {
		t.Errorf("status = %d, want 500", got)
	}
}

func TestDo_Status200_Fast_NoSampling_DoesNotEmit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c, sink := newTestClient(t, time.Second, 0)
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	if _, _, err := c.Do(context.Background(), req, "op"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if sink.count() != 0 {
		t.Fatalf("expected no emission for fast 2xx with sampling=0, got %d", sink.count())
	}
}

func TestDo_Status200_Slow_Emits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(60 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	// Very low slow threshold so 60ms server delay trips it.
	c, sink := newTestClient(t, 20*time.Millisecond, 0)
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	if _, _, err := c.Do(context.Background(), req, "op"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if sink.count() != 1 {
		t.Fatalf("expected 1 emission for slow 2xx, got %d", sink.count())
	}
	e := sink.last()
	if e.Outcome != OutcomeSlow {
		t.Errorf("outcome = %q, want %q", e.Outcome, OutcomeSlow)
	}
	if e.ResponseStatus != 200 {
		t.Errorf("status = %d, want 200", e.ResponseStatus)
	}
}

func TestDo_Status200_Fast_Sampling100pct_Emits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c, sink := newTestClient(t, time.Second, 1.0)
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	if _, _, err := c.Do(context.Background(), req, "op"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if sink.count() != 1 {
		t.Fatalf("expected 1 emission for fast 2xx at sampling=1.0, got %d", sink.count())
	}
	if got := sink.last().Outcome; got != OutcomeSampled {
		t.Errorf("outcome = %q, want %q", got, OutcomeSampled)
	}
}

func TestClassify_Boundaries(t *testing.T) {
	c := &Client{slowThreshold: 100 * time.Millisecond, successSampleRate: 0}

	if got := c.classify(nil, 200, 50*time.Millisecond); got != "" {
		t.Errorf("fast 200 w/ sample=0 should skip, got %q", got)
	}
	if got := c.classify(nil, 200, 101*time.Millisecond); got != OutcomeSlow {
		t.Errorf("slow 200 should be slow, got %q", got)
	}
	if got := c.classify(nil, 400, 10*time.Millisecond); got != OutcomeError {
		t.Errorf("400 should be error, got %q", got)
	}
	if got := c.classify(nil, 299, 10*time.Millisecond); got != "" {
		t.Errorf("299 fast w/ sample=0 should skip, got %q", got)
	}

	c.successSampleRate = 1.0
	if got := c.classify(nil, 200, 10*time.Millisecond); got != OutcomeSampled {
		t.Errorf("fast 200 w/ sample=1.0 should be sampled, got %q", got)
	}
}
