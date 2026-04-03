package asynq

import (
	"testing"

	"gct/config"
)

// ---------------------------------------------------------------------------
// Task type constants
// ---------------------------------------------------------------------------

func TestTaskTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"TypeImageResize", TypeImageResize, "image:resize"},
		{"TypeImageOptimize", TypeImageOptimize, "image:optimize"},
		{"TypeImageThumbnail", TypeImageThumbnail, "image:thumbnail"},
		{"TypeFileUpload", TypeFileUpload, "file:upload"},
		{"TypeFileDelete", TypeFileDelete, "file:delete"},
		{"TypeFileCompress", TypeFileCompress, "file:compress"},
		{"TypePushNotification", TypePushNotification, "notification:push"},
		{"TypeSMSNotification", TypeSMSNotification, "notification:sms"},
		{"TypeReportGenerate", TypeReportGenerate, "report:generate"},
		{"TypeReportExport", TypeReportExport, "report:export"},
		{"TypeCleanupOldSessions", TypeCleanupOldSessions, "cleanup:old_sessions"},
		{"TypeCleanupTempFiles", TypeCleanupTempFiles, "cleanup:temp_files"},
		{"TypeSystemSeed", TypeSystemSeed, "system:seed"},
		{"TypeAuditLog", TypeAuditLog, "audit:log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, tt.got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Queue name constants
// ---------------------------------------------------------------------------

func TestQueueNameConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"QueueCritical", QueueCritical, "critical"},
		{"QueueDefault", QueueDefault, "default"},
		{"QueueLow", QueueLow, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, tt.got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AsynqConfig.GetDefaultQueues
// ---------------------------------------------------------------------------

func TestGetDefaultQueues_ReturnsDefaults(t *testing.T) {
	cfg := config.AsynqConfig{}

	queues := cfg.GetDefaultQueues()

	expected := map[string]int{
		"critical": 6,
		"default":  3,
		"external": 2,
		"low":      1,
	}

	if len(queues) != len(expected) {
		t.Fatalf("expected %d queues, got %d", len(expected), len(queues))
	}

	for name, wantPrio := range expected {
		gotPrio, ok := queues[name]
		if !ok {
			t.Fatalf("expected queue %q not found", name)
		}
		if gotPrio != wantPrio {
			t.Fatalf("queue %q: expected priority %d, got %d", name, wantPrio, gotPrio)
		}
	}
}

func TestGetDefaultQueues_ReturnsCustomWhenSet(t *testing.T) {
	custom := map[string]int{
		"high":   10,
		"medium": 5,
	}
	cfg := config.AsynqConfig{
		Queues: custom,
	}

	queues := cfg.GetDefaultQueues()

	if len(queues) != len(custom) {
		t.Fatalf("expected %d queues, got %d", len(custom), len(queues))
	}

	for name, wantPrio := range custom {
		gotPrio, ok := queues[name]
		if !ok {
			t.Fatalf("expected queue %q not found", name)
		}
		if gotPrio != wantPrio {
			t.Fatalf("queue %q: expected priority %d, got %d", name, wantPrio, gotPrio)
		}
	}
}

func TestGetDefaultQueues_EmptyMapReturnsDefaults(t *testing.T) {
	cfg := config.AsynqConfig{
		Queues: map[string]int{},
	}

	queues := cfg.GetDefaultQueues()

	// Empty map has len 0, so defaults should be returned
	if _, ok := queues["critical"]; !ok {
		t.Fatal("expected default queue 'critical' when Queues is empty map")
	}
	if len(queues) != 4 {
		t.Fatalf("expected 4 default queues, got %d", len(queues))
	}
}
