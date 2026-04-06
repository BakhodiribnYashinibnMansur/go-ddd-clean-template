package jwks

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"sync"
)

// Key is one entry in the JWKS response.
type Key struct {
	KTY string `json:"kty"`
	Use string `json:"use"`
	KID string `json:"kid"`
	ALG string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// KeySet is the full JWKS response payload.
type KeySet struct {
	Keys []Key `json:"keys"`
}

// Store holds the current set of public keys. Safe for concurrent use.
type Store struct {
	mu   sync.RWMutex
	keys []Key
}

// New creates an empty Store.
func New() *Store { return &Store{} }

// SetKeys replaces the current key set. Called on boot and after rotation.
func (s *Store) SetKeys(keys []Key) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys = keys
}

// KeySet returns the current JWKS response payload.
func (s *Store) KeySet() KeySet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.keys == nil {
		return KeySet{Keys: []Key{}}
	}

	out := make([]Key, len(s.keys))
	copy(out, s.keys)

	return KeySet{Keys: out}
}

// RSAPublicKeyToJWK converts an RSA public key + metadata to a JWK Key.
func RSAPublicKeyToJWK(pub *rsa.PublicKey, kid, alg, use string) Key {
	return Key{
		KTY: "RSA",
		Use: use,
		KID: kid,
		ALG: alg,
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(new(big.Int).SetInt64(int64(pub.E)).Bytes()),
	}
}
