package validation

import "testing"

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		// Valid emails
		{"simple", "user@example.com", true},
		{"with dots", "first.last@example.com", true},
		{"with plus", "user+tag@example.com", true},
		{"with hyphen domain", "user@my-domain.com", true},
		{"with subdomain", "user@mail.example.co.uk", true},
		{"with percent", "user%name@example.com", true},
		{"with underscore", "user_name@example.com", true},
		{"numeric local", "123@example.com", true},
		{"short TLD two chars", "u@ab.cd", true},

		// Invalid emails
		{"empty string", "", false},
		{"no at sign", "userexample.com", false},
		{"no domain", "user@", false},
		{"no local part", "@example.com", false},
		{"double at", "user@@example.com", false},
		{"no TLD", "user@example", false},
		{"single char TLD", "user@example.c", false},
		{"space in local", "us er@example.com", false},
		{"space in domain", "user@exam ple.com", false},
		{"too short", "a@", false},
		{"just at", "@", false},
		{"missing local and domain", "@.", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestIsValidEmail_MaxLength(t *testing.T) {
	// Build an email that is exactly MaxEmailLen (254) characters.
	// local@domain.com => need local + 1(@) + domain + 4(.com) = 254
	// domain part: "d...d.com" where d-part = 254 - len(local) - 1(@) - 4(.com)
	local := "a"
	// remaining for domain label: 254 - 1 - 1 - 4 = 248
	domain := make([]byte, 248)
	for i := range domain {
		domain[i] = 'b'
	}
	email254 := local + "@" + string(domain) + ".com"
	if len(email254) != 254 {
		t.Fatalf("expected length 254, got %d", len(email254))
	}
	if !IsValidEmail(email254) {
		t.Error("expected 254-char email to be valid")
	}

	// 255 chars should be invalid
	email255 := "a" + email254
	if IsValidEmail(email255) {
		t.Error("expected 255-char email to be invalid")
	}
}
