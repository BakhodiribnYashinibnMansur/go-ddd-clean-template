package validation

import (
	"testing"

	"github.com/google/uuid"
)

func TestIsValidUUID(t *testing.T) {
	validID := uuid.New().String()

	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"valid v4", validID, true},
		{"valid nil uuid", "00000000-0000-0000-0000-000000000000", true},
		{"valid uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"valid lowercase", "550e8400-e29b-41d4-a716-446655440000", true},

		{"empty", "", false},
		{"too short", "550e8400-e29b", false},
		{"invalid chars", "gggggggg-gggg-gggg-gggg-gggggggggggg", false},
		{"no dashes", "550e8400e29b41d4a716446655440000", true}, // google/uuid accepts this
		{"random string", "not-a-uuid", false},
		{"spaces", "550e8400 e29b 41d4 a716 446655440000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidUUID(tt.id)
			if got != tt.want {
				t.Errorf("IsValidUUID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}
