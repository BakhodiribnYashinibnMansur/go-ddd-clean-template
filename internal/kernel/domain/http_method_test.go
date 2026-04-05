package domain_test

import (
	"testing"

	"gct/internal/kernel/domain"
)

func TestNewHTTPMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    domain.HTTPMethod
		wantErr bool
	}{
		{"get upper", "GET", domain.HTTPMethodGet, false},
		{"post upper", "POST", domain.HTTPMethodPost, false},
		{"put upper", "PUT", domain.HTTPMethodPut, false},
		{"patch upper", "PATCH", domain.HTTPMethodPatch, false},
		{"delete upper", "DELETE", domain.HTTPMethodDelete, false},
		{"head upper", "HEAD", domain.HTTPMethodHead, false},
		{"options upper", "OPTIONS", domain.HTTPMethodOptions, false},
		{"get lower normalized", "get", domain.HTTPMethodGet, false},
		{"post mixed normalized", "PoSt", domain.HTTPMethodPost, false},
		{"delete lower normalized", "delete", domain.HTTPMethodDelete, false},
		{"whitespace trimmed", "  GET  ", domain.HTTPMethodGet, false},
		{"invalid empty", "", "", true},
		{"invalid foo", "FOO", "", true},
		{"invalid connect", "CONNECT", "", true},
		{"invalid trace", "TRACE", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := domain.NewHTTPMethod(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewHTTPMethod(%q) err = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewHTTPMethod(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHTTPMethod_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input domain.HTTPMethod
		want  bool
	}{
		{"GET valid", domain.HTTPMethodGet, true},
		{"POST valid", domain.HTTPMethodPost, true},
		{"PUT valid", domain.HTTPMethodPut, true},
		{"PATCH valid", domain.HTTPMethodPatch, true},
		{"DELETE valid", domain.HTTPMethodDelete, true},
		{"HEAD valid", domain.HTTPMethodHead, true},
		{"OPTIONS valid", domain.HTTPMethodOptions, true},
		{"empty invalid", domain.HTTPMethod(""), false},
		{"lowercase invalid", domain.HTTPMethod("get"), false},
		{"unknown invalid", domain.HTTPMethod("FOO"), false},
		{"CONNECT invalid", domain.HTTPMethod("CONNECT"), false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.input.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPMethod_String(t *testing.T) {
	t.Parallel()

	if got := domain.HTTPMethodGet.String(); got != "GET" {
		t.Errorf("String() = %q, want GET", got)
	}
	if got := domain.HTTPMethodPost.String(); got != "POST" {
		t.Errorf("String() = %q, want POST", got)
	}
}
