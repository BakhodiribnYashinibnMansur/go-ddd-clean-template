package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/admin/supporting/dataexport/domain"

	"github.com/google/uuid"
)

func TestNewDataExport(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	de := domain.NewDataExport(userID, "users", "csv")

	if de.UserID() != userID {
		t.Fatal("user ID mismatch")
	}
	if de.DataType() != "users" {
		t.Fatalf("expected data type users, got %s", de.DataType())
	}
	if de.Format() != "csv" {
		t.Fatalf("expected format csv, got %s", de.Format())
	}
	if de.Status() != domain.ExportStatusPending {
		t.Fatalf("expected status PENDING, got %s", de.Status())
	}
	if de.FileURL() != nil {
		t.Fatal("file URL should be nil")
	}

	events := de.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "dataexport.requested" {
		t.Fatalf("expected dataexport.requested, got %s", events[0].EventName())
	}
}

func TestDataExport_Complete(t *testing.T) {
	t.Parallel()

	de := domain.NewDataExport(uuid.New(), "orders", "xlsx")
	de.StartProcessing()

	if de.Status() != domain.ExportStatusProcessing {
		t.Fatalf("expected status PROCESSING, got %s", de.Status())
	}

	de.Complete("https://example.com/export.xlsx")

	if de.Status() != domain.ExportStatusCompleted {
		t.Fatalf("expected status COMPLETED, got %s", de.Status())
	}
	if de.FileURL() == nil || *de.FileURL() != "https://example.com/export.xlsx" {
		t.Fatal("file URL mismatch")
	}

	events := de.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[1].EventName() != "dataexport.completed" {
		t.Fatalf("expected dataexport.completed, got %s", events[1].EventName())
	}
}

func TestDataExport_Fail(t *testing.T) {
	t.Parallel()

	de := domain.NewDataExport(uuid.New(), "logs", "json")
	de.Fail("disk full")

	if de.Status() != domain.ExportStatusFailed {
		t.Fatalf("expected status FAILED, got %s", de.Status())
	}
	if de.Error() == nil || *de.Error() != "disk full" {
		t.Fatal("error message mismatch")
	}
}

func TestReconstructDataExport(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	fileURL := "https://example.com/file.csv"

	de := domain.ReconstructDataExport(id, now, now, userID, "users", "csv", domain.ExportStatusCompleted, &fileURL, nil)

	if de.ID() != id {
		t.Fatal("ID mismatch")
	}
	if de.UserID() != userID {
		t.Fatal("user ID mismatch")
	}
	if de.Status() != domain.ExportStatusCompleted {
		t.Fatal("status mismatch")
	}
	if len(de.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(de.Events()))
	}
}
