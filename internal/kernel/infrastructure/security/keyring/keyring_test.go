package keyring

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

const testBits = 2048

func TestNew_RejectsNonexistentDir(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
	}{
		{name: "missing dir", path: filepath.Join(t.TempDir(), "does-not-exist")},
		{name: "path is file", path: ""},
	}
	// Build "path is file" case.
	fileDir := t.TempDir()
	filePath := filepath.Join(fileDir, "not-a-dir")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	tests[1].path = filePath

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := New(tc.path, testBits); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.path)
			}
		})
	}
}

func TestEnsureAndLoad_GeneratesThenReads(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	kp1, err := kr.EnsureAndLoad("svcA", "")
	if err != nil {
		t.Fatalf("first EnsureAndLoad: %v", err)
	}
	if kp1.KeyID == "" {
		t.Fatalf("expected generated keyID")
	}
	if kp1.PrivateKey == nil {
		t.Fatalf("expected private key")
	}
	if len(kp1.PublicKeyPEM) == 0 {
		t.Fatalf("expected public PEM")
	}

	// Verify files exist.
	privPath := filepath.Join(dir, "svcA_private.pem")
	pubPath := filepath.Join(dir, "svcA_public.pem")
	if _, err := os.Stat(privPath); err != nil {
		t.Fatalf("private missing: %v", err)
	}
	if _, err := os.Stat(pubPath); err != nil {
		t.Fatalf("public missing: %v", err)
	}

	// Second call: should load from disk; passing a keyID should be honoured.
	kr2, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New second: %v", err)
	}
	kp2, err := kr2.EnsureAndLoad("svcA", "explicit-kid")
	if err != nil {
		t.Fatalf("second EnsureAndLoad: %v", err)
	}
	if kp2.KeyID != "explicit-kid" {
		t.Fatalf("expected keyID explicit-kid, got %q", kp2.KeyID)
	}
	if kp2.PrivateKey.N.Cmp(kp1.PrivateKey.N) != 0 {
		t.Fatalf("loaded key modulus differs from generated")
	}
}

func TestEnsureAndLoad_FileModes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := kr.EnsureAndLoad("svcB", ""); err != nil {
		t.Fatalf("EnsureAndLoad: %v", err)
	}

	privInfo, err := os.Stat(filepath.Join(dir, "svcB_private.pem"))
	if err != nil {
		t.Fatalf("stat private: %v", err)
	}
	if got := privInfo.Mode().Perm(); got != 0600 {
		t.Fatalf("private mode = %o, want 0600", got)
	}
	pubInfo, err := os.Stat(filepath.Join(dir, "svcB_public.pem"))
	if err != nil {
		t.Fatalf("stat public: %v", err)
	}
	if got := pubInfo.Mode().Perm(); got != 0644 {
		t.Fatalf("public mode = %o, want 0644", got)
	}
}

func TestEnsureAndLoad_SignVerifyRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	kp, err := kr.EnsureAndLoad("svcSign", "")
	if err != nil {
		t.Fatalf("EnsureAndLoad: %v", err)
	}

	msg := []byte("hello keyring")
	digest := sha256.Sum256(msg)
	sig, err := rsa.SignPKCS1v15(rand.Reader, kp.PrivateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	// Parse public key from PEM and verify.
	block, _ := pem.Decode(kp.PublicKeyPEM)
	if block == nil {
		t.Fatalf("decode public PEM")
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("parse public: %v", err)
	}
	rsaPub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("not rsa public")
	}
	if err := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, digest[:], sig); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestGet_ReturnsCached(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := kr.Get("svcC"); ok {
		t.Fatalf("expected miss before load")
	}
	loaded, err := kr.EnsureAndLoad("svcC", "kid-c")
	if err != nil {
		t.Fatalf("EnsureAndLoad: %v", err)
	}
	got, ok := kr.Get("svcC")
	if !ok {
		t.Fatalf("expected hit after load")
	}
	if got != loaded {
		t.Fatalf("Get returned different pointer than EnsureAndLoad")
	}
	if got.KeyID != "kid-c" {
		t.Fatalf("keyID mismatch: %q", got.KeyID)
	}
}

func TestRotate_InstallsNewAndPreservesPrevious(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	orig, err := kr.EnsureAndLoad("svcR", "")
	if err != nil {
		t.Fatalf("EnsureAndLoad: %v", err)
	}
	origPublicPEM := make([]byte, len(orig.PublicKeyPEM))
	copy(origPublicPEM, orig.PublicKeyPEM)

	rotated, err := kr.Rotate("svcR")
	if err != nil {
		t.Fatalf("Rotate: %v", err)
	}
	if rotated.PrivateKey.N.Cmp(orig.PrivateKey.N) == 0 {
		t.Fatalf("rotated key modulus equals original")
	}
	if rotated.KeyID == orig.KeyID {
		t.Fatalf("rotated keyID equals original")
	}

	// Previous files contain the old public bytes.
	prevPubPath := filepath.Join(dir, "svcR_public_previous.pem")
	prevPubBytes, err := os.ReadFile(prevPubPath)
	if err != nil {
		t.Fatalf("read prev public: %v", err)
	}
	if !bytes.Equal(prevPubBytes, origPublicPEM) {
		t.Fatalf("previous public bytes do not match original")
	}
	if _, err := os.Stat(filepath.Join(dir, "svcR_private_previous.pem")); err != nil {
		t.Fatalf("prev private missing: %v", err)
	}

	// Cache reflects new pair.
	cached, ok := kr.Get("svcR")
	if !ok {
		t.Fatalf("expected cache hit after rotate")
	}
	if cached != rotated {
		t.Fatalf("cache does not hold rotated pair")
	}
}

func TestRotate_TwiceLeavesFourFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := kr.EnsureAndLoad("svcT", ""); err != nil {
		t.Fatalf("EnsureAndLoad: %v", err)
	}
	if _, err := kr.Rotate("svcT"); err != nil {
		t.Fatalf("rotate 1: %v", err)
	}
	second, err := kr.Rotate("svcT")
	if err != nil {
		t.Fatalf("rotate 2: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	// Expect exactly 4 files: current priv+pub, previous priv+pub.
	// (The oldest generation from the first rotation should have been
	// overwritten by the second rotation.)
	if len(entries) != 4 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("expected 4 files, got %d: %v", len(entries), names)
	}

	wantNames := map[string]bool{
		"svcT_private.pem":          false,
		"svcT_public.pem":           false,
		"svcT_private_previous.pem": false,
		"svcT_public_previous.pem":  false,
	}
	for _, e := range entries {
		if _, ok := wantNames[e.Name()]; !ok {
			t.Fatalf("unexpected file %q", e.Name())
		}
		wantNames[e.Name()] = true
	}
	for n, seen := range wantNames {
		if !seen {
			t.Fatalf("missing expected file %q", n)
		}
	}

	// Current private on disk should match second-rotation key.
	curPrivBytes, err := os.ReadFile(filepath.Join(dir, "svcT_private.pem"))
	if err != nil {
		t.Fatalf("read current private: %v", err)
	}
	block, _ := pem.Decode(curPrivBytes)
	if block == nil {
		t.Fatalf("decode current private")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("parse current private: %v", err)
	}
	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		t.Fatalf("current private not rsa")
	}
	if rsaKey.N.Cmp(second.PrivateKey.N) != 0 {
		t.Fatalf("current private on disk does not match last rotation")
	}
}
