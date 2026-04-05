package validation

import "testing"

func TestIsValidPhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		// Valid phones
		{"7 digits", "1234567", true},
		{"10 digits", "1234567890", true},
		{"with plus prefix", "+1234567890", true},
		{"with country code", "+998901234567", true},
		{"with spaces", "+1 234 567 890", true},
		{"with dashes", "+1-234-567-890", true},
		{"with parens", "+1(234)5678901", true},
		{"mixed format", "+1 (234) 567-8901", true},

		// Invalid phones
		{"empty", "", false},
		{"too few digits", "12345", false},
		{"6 digits", "123456", false},
		{"letters only", "abcdefg", false},
		{"mixed letters digits short", "abc1234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPhone(tt.phone)
			if got != tt.want {
				t.Errorf("IsValidPhone(%q) = %v, want %v", tt.phone, got, tt.want)
			}
		})
	}
}

func TestSanitizePhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"digits only", "1234567890", "1234567890"},
		{"with plus", "+1234567890", "+1234567890"},
		{"with spaces", "+1 234 567 890", "+1234567890"},
		{"with dashes", "+1-234-567-890", "+1234567890"},
		{"with parens", "+1(234)5678901", "+12345678901"},
		{"mixed format", "+1 (234) 567-8901", "+12345678901"},
		{"no plus with formatting", "1 (234) 567-8901", "12345678901"},
		{"plus in middle ignored", "123+456", "123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizePhone(tt.input)
			if got != tt.want {
				t.Errorf("SanitizePhone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
