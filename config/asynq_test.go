package config

import "testing"

func TestAsynqConfig_GetDefaultQueues(t *testing.T) {
	tests := []struct {
		name   string
		queues map[string]int
		want   map[string]int
	}{
		{
			name:   "returns default queues when none configured",
			queues: nil,
			want: map[string]int{
				"critical": 6,
				"default":  3,
				"external": 2,
				"low":      1,
			},
		},
		{
			name:   "returns default queues when empty map configured",
			queues: map[string]int{},
			want: map[string]int{
				"critical": 6,
				"default":  3,
				"external": 2,
				"low":      1,
			},
		},
		{
			name: "returns custom queues when configured",
			queues: map[string]int{
				"high":   10,
				"medium": 5,
			},
			want: map[string]int{
				"high":   10,
				"medium": 5,
			},
		},
		{
			name: "returns single custom queue",
			queues: map[string]int{
				"only": 1,
			},
			want: map[string]int{
				"only": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &AsynqConfig{Queues: tt.queues}
			got := cfg.GetDefaultQueues()

			if len(got) != len(tt.want) {
				t.Fatalf("GetDefaultQueues() returned %d queues, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("GetDefaultQueues() missing key %q", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("GetDefaultQueues()[%q] = %d, want %d", k, gotV, wantV)
				}
			}
		})
	}
}
