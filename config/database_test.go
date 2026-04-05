package config

import "testing"

func TestRedisStore_Addr(t *testing.T) {
	tests := []struct {
		name string
		r    RedisStore
		want string
	}{
		{
			name: "returns host:port with custom values",
			r:    RedisStore{Host: "redis.example.com", Port: "6380"},
			want: "redis.example.com:6380",
		},
		{
			name: "defaults port when empty",
			r:    RedisStore{Host: "redis.example.com", Port: ""},
			want: "redis.example.com:6379",
		},
		{
			name: "defaults host when empty",
			r:    RedisStore{Host: "", Port: "6380"},
			want: "localhost:6380",
		},
		{
			name: "defaults both host and port when empty",
			r:    RedisStore{Host: "", Port: ""},
			want: "localhost:6379",
		},
		{
			name: "uses localhost and default port from zero value",
			r:    RedisStore{},
			want: "localhost:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Addr()
			if got != tt.want {
				t.Errorf("Addr() = %q, want %q", got, tt.want)
			}
		})
	}
}
