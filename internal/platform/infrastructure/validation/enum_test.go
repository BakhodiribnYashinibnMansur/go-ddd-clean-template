package validation

import "testing"

func TestIsEnumValid_String(t *testing.T) {
	allowed := []string{"active", "inactive", "pending"}

	tests := []struct {
		name string
		val  string
		want bool
	}{
		{"valid first", "active", true},
		{"valid middle", "inactive", true},
		{"valid last", "pending", true},
		{"invalid value", "deleted", false},
		{"empty string", "", false},
		{"case sensitive", "Active", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEnumValid(tt.val, allowed)
			if got != tt.want {
				t.Errorf("IsEnumValid(%q, ...) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestIsEnumValid_Int(t *testing.T) {
	allowed := []int{1, 2, 3}

	tests := []struct {
		name string
		val  int
		want bool
	}{
		{"valid 1", 1, true},
		{"valid 3", 3, true},
		{"invalid 0", 0, false},
		{"invalid 4", 4, false},
		{"invalid negative", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEnumValid(tt.val, allowed)
			if got != tt.want {
				t.Errorf("IsEnumValid(%d, ...) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func TestIsEnumValid_EmptyAllowed(t *testing.T) {
	if IsEnumValid("anything", []string{}) {
		t.Error("expected false for empty allowed list")
	}
}
