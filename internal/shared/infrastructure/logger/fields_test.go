package logger

import (
	"errors"
	"fmt"
	"testing"
)

func TestF_KV(t *testing.T) {
	t.Helper()

	tests := []struct {
		name     string
		f        F
		wantPairs map[string]any
		wantLen  int
	}{
		{
			name: "all fields populated",
			f: F{
				Op:       "CreateUser",
				Entity:   "user",
				EntityID: "uuid-123",
				Err:      errors.New("something failed"),
			},
			wantPairs: map[string]any{
				"operation": "CreateUser",
				"entity":    "user",
				"entity_id": "uuid-123",
			},
			wantLen: 8, // 4 pairs = 8 elements
		},
		{
			name: "only Op set",
			f:    F{Op: "ListItems"},
			wantPairs: map[string]any{
				"operation": "ListItems",
			},
			wantLen: 2,
		},
		{
			name: "only Err set",
			f:    F{Err: errors.New("timeout")},
			wantPairs: map[string]any{},
			wantLen: 2, // "error", <err>
		},
		{
			name:      "empty F - no fields",
			f:         F{},
			wantPairs: map[string]any{},
			wantLen:   0,
		},
		{
			name: "EntityID as int",
			f:    F{Op: "GetItem", EntityID: 42},
			wantPairs: map[string]any{
				"operation": "GetItem",
				"entity_id": 42,
			},
			wantLen: 4,
		},
		{
			name: "EntityID as string",
			f:    F{Op: "GetItem", EntityID: "abc-def"},
			wantPairs: map[string]any{
				"operation": "GetItem",
				"entity_id": "abc-def",
			},
			wantLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			got := tt.f.KV()

			if len(got) != tt.wantLen {
				t.Fatalf("len(KV()) = %d, want %d; got: %v", len(got), tt.wantLen, got)
			}

			// Verify key-value pairs (except error which needs special handling)
			kvMap := make(map[string]any)
			for i := 0; i < len(got)-1; i += 2 {
				key, ok := got[i].(string)
				if !ok {
					t.Fatalf("expected string key at index %d, got %T", i, got[i])
				}
				kvMap[key] = got[i+1]
			}

			for k, want := range tt.wantPairs {
				v, ok := kvMap[k]
				if !ok {
					t.Errorf("missing key %q in KV output", k)
					continue
				}
				if fmt.Sprint(v) != fmt.Sprint(want) {
					t.Errorf("key %q: got %v, want %v", k, v, want)
				}
			}

			// Check error field if Err is set
			if tt.f.Err != nil {
				errVal, ok := kvMap["error"]
				if !ok {
					t.Error("expected 'error' key in KV output")
				} else if errVal.(error).Error() != tt.f.Err.Error() {
					t.Errorf("error value: got %v, want %v", errVal, tt.f.Err)
				}
			}
		})
	}
}
