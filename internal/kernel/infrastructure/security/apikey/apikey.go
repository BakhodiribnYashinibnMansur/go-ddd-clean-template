package apikey

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// Prefix constants for structured API key format.
const (
	PrefixAdmin  = "gct_adm_"
	PrefixClient = "gct_cli_"
	PrefixMobile = "gct_mob_"
	PrefixOther  = "gct_int_"
)

// Generate creates a new API key with the appropriate prefix.
// integrationName is used to pick the prefix: "gct-admin" -> "gct_adm_",
// "gct-client" -> "gct_cli_", "gct-mobile" -> "gct_mob_", else "gct_int_".
// The random part is 32 bytes base64url-encoded.
func Generate(integrationName string) (string, error) {
	prefix := prefixFor(integrationName)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("apikey.Generate: %w", err)
	}
	return prefix + base64.RawURLEncoding.EncodeToString(b), nil
}

// Mask returns a display-safe version: "gct_adm_****...last4".
// If the key doesn't match the expected format, returns "****...last4"
// (or "****" if too short).
func Mask(key string) string {
	if len(key) < 8 {
		return "****"
	}
	// Find prefix (first 8 chars like "gct_adm_")
	prefix := ""
	for _, p := range []string{PrefixAdmin, PrefixClient, PrefixMobile, PrefixOther} {
		if strings.HasPrefix(key, p) {
			prefix = p
			break
		}
	}
	last4 := key[len(key)-4:]
	return prefix + "****..." + last4
}

func prefixFor(name string) string {
	switch name {
	case "gct-admin":
		return PrefixAdmin
	case "gct-client":
		return PrefixClient
	case "gct-mobile":
		return PrefixMobile
	default:
		return PrefixOther
	}
}
