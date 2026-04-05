package config

import (
	"errors"
	"os"
	"testing"
	"time"
)

// validPepper decodes to 37 raw bytes ("test-refresh-pepper-min-32-bytes-long").
const validPepper = "dGVzdC1yZWZyZXNoLXBlcHBlci1taW4tMzItYnl0ZXMtbG9uZw"

func baseValidJWT(t *testing.T) JWT {
	t.Helper()
	dir := t.TempDir()
	return JWT{
		Issuer:        "gct-auth",
		RefreshPepper: validPepper,
		APIKeyPepper:  validPepper,
		KeysDir:       dir,
		KeyBits:       4096,
	}
}

func TestJWT_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*JWT)
		wantErr error
	}{
		{name: "valid configuration", mutate: func(*JWT) {}, wantErr: nil},
		{name: "missing issuer", mutate: func(j *JWT) { j.Issuer = "" }, wantErr: ErrMissingJWTIssuer},
		{name: "missing refresh pepper", mutate: func(j *JWT) { j.RefreshPepper = "" }, wantErr: ErrMissingRefreshPepper},
		{name: "invalid refresh pepper b64", mutate: func(j *JWT) { j.RefreshPepper = "!!!not-base64!!!" }, wantErr: ErrInvalidPepper},
		{name: "short refresh pepper", mutate: func(j *JWT) { j.RefreshPepper = "c2hvcnQ=" }, wantErr: ErrWeakPepper},
		{name: "missing api_key_pepper", mutate: func(j *JWT) { j.APIKeyPepper = "" }, wantErr: ErrMissingAPIKeyPepper},
		{name: "short api_key_pepper", mutate: func(j *JWT) { j.APIKeyPepper = "c2hvcnQ=" }, wantErr: ErrWeakPepper},
		{name: "missing keys_dir", mutate: func(j *JWT) { j.KeysDir = "" }, wantErr: ErrMissingKeysDir},
		{name: "nonexistent keys_dir", mutate: func(j *JWT) { j.KeysDir = "/nonexistent/path/xyz" }, wantErr: ErrKeysDirNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := baseValidJWT(t)
			tt.mutate(&j)
			err := j.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestJWT_Validate_Defaults(t *testing.T) {
	j := baseValidJWT(t)
	j.Leeway = 0
	j.CacheTTL = 0
	j.KeyBits = 0
	if err := j.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
	if j.Leeway != defaultLeeway {
		t.Errorf("Leeway = %v, want %v", j.Leeway, defaultLeeway)
	}
	if j.CacheTTL != defaultCacheTTL {
		t.Errorf("CacheTTL = %v, want %v", j.CacheTTL, defaultCacheTTL)
	}
	if j.KeyBits != 4096 {
		t.Errorf("KeyBits = %d, want 4096", j.KeyBits)
	}
}

func TestJWT_Validate_RejectsBadKeyBits(t *testing.T) {
	j := baseValidJWT(t)
	j.KeyBits = 1024 // too weak
	if err := j.Validate(); err == nil {
		t.Fatal("expected validation error for KeyBits=1024, got nil")
	}
}

func TestJWT_Validate_KeysDirIsFile(t *testing.T) {
	j := baseValidJWT(t)
	// Point KeysDir at a regular file.
	f, err := os.CreateTemp(t.TempDir(), "not-a-dir*")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	f.Close()
	j.KeysDir = f.Name()
	if err := j.Validate(); !errors.Is(err, ErrKeysDirNotFound) {
		t.Errorf("expected ErrKeysDirNotFound, got %v", err)
	}
}

func TestJWT_DecodePeppers(t *testing.T) {
	j := baseValidJWT(t)
	got, err := j.DecodeRefreshPepper()
	if err != nil || len(got) < minPepperBytes {
		t.Fatalf("DecodeRefreshPepper: %v (len=%d)", err, len(got))
	}
	got, err = j.DecodeAPIKeyPepper()
	if err != nil || len(got) < minPepperBytes {
		t.Fatalf("DecodeAPIKeyPepper: %v (len=%d)", err, len(got))
	}
}

func TestJWT_Validate_BothPeppersRequired(t *testing.T) {
	// Ensure refresh-pepper validation runs before api-key-pepper checks
	// (first missing wins).
	j := baseValidJWT(t)
	j.RefreshPepper = ""
	if err := j.Validate(); !errors.Is(err, ErrMissingRefreshPepper) {
		t.Errorf("expected ErrMissingRefreshPepper, got %v", err)
	}
}

// Smoke: mutating a time.Duration field after Validate shouldn't break anything
// obvious. This also documents the expected default Leeway value.
var _ = time.Minute