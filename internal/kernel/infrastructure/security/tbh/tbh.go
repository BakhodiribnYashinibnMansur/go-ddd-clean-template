package tbh

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
)

// Length is the byte length of the truncated HMAC (16 bytes = 128 bits).
const Length = 16

// Compute returns a base64url-encoded TBH for the given IP + User-Agent.
// The pepper is a server-side secret (same as the refresh pepper or a
// dedicated one). If ip or ua is empty, the function still computes a
// deterministic hash — callers decide whether to enforce.
func Compute(pepper []byte, ip, userAgent string) string {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte("tbh:v1|"))
	mac.Write([]byte(ip))
	mac.Write([]byte("|"))
	mac.Write([]byte(userAgent))
	full := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(full[:Length])
}

// Verify compares an incoming TBH against a recomputed one, constant-time.
func Verify(pepper []byte, ip, userAgent, storedTBH string) bool {
	computed := Compute(pepper, ip, userAgent)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(storedTBH)) == 1
}
