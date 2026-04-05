// Package keyring manages per-integration RSA key pairs stored as PEM files
// on disk. It auto-generates missing keys at boot, caches parsed *rsa.PrivateKey
// values in memory, and exposes them to the signing/verifying hot path.
package keyring

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// KeyPair bundles a loaded RSA key pair along with its kid + rotation metadata.
type KeyPair struct {
	PrivateKey   *rsa.PrivateKey
	PublicKeyPEM []byte    // same bytes as on disk, for JWKS
	KeyID        string    // "kid" header value
	RotatedAt    time.Time // when this key was installed
}

// Keyring is the per-process cache of loaded private key pairs, keyed by
// integration name.
type Keyring struct {
	mu   sync.RWMutex
	keys map[string]*KeyPair
	dir  string
	bits int
}

// New constructs an empty Keyring. Caller must subsequently call EnsureAndLoad
// (usually from bootstrap) for each integration name.
func New(dir string, bits int) (*Keyring, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("keyring: stat dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("keyring: path %q is not a directory", dir)
	}
	return &Keyring{
		keys: make(map[string]*KeyPair),
		dir:  dir,
		bits: bits,
	}, nil
}

// EnsureAndLoad guarantees that <dir>/<name>_private.pem exists (generating
// an RSA-<bits> pair when missing), then loads and caches it. Returns the
// loaded KeyPair. Thread-safe. keyID is supplied by the caller when known
// (e.g. from DB); if empty, a new UUID is generated during auto-gen.
func (k *Keyring) EnsureAndLoad(name, keyID string) (*KeyPair, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	privPath := k.privatePath(name)
	pubPath := k.publicPath(name)

	if _, err := os.Stat(privPath); err == nil {
		// Load existing.
		kp, err := loadPair(privPath, pubPath)
		if err != nil {
			return nil, err
		}
		if keyID != "" {
			kp.KeyID = keyID
		} else if kp.KeyID == "" {
			kp.KeyID = uuid.NewString()
		}
		k.keys[name] = kp
		return kp, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("keyring: stat %q: %w", privPath, err)
	}

	// Generate a new pair.
	kp, err := generatePair(k.bits)
	if err != nil {
		return nil, err
	}
	if keyID == "" {
		keyID = uuid.NewString()
	}
	kp.KeyID = keyID
	kp.RotatedAt = time.Now()

	if err := writePair(privPath, pubPath, kp); err != nil {
		return nil, err
	}
	k.keys[name] = kp
	return kp, nil
}

// Get returns the cached KeyPair for an integration. Returns false if not
// loaded yet. Thread-safe.
func (k *Keyring) Get(name string) (*KeyPair, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	kp, ok := k.keys[name]
	return kp, ok
}

// Rotate generates a new RSA pair for the integration, renames the current
// files to *_previous.pem, writes the new pair, updates the cache, and
// returns the new KeyPair (so caller can update the DB row).
func (k *Keyring) Rotate(name string) (*KeyPair, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	privPath := k.privatePath(name)
	pubPath := k.publicPath(name)
	prevPrivPath := k.previousPrivatePath(name)
	prevPubPath := k.previousPublicPath(name)

	// If a previous file already exists, remove it so Rename overwrites cleanly.
	if _, err := os.Stat(prevPrivPath); err == nil {
		if err := os.Remove(prevPrivPath); err != nil {
			return nil, fmt.Errorf("keyring: remove %q: %w", prevPrivPath, err)
		}
	}
	if _, err := os.Stat(prevPubPath); err == nil {
		if err := os.Remove(prevPubPath); err != nil {
			return nil, fmt.Errorf("keyring: remove %q: %w", prevPubPath, err)
		}
	}

	if _, err := os.Stat(privPath); err == nil {
		if err := os.Rename(privPath, prevPrivPath); err != nil {
			return nil, fmt.Errorf("keyring: rename %q: %w", privPath, err)
		}
	}
	if _, err := os.Stat(pubPath); err == nil {
		if err := os.Rename(pubPath, prevPubPath); err != nil {
			return nil, fmt.Errorf("keyring: rename %q: %w", pubPath, err)
		}
	}

	kp, err := generatePair(k.bits)
	if err != nil {
		return nil, err
	}
	kp.KeyID = uuid.NewString()
	kp.RotatedAt = time.Now()

	if err := writePair(privPath, pubPath, kp); err != nil {
		return nil, err
	}
	k.keys[name] = kp
	return kp, nil
}

