package apikey_test

import (
	"testing"

	"gct/internal/kernel/infrastructure/security/apikey"
)

func TestGenerate_Prefixes(t *testing.T) {
	tests := []struct {
		integration string
		wantPrefix  string
	}{
		{"gct-admin", apikey.PrefixAdmin},
		{"gct-client", apikey.PrefixClient},
		{"gct-mobile", apikey.PrefixMobile},
		{"something-else", apikey.PrefixOther},
	}
	for _, tt := range tests {
		key, err := apikey.Generate(tt.integration)
		if err != nil {
			t.Fatalf("Generate(%q) error: %v", tt.integration, err)
		}
		if len(key) < len(tt.wantPrefix) || key[:len(tt.wantPrefix)] != tt.wantPrefix {
			t.Errorf("Generate(%q) = %q, want prefix %q", tt.integration, key, tt.wantPrefix)
		}
	}
}

func TestGenerate_MinLength(t *testing.T) {
	key, err := apikey.Generate("gct-admin")
	if err != nil {
		t.Fatal(err)
	}
	// prefix (8) + 43 base64url chars for 32 bytes = 51 minimum.
	if len(key) < 40 {
		t.Fatalf("key too short: %d chars, got %q", len(key), key)
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	a, err := apikey.Generate("gct-admin")
	if err != nil {
		t.Fatal(err)
	}
	b, err := apikey.Generate("gct-admin")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("two Generate calls must produce different keys")
	}
}

func TestMask_KnownPrefix(t *testing.T) {
	got := apikey.Mask("gct_adm_abc123xyz4567890")
	want := "gct_adm_****...7890"
	if got != want {
		t.Errorf("Mask = %q, want %q", got, want)
	}
}

func TestMask_Short(t *testing.T) {
	got := apikey.Mask("short")
	if got != "****" {
		t.Errorf("Mask(short) = %q, want %q", got, "****")
	}
}

func TestMask_UnknownPrefix(t *testing.T) {
	got := apikey.Mask("unknown_prefix_long_key_here_1234")
	want := "****...1234"
	if got != want {
		t.Errorf("Mask = %q, want %q", got, want)
	}
}

func TestMask_ClientPrefix(t *testing.T) {
	got := apikey.Mask("gct_cli_something_else_abcd")
	want := "gct_cli_****...abcd"
	if got != want {
		t.Errorf("Mask = %q, want %q", got, want)
	}
}
