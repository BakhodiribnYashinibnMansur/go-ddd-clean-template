package ptrutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --------------- StrVal ---------------

func TestStrVal(t *testing.T) {
	tests := []struct {
		name string
		in   *string
		want string
	}{
		{"nil returns empty", nil, ""},
		{"non-nil returns value", strPtr("hello"), "hello"},
		{"empty string pointer", strPtr(""), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, StrVal(tt.in))
		})
	}
}

// --------------- IntVal ---------------

func TestIntVal(t *testing.T) {
	tests := []struct {
		name string
		in   *int
		want int
	}{
		{"nil returns zero", nil, 0},
		{"positive value", intPtr(42), 42},
		{"negative value", intPtr(-1), -1},
		{"zero value", intPtr(0), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IntVal(tt.in))
		})
	}
}

// --------------- BoolVal ---------------

func TestBoolVal(t *testing.T) {
	tests := []struct {
		name string
		in   *bool
		want bool
	}{
		{"nil returns false", nil, false},
		{"true pointer", boolPtr(true), true},
		{"false pointer", boolPtr(false), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BoolVal(tt.in))
		})
	}
}

// --------------- Ptr ---------------

func TestPtr(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		p := Ptr("hello")
		assert.NotNil(t, p)
		assert.Equal(t, "hello", *p)
	})

	t.Run("int", func(t *testing.T) {
		p := Ptr(42)
		assert.NotNil(t, p)
		assert.Equal(t, 42, *p)
	})

	t.Run("bool", func(t *testing.T) {
		p := Ptr(true)
		assert.NotNil(t, p)
		assert.Equal(t, true, *p)
	})

	t.Run("struct", func(t *testing.T) {
		type S struct{ X int }
		p := Ptr(S{X: 7})
		assert.NotNil(t, p)
		assert.Equal(t, 7, p.X)
	})

	t.Run("zero value", func(t *testing.T) {
		p := Ptr(0)
		assert.NotNil(t, p)
		assert.Equal(t, 0, *p)
	})
}

// helpers
func strPtr(s string) *string  { return &s }
func intPtr(i int) *int        { return &i }
func boolPtr(b bool) *bool     { return &b }