// privatePath returns the path of the current private key PEM file.
func (k *Keyring) privatePath(name string) string {
	return filepath.Join(k.dir, name+"_private.pem")
}

// publicPath returns the path of the current public key PEM file.
func (k *Keyring) publicPath(name string) string {
	return filepath.Join(k.dir, name+"_public.pem")
}

// previousPrivatePath returns the path of the rotated-out private key PEM file.
func (k *Keyring) previousPrivatePath(name string) string {
	return filepath.Join(k.dir, name+"_private_previous.pem")
}

// previousPublicPath returns the path of the rotated-out public key PEM file.
func (k *Keyring) previousPublicPath(name string) string {
	return filepath.Join(k.dir, name+"_public_previous.pem")
}

// generatePair produces a fresh in-memory RSA key pair with PEM-encoded public
// bytes ready for disk persistence.
func generatePair(bits int) (*KeyPair, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("keyring: generate rsa key: %w", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("keyring: marshal public key: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return &KeyPair{
		PrivateKey:   priv,
		PublicKeyPEM: pubPEM,
	}, nil
}

// writePair persists a KeyPair to disk as PKCS#8/PKIX PEM with strict modes.
func writePair(privPath, pubPath string, kp *KeyPair) error {
	privDER, err := x509.MarshalPKCS8PrivateKey(kp.PrivateKey)
	if err != nil {
		return fmt.Errorf("keyring: marshal private key: %w", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})

	privFile, err := os.OpenFile(privPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("keyring: open %q: %w", privPath, err)
	}
	if _, err := privFile.Write(privPEM); err != nil {
		_ = privFile.Close()
		return fmt.Errorf("keyring: write %q: %w", privPath, err)
	}
	if err := privFile.Close(); err != nil {
		return fmt.Errorf("keyring: close %q: %w", privPath, err)
	}
	if err := os.Chmod(privPath, 0600); err != nil {
		return fmt.Errorf("keyring: chmod %q: %w", privPath, err)
	}

	pubFile, err := os.OpenFile(pubPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("keyring: open %q: %w", pubPath, err)
	}
	if _, err := pubFile.Write(kp.PublicKeyPEM); err != nil {
		_ = pubFile.Close()
		return fmt.Errorf("keyring: write %q: %w", pubPath, err)
	}
	if err := pubFile.Close(); err != nil {
		return fmt.Errorf("keyring: close %q: %w", pubPath, err)
	}
	if err := os.Chmod(pubPath, 0644); err != nil {
		return fmt.Errorf("keyring: chmod %q: %w", pubPath, err)
	}
	return nil
}

// loadPair reads a PEM-encoded PKCS#8 private key and its paired public PEM.
// RotatedAt is populated from the private file's modification time.
func loadPair(privPath, pubPath string) (*KeyPair, error) {
	privBytes, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("keyring: read %q: %w", privPath, err)
	}
	block, _ := pem.Decode(privBytes)
	if block == nil {
		return nil, fmt.Errorf("keyring: decode %q: no PEM data", privPath)
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("keyring: parse private key %q: %w", privPath, err)
	}
	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("keyring: %q is not an RSA key", privPath)
	}

	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, fmt.Errorf("keyring: read %q: %w", pubPath, err)
	}

	rotatedAt := time.Time{}
	if info, err := os.Stat(privPath); err == nil {
		rotatedAt = info.ModTime()
	}

	return &KeyPair{
		PrivateKey:   rsaKey,
		PublicKeyPEM: pubBytes,
		RotatedAt:    rotatedAt,
	}, nil
}
