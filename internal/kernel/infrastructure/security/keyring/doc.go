// Package keyring provides per-integration RSA key pair management with
// disk persistence, in-memory caching, and scheduled rotation support.
//
// # Purpose
//
// Each third-party integration that requires asymmetric signing (JWT bearer
// assertions, webhook signatures, etc.) owns a dedicated RSA key pair. This
// package centralises the lifecycle of those pairs so the rest of the
// application can focus on signing and verifying.
//
// # Lifecycle
//
//  1. Call New at boot with the directory holding PEM files and the RSA bit
//     size to use when generating new keys.
//  2. For each active integration, call EnsureAndLoad(name, keyID). If no
//     private key exists on disk, a fresh RSA pair is generated, written with
//     strict file permissions, and cached. Otherwise the existing files are
//     loaded and cached.
//  3. The signing/verifying hot path calls Get(name) to retrieve the cached
//     *KeyPair, which holds the parsed *rsa.PrivateKey.
//  4. A scheduled task (typically daily) calls Rotate(name) to generate a new
//     pair; the previous files are preserved as *_previous.pem so verifiers
//     can still validate signatures produced just before the rotation.
//
// # File layout
//
// Given a base directory dir and integration name foo, the package manages
// these files:
//
//	<dir>/foo_private.pem           // current private key, mode 0600, PKCS#8
//	<dir>/foo_public.pem            // current public key,  mode 0644, PKIX
//	<dir>/foo_private_previous.pem  // previous private key (after first Rotate)
//	<dir>/foo_public_previous.pem   // previous public key  (after first Rotate)
//
// # Thread-safety
//
// A Keyring is safe for concurrent use by multiple goroutines. Reads via Get
// take a read lock; EnsureAndLoad and Rotate take a write lock.
package keyring
