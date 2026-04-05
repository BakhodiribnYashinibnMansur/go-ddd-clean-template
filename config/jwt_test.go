package config

import (
	"errors"
	"testing"
	"time"
)

func TestJWT_Validate(t *testing.T) {
	validJWT := JWT{
		PrivateKey: "test-private-key",
		PublicKey:  "test-public-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
	}

	tests := []struct {
		name    string
		jwt     JWT
		wantErr error
	}{
		{
			name:    "valid configuration",
			jwt:     validJWT,
			wantErr: nil,
		},
		{
			name: "missing private key",
			jwt: JWT{
				PublicKey:  "test-public-key",
				AccessTTL:  15 * time.Minute,
				RefreshTTL: 24 * time.Hour,
			},
			wantErr: ErrMissingJWTPrivateKey,
		},
		{
			name: "missing public key",
			jwt: JWT{
				PrivateKey: "test-private-key",
				AccessTTL:  15 * time.Minute,
				RefreshTTL: 24 * time.Hour,
			},
			wantErr: ErrMissingJWTPublicKey,
		},
		{
			name: "zero access TTL",
			jwt: JWT{
				PrivateKey: "test-private-key",
				PublicKey:  "test-public-key",
				AccessTTL:  0,
				RefreshTTL: 24 * time.Hour,
			},
			wantErr: ErrInvalidAccessTTL,
		},
		{
			name: "negative access TTL",
			jwt: JWT{
				PrivateKey: "test-private-key",
				PublicKey:  "test-public-key",
				AccessTTL:  -1 * time.Minute,
				RefreshTTL: 24 * time.Hour,
			},
			wantErr: ErrInvalidAccessTTL,
		},
		{
			name: "zero refresh TTL",
			jwt: JWT{
				PrivateKey: "test-private-key",
				PublicKey:  "test-public-key",
				AccessTTL:  15 * time.Minute,
				RefreshTTL: 0,
			},
			wantErr: ErrInvalidRefreshTTL,
		},
		{
			name: "negative refresh TTL",
			jwt: JWT{
				PrivateKey: "test-private-key",
				PublicKey:  "test-public-key",
				AccessTTL:  15 * time.Minute,
				RefreshTTL: -1 * time.Hour,
			},
			wantErr: ErrInvalidRefreshTTL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.jwt.Validate()

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate() expected error %v, got nil", tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
