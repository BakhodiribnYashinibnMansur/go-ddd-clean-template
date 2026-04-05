package domain_test

import (
	"errors"
	"testing"

	"gct/internal/context/content/generic/file/domain"
)

func TestNewMimeType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		wantType    string
		wantSubtype string
	}{
		{name: "valid image/png", input: "image/png", wantType: "image", wantSubtype: "png"},
		{name: "valid application/pdf", input: "application/pdf", wantType: "application", wantSubtype: "pdf"},
		{name: "valid text/html with charset", input: "text/html; charset=utf-8", wantType: "text", wantSubtype: "html; charset=utf-8"},
		{name: "invalid notvalid", input: "notvalid", wantErr: true},
		{name: "invalid empty", input: "", wantErr: true},
		{name: "invalid missing subtype", input: "image/", wantErr: true},
		{name: "invalid missing type", input: "/png", wantErr: true},
		{name: "invalid multiple slashes", input: "a/b/c", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := domain.NewMimeType(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tt.input)
				}
				if !errors.Is(err, domain.ErrInvalidMimeType) {
					t.Fatalf("expected ErrInvalidMimeType, got %v", err)
				}
				if !got.IsZero() {
					t.Fatalf("expected zero value on error, got %q", got.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.input {
				t.Errorf("String() = %q, want %q", got.String(), tt.input)
			}
			if got.Type() != tt.wantType {
				t.Errorf("Type() = %q, want %q", got.Type(), tt.wantType)
			}
			if got.Subtype() != tt.wantSubtype {
				t.Errorf("Subtype() = %q, want %q", got.Subtype(), tt.wantSubtype)
			}
			if got.IsZero() {
				t.Errorf("IsZero() = true, want false")
			}
		})
	}
}

func TestMimeType_IsZero(t *testing.T) {
	t.Parallel()
	var m domain.MimeType
	if !m.IsZero() {
		t.Errorf("zero value IsZero() = false, want true")
	}
}
