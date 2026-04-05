package redis

import (
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func TestWithPoolSize(t *testing.T) {
	opts := &goredis.Options{}
	WithPoolSize(20)(opts)
	if opts.PoolSize != 20 {
		t.Errorf("PoolSize = %d, want 20", opts.PoolSize)
	}
}

func TestWithMinIdleConns(t *testing.T) {
	opts := &goredis.Options{}
	WithMinIdleConns(5)(opts)
	if opts.MinIdleConns != 5 {
		t.Errorf("MinIdleConns = %d, want 5", opts.MinIdleConns)
	}
}

func TestWithDB(t *testing.T) {
	opts := &goredis.Options{}
	WithDB(3)(opts)
	if opts.DB != 3 {
		t.Errorf("DB = %d, want 3", opts.DB)
	}
}

func TestWithDialTimeout(t *testing.T) {
	opts := &goredis.Options{}
	WithDialTimeout(10 * time.Second)(opts)
	if opts.DialTimeout != 10*time.Second {
		t.Errorf("DialTimeout = %v, want 10s", opts.DialTimeout)
	}
}

func TestWithReadTimeout(t *testing.T) {
	opts := &goredis.Options{}
	WithReadTimeout(5 * time.Second)(opts)
	if opts.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout = %v, want 5s", opts.ReadTimeout)
	}
}

func TestWithWriteTimeout(t *testing.T) {
	opts := &goredis.Options{}
	WithWriteTimeout(3 * time.Second)(opts)
	if opts.WriteTimeout != 3*time.Second {
		t.Errorf("WriteTimeout = %v, want 3s", opts.WriteTimeout)
	}
}

func TestMultipleOptionsAppliedInOrder(t *testing.T) {
	opts := &goredis.Options{}

	// Apply first pool size, then override with a second value.
	options := []Option{
		WithPoolSize(10),
		WithMinIdleConns(2),
		WithDB(1),
		WithDialTimeout(5 * time.Second),
		WithReadTimeout(3 * time.Second),
		WithWriteTimeout(2 * time.Second),
		WithPoolSize(50), // override earlier value
	}

	for _, o := range options {
		o(opts)
	}

	if opts.PoolSize != 50 {
		t.Errorf("PoolSize = %d, want 50 (last write wins)", opts.PoolSize)
	}
	if opts.MinIdleConns != 2 {
		t.Errorf("MinIdleConns = %d, want 2", opts.MinIdleConns)
	}
	if opts.DB != 1 {
		t.Errorf("DB = %d, want 1", opts.DB)
	}
	if opts.DialTimeout != 5*time.Second {
		t.Errorf("DialTimeout = %v, want 5s", opts.DialTimeout)
	}
	if opts.ReadTimeout != 3*time.Second {
		t.Errorf("ReadTimeout = %v, want 3s", opts.ReadTimeout)
	}
	if opts.WriteTimeout != 2*time.Second {
		t.Errorf("WriteTimeout = %v, want 2s", opts.WriteTimeout)
	}
}

func TestZeroValues(t *testing.T) {
	// Pre-populate with non-zero values, then apply zero-value options.
	opts := &goredis.Options{
		PoolSize:     10,
		MinIdleConns: 5,
		DB:           3,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	options := []Option{
		WithPoolSize(0),
		WithMinIdleConns(0),
		WithDB(0),
		WithDialTimeout(0),
		WithReadTimeout(0),
		WithWriteTimeout(0),
	}

	for _, o := range options {
		o(opts)
	}

	if opts.PoolSize != 0 {
		t.Errorf("PoolSize = %d, want 0", opts.PoolSize)
	}
	if opts.MinIdleConns != 0 {
		t.Errorf("MinIdleConns = %d, want 0", opts.MinIdleConns)
	}
	if opts.DB != 0 {
		t.Errorf("DB = %d, want 0", opts.DB)
	}
	if opts.DialTimeout != 0 {
		t.Errorf("DialTimeout = %v, want 0", opts.DialTimeout)
	}
	if opts.ReadTimeout != 0 {
		t.Errorf("ReadTimeout = %v, want 0", opts.ReadTimeout)
	}
	if opts.WriteTimeout != 0 {
		t.Errorf("WriteTimeout = %v, want 0", opts.WriteTimeout)
	}
}

func TestAllOptions_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		option  Option
		check   func(*goredis.Options) bool
		wantStr string
	}{
		{
			name:    "pool size 25",
			option:  WithPoolSize(25),
			check:   func(o *goredis.Options) bool { return o.PoolSize == 25 },
			wantStr: "PoolSize=25",
		},
		{
			name:    "min idle conns 8",
			option:  WithMinIdleConns(8),
			check:   func(o *goredis.Options) bool { return o.MinIdleConns == 8 },
			wantStr: "MinIdleConns=8",
		},
		{
			name:    "db 7",
			option:  WithDB(7),
			check:   func(o *goredis.Options) bool { return o.DB == 7 },
			wantStr: "DB=7",
		},
		{
			name:    "dial timeout 15s",
			option:  WithDialTimeout(15 * time.Second),
			check:   func(o *goredis.Options) bool { return o.DialTimeout == 15*time.Second },
			wantStr: "DialTimeout=15s",
		},
		{
			name:    "read timeout 7s",
			option:  WithReadTimeout(7 * time.Second),
			check:   func(o *goredis.Options) bool { return o.ReadTimeout == 7*time.Second },
			wantStr: "ReadTimeout=7s",
		},
		{
			name:    "write timeout 4s",
			option:  WithWriteTimeout(4 * time.Second),
			check:   func(o *goredis.Options) bool { return o.WriteTimeout == 4*time.Second },
			wantStr: "WriteTimeout=4s",
		},
		{
			name:    "pool size zero",
			option:  WithPoolSize(0),
			check:   func(o *goredis.Options) bool { return o.PoolSize == 0 },
			wantStr: "PoolSize=0",
		},
		{
			name:    "dial timeout zero",
			option:  WithDialTimeout(0),
			check:   func(o *goredis.Options) bool { return o.DialTimeout == 0 },
			wantStr: "DialTimeout=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &goredis.Options{}
			tt.option(opts)
			if !tt.check(opts) {
				t.Errorf("option did not set expected value: want %s", tt.wantStr)
			}
		})
	}
}
