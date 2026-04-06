package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Compute generates a deterministic fingerprint from device-identifying
// request headers. The fingerprint is a hex-encoded SHA-256 of the
// concatenated values. Order is fixed; missing headers contribute empty
// strings (still deterministic).
func Compute(userAgent, acceptLanguage, secCHUA string) string {
	h := sha256.New()
	h.Write([]byte("fp:v1|"))
	h.Write([]byte(strings.TrimSpace(userAgent)))
	h.Write([]byte("|"))
	h.Write([]byte(strings.TrimSpace(acceptLanguage)))
	h.Write([]byte("|"))
	h.Write([]byte(strings.TrimSpace(secCHUA)))
	return hex.EncodeToString(h.Sum(nil))
}
