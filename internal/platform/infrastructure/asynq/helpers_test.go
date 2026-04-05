package asynq

import (
	"testing"
	"time"
)

func TestBuildOptions_Empty(t *testing.T) {
	opts := TaskOptions{}
	options := opts.BuildOptions()
	if len(options) != 0 {
		t.Errorf("expected 0 options for empty TaskOptions, got %d", len(options))
	}
}

func TestBuildOptions_Queue(t *testing.T) {
	opts := TaskOptions{Queue: "critical"}
	options := opts.BuildOptions()
	if len(options) != 1 {
		t.Errorf("expected 1 option, got %d", len(options))
	}
}

func TestBuildOptions_MaxRetry(t *testing.T) {
	opts := TaskOptions{MaxRetry: 3}
	options := opts.BuildOptions()
	if len(options) != 1 {
		t.Errorf("expected 1 option, got %d", len(options))
	}
}

func TestBuildOptions_MaxRetryZero(t *testing.T) {
	opts := TaskOptions{MaxRetry: 0}
	options := opts.BuildOptions()
	if len(options) != 0 {
		t.Errorf("expected 0 options for zero MaxRetry, got %d", len(options))
	}
}

func TestBuildOptions_Timeout(t *testing.T) {
	opts := TaskOptions{Timeout: 5 * time.Second}
	options := opts.BuildOptions()
	if len(options) != 1 {
		t.Errorf("expected 1 option, got %d", len(options))
	}
}

func TestBuildOptions_Deadline(t *testing.T) {
	opts := TaskOptions{Deadline: time.Now().Add(1 * time.Hour)}
	options := opts.BuildOptions()
	if len(options) != 1 {
		t.Errorf("expected 1 option, got %d", len(options))
	}
}

func TestBuildOptions_DeadlineZero(t *testing.T) {
	opts := TaskOptions{Deadline: time.Time{}}
	options := opts.BuildOptions()
	if len(options) != 0 {
		t.Errorf("expected 0 options for zero Deadline, got %d", len(options))
	}
}

func TestBuildOptions_UniqueKey(t *testing.T) {
	opts := TaskOptions{UniqueKey: "my-unique-key"}
	options := opts.BuildOptions()
	if len(options) != 1 {
		t.Errorf("expected 1 option, got %d", len(options))
	}
}

func TestBuildOptions_UniqueTTL(t *testing.T) {
	opts := TaskOptions{UniqueTTL: 10 * time.Minute}
	options := opts.BuildOptions()
	if len(options) != 1 {
		t.Errorf("expected 1 option, got %d", len(options))
	}
}

func TestBuildOptions_ProcessIn(t *testing.T) {
	opts := TaskOptions{ProcessIn: 30 * time.Second}
	options := opts.BuildOptions()
	if len(options) != 1 {
		t.Errorf("expected 1 option, got %d", len(options))
	}
}

func TestBuildOptions_ProcessAt(t *testing.T) {
	opts := TaskOptions{ProcessAt: time.Now().Add(1 * time.Hour)}
	options := opts.BuildOptions()
	if len(options) != 1 {
		t.Errorf("expected 1 option, got %d", len(options))
	}
}

func TestBuildOptions_ProcessAtZero(t *testing.T) {
	opts := TaskOptions{ProcessAt: time.Time{}}
	options := opts.BuildOptions()
	if len(options) != 0 {
		t.Errorf("expected 0 options for zero ProcessAt, got %d", len(options))
	}
}

func TestBuildOptions_Retention(t *testing.T) {
	opts := TaskOptions{Retention: 24 * time.Hour}
	options := opts.BuildOptions()
	if len(options) != 1 {
		t.Errorf("expected 1 option, got %d", len(options))
	}
}

func TestBuildOptions_AllFields(t *testing.T) {
	opts := TaskOptions{
		Queue:     "critical",
		MaxRetry:  5,
		Timeout:   10 * time.Second,
		Deadline:  time.Now().Add(1 * time.Hour),
		UniqueKey: "unique-123",
		UniqueTTL: 15 * time.Minute,
		ProcessIn: 1 * time.Minute,
		ProcessAt: time.Now().Add(30 * time.Minute),
		Retention: 48 * time.Hour,
	}
	options := opts.BuildOptions()
	if len(options) != 9 {
		t.Errorf("expected 9 options for all fields set, got %d", len(options))
	}
}

func TestBuildOptions_MultipleFields(t *testing.T) {
	opts := TaskOptions{
		Queue:    "default",
		MaxRetry: 3,
		Timeout:  5 * time.Second,
	}
	options := opts.BuildOptions()
	if len(options) != 3 {
		t.Errorf("expected 3 options, got %d", len(options))
	}
}
