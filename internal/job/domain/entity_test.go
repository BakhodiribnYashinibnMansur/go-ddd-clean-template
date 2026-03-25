package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/job/domain"

	"github.com/google/uuid"
)

func TestNewJob(t *testing.T) {
	payload := map[string]any{"key": "value"}
	scheduledAt := time.Now().Add(time.Hour)
	j := domain.NewJob("send_email", payload, 3, &scheduledAt)

	if j.TaskName() != "send_email" {
		t.Fatalf("expected task name send_email, got %s", j.TaskName())
	}
	if j.Status() != domain.JobStatusPending {
		t.Fatalf("expected status PENDING, got %s", j.Status())
	}
	if j.Attempts() != 0 {
		t.Fatalf("expected 0 attempts, got %d", j.Attempts())
	}
	if j.MaxAttempts() != 3 {
		t.Fatalf("expected 3 max attempts, got %d", j.MaxAttempts())
	}
	if j.ScheduledAt() == nil {
		t.Fatal("scheduled_at should not be nil")
	}

	events := j.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "job.scheduled" {
		t.Fatalf("expected job.scheduled, got %s", events[0].EventName())
	}
}

func TestJob_Complete(t *testing.T) {
	j := domain.NewJob("process_data", nil, 1, nil)
	j.Start()
	result := map[string]any{"processed": 100}
	j.Complete(result)

	if j.Status() != domain.JobStatusCompleted {
		t.Fatalf("expected status COMPLETED, got %s", j.Status())
	}
	if j.CompletedAt() == nil {
		t.Fatal("completed_at should not be nil")
	}
	if j.Attempts() != 1 {
		t.Fatalf("expected 1 attempt, got %d", j.Attempts())
	}

	// Should have JobScheduled + JobCompleted events.
	events := j.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[1].EventName() != "job.completed" {
		t.Fatalf("expected job.completed, got %s", events[1].EventName())
	}
}

func TestJob_Fail(t *testing.T) {
	j := domain.NewJob("import_data", nil, 3, nil)
	j.Start()
	j.Fail("connection timeout")

	if j.Status() != domain.JobStatusFailed {
		t.Fatalf("expected status FAILED, got %s", j.Status())
	}
	if j.Error() == nil || *j.Error() != "connection timeout" {
		t.Fatal("error message mismatch")
	}
}

func TestReconstructJob(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	payload := map[string]any{"k": "v"}
	result := map[string]any{"ok": true}
	errMsg := "some error"

	j := domain.ReconstructJob(
		id, now, now,
		"my_task", domain.JobStatusFailed,
		payload, result, 2, 5,
		nil, &now, nil, &errMsg,
	)

	if j.ID() != id {
		t.Fatal("ID mismatch")
	}
	if j.TaskName() != "my_task" {
		t.Fatal("task name mismatch")
	}
	if j.Status() != domain.JobStatusFailed {
		t.Fatal("status mismatch")
	}
	if len(j.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(j.Events()))
	}
}
