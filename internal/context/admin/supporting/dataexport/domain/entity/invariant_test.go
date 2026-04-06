package entity

import (
	"testing"

	"gct/internal/context/admin/supporting/dataexport/domain/event"

	"github.com/google/uuid"
)

func newTestExport() *DataExport {
	return NewDataExport(uuid.New(), "users", "csv")
}

func TestInvariant_NewExport_IsPending(t *testing.T) {
	de := newTestExport()
	if de.Status() != ExportStatusPending {
		t.Fatalf("expected status %q, got %q", ExportStatusPending, de.Status())
	}
}

func TestInvariant_NewExport_HasNoFileURL(t *testing.T) {
	de := newTestExport()
	if de.FileURL() != nil {
		t.Fatalf("expected nil FileURL on new export, got %v", *de.FileURL())
	}
}

func TestInvariant_NewExport_HasNoError(t *testing.T) {
	de := newTestExport()
	if de.Error() != nil {
		t.Fatalf("expected nil Error on new export, got %v", *de.Error())
	}
}

func TestInvariant_NewExport_RaisesEvent(t *testing.T) {
	de := newTestExport()
	events := de.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	evt, ok := events[0].(event.ExportRequested)
	if !ok {
		t.Fatalf("expected ExportRequested event, got %T", events[0])
	}
	if evt.EventName() != "dataexport.requested" {
		t.Errorf("expected event name %q, got %q", "dataexport.requested", evt.EventName())
	}
}

func TestInvariant_Complete_SetsFileURL(t *testing.T) {
	de := newTestExport()
	de.StartProcessing()
	de.Complete("https://storage.example.com/export.csv")

	if de.FileURL() == nil {
		t.Fatal("expected non-nil FileURL after Complete")
	}
	if *de.FileURL() != "https://storage.example.com/export.csv" {
		t.Errorf("expected FileURL %q, got %q", "https://storage.example.com/export.csv", *de.FileURL())
	}
	if de.Status() != ExportStatusCompleted {
		t.Errorf("expected status %q, got %q", ExportStatusCompleted, de.Status())
	}
}

func TestInvariant_Complete_RaisesEvent(t *testing.T) {
	de := newTestExport()
	de.ClearEvents() // clear the ExportRequested event
	de.StartProcessing()
	de.Complete("https://storage.example.com/export.csv")

	events := de.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event after Complete, got %d", len(events))
	}
	evt, ok := events[0].(event.ExportCompleted)
	if !ok {
		t.Fatalf("expected ExportCompleted event, got %T", events[0])
	}
	if evt.EventName() != "dataexport.completed" {
		t.Errorf("expected event name %q, got %q", "dataexport.completed", evt.EventName())
	}
	if evt.FileURL != "https://storage.example.com/export.csv" {
		t.Errorf("expected FileURL in event %q, got %q", "https://storage.example.com/export.csv", evt.FileURL)
	}
}

func TestInvariant_Fail_SetsErrorMsg(t *testing.T) {
	de := newTestExport()
	de.StartProcessing()
	de.Fail("disk full")

	if de.Error() == nil {
		t.Fatal("expected non-nil Error after Fail")
	}
	if *de.Error() != "disk full" {
		t.Errorf("expected error message %q, got %q", "disk full", *de.Error())
	}
	if de.Status() != ExportStatusFailed {
		t.Errorf("expected status %q, got %q", ExportStatusFailed, de.Status())
	}
}

func TestInvariant_Fail_NoEvent(t *testing.T) {
	de := newTestExport()
	de.ClearEvents() // clear the ExportRequested event
	de.StartProcessing()
	de.Fail("timeout")

	events := de.Events()
	if len(events) != 0 {
		t.Fatalf("expected 0 events after Fail, got %d", len(events))
	}
}

func TestInvariant_FullLifecycle_HappyPath(t *testing.T) {
	de := newTestExport()

	// Step 1: starts PENDING
	if de.Status() != ExportStatusPending {
		t.Fatalf("expected initial status %q, got %q", ExportStatusPending, de.Status())
	}

	// Step 2: transition to PROCESSING
	de.StartProcessing()
	if de.Status() != ExportStatusProcessing {
		t.Fatalf("expected status %q after StartProcessing, got %q", ExportStatusProcessing, de.Status())
	}

	// Step 3: transition to COMPLETED
	de.Complete("https://cdn.example.com/file.csv")
	if de.Status() != ExportStatusCompleted {
		t.Fatalf("expected status %q after Complete, got %q", ExportStatusCompleted, de.Status())
	}
	if de.FileURL() == nil {
		t.Fatal("expected non-nil FileURL after Complete")
	}
	if de.Error() != nil {
		t.Fatal("expected nil Error on happy path")
	}

	// Verify events: ExportRequested + ExportCompleted
	events := de.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].EventName() != "dataexport.requested" {
		t.Errorf("expected first event %q, got %q", "dataexport.requested", events[0].EventName())
	}
	if events[1].EventName() != "dataexport.completed" {
		t.Errorf("expected second event %q, got %q", "dataexport.completed", events[1].EventName())
	}
}

func TestInvariant_FullLifecycle_FailurePath(t *testing.T) {
	de := newTestExport()

	// Step 1: starts PENDING
	if de.Status() != ExportStatusPending {
		t.Fatalf("expected initial status %q, got %q", ExportStatusPending, de.Status())
	}

	// Step 2: transition to PROCESSING
	de.StartProcessing()
	if de.Status() != ExportStatusProcessing {
		t.Fatalf("expected status %q after StartProcessing, got %q", ExportStatusProcessing, de.Status())
	}

	// Step 3: transition to FAILED
	de.Fail("connection refused")
	if de.Status() != ExportStatusFailed {
		t.Fatalf("expected status %q after Fail, got %q", ExportStatusFailed, de.Status())
	}
	if de.Error() == nil {
		t.Fatal("expected non-nil Error after Fail")
	}
	if de.FileURL() != nil {
		t.Fatal("expected nil FileURL on failure path")
	}

	// Verify events: only ExportRequested (Fail raises none)
	events := de.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "dataexport.requested" {
		t.Errorf("expected event %q, got %q", "dataexport.requested", events[0].EventName())
	}
}
