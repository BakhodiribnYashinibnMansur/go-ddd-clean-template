package jwks_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"gct/internal/kernel/infrastructure/security/jwks"
)

func TestRSAPublicKeyToJWK(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	key := jwks.RSAPublicKeyToJWK(&priv.PublicKey, "test-kid", "RS256", "sig")

	if key.KTY != "RSA" {
		t.Errorf("kty = %q, want RSA", key.KTY)
	}
	if key.Use != "sig" {
		t.Errorf("use = %q, want sig", key.Use)
	}
	if key.ALG != "RS256" {
		t.Errorf("alg = %q, want RS256", key.ALG)
	}
	if key.KID != "test-kid" {
		t.Errorf("kid = %q, want test-kid", key.KID)
	}
	if key.N == "" {
		t.Error("n is empty")
	}
	if key.E == "" {
		t.Error("e is empty")
	}
}

func TestSetKeysAndKeySetRoundTrip(t *testing.T) {
	store := jwks.New()

	keys := []jwks.Key{
		{KTY: "RSA", Use: "sig", KID: "k1", ALG: "RS256", N: "abc", E: "def"},
		{KTY: "RSA", Use: "sig", KID: "k2", ALG: "RS256", N: "ghi", E: "jkl"},
	}
	store.SetKeys(keys)

	ks := store.KeySet()
	if len(ks.Keys) != 2 {
		t.Fatalf("got %d keys, want 2", len(ks.Keys))
	}
	if ks.Keys[0].KID != "k1" || ks.Keys[1].KID != "k2" {
		t.Errorf("unexpected kids: %v", ks.Keys)
	}
}

func TestEmptyStoreReturnsEmptyKeys(t *testing.T) {
	store := jwks.New()
	ks := store.KeySet()

	if ks.Keys == nil {
		t.Fatal("keys should not be nil")
	}
	if len(ks.Keys) != 0 {
		t.Errorf("got %d keys, want 0", len(ks.Keys))
	}
}
