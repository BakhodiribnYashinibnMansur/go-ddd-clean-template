package domain_test

import (
	"testing"

	"gct/internal/kernel/domain"
)

func TestNewPercentage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"zero", 0, false},
		{"fifty", 50, false},
		{"hundred", 100, false},
		{"negative", -1, true},
		{"over hundred", 101, true},
		{"far negative", -100, true},
		{"far over", 1000, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := domain.NewPercentage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPercentage(%d) err = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if err == nil && got.Int() != tt.input {
				t.Errorf("Int() = %d, want %d", got.Int(), tt.input)
			}
		})
	}
}

func TestPercentage_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input int
		want  string
	}{
		{0, "0%"},
		{50, "50%"},
		{100, "100%"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			p, err := domain.NewPercentage(tt.input)
			if err != nil {
				t.Fatalf("NewPercentage(%d) unexpected error: %v", tt.input, err)
			}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
