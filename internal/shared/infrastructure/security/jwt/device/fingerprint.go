package device

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"gct/internal/shared/infrastructure/useragent"
)

// Fingerprint represents a device fingerprint
type Fingerprint struct {
	Hash            string               `json:"hash"`
	UserAgent       string               `json:"user_agent"`
	IP              string               `json:"ip"`
	Language        string               `json:"language"`
	Platform        string               `json:"platform"`
	ParsedUserAgent *useragent.UserAgent `json:"parsed_user_agent"`
	DeviceType      string               `json:"device_type"`
	Browser         string               `json:"browser"`
	OS              string               `json:"os"`
}

// Generate creates a device fingerprint from HTTP request
func Generate(r *http.Request) *Fingerprint {
	// Collect various device attributes
	userAgentStr := r.Header.Get("User-Agent")
	ip := getClientIP(r)
	language := r.Header.Get("Accept-Language")

	// Parse user agent using the useragent package
	parsedUA := useragent.ParseUserAgent(userAgentStr)

	// Create a unique hash from these attributes
	fingerprintData := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		parsedUA.Browser,
		parsedUA.OS,
		parsedUA.DeviceType,
		ip,
		language,
		parsedUA.Device)
	hash := sha256.Sum256([]byte(fingerprintData))

	return &Fingerprint{
		Hash:            hex.EncodeToString(hash[:]),
		UserAgent:       userAgentStr,
		IP:              ip,
		Language:        language,
		Platform:        parsedUA.OS,
		ParsedUserAgent: parsedUA,
		DeviceType:      parsedUA.DeviceType,
		Browser:         parsedUA.Browser,
		OS:              parsedUA.OS,
	}
}

// Verify checks if the current request matches the stored fingerprint
func (f *Fingerprint) Verify(current *Fingerprint) bool {
	// Allow some flexibility for IP changes (e.g., mobile networks)
	ipMatch := verifyIP(f.IP, current.IP)

	// Compare parsed user agent details for more accurate matching
	browserMatch := f.Browser == current.Browser
	osMatch := f.OS == current.OS
	deviceTypeMatch := f.DeviceType == current.DeviceType

	return ipMatch && browserMatch && osMatch && deviceTypeMatch
}

// getClientIP extracts the real client IP
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// verifyIP allows for some IP flexibility (e.g., mobile networks)
func verifyIP(stored, current string) bool {
	// Exact match
	if stored == current {
		return true
	}

	// Allow same subnet for IPv4
	if strings.Contains(stored, ".") && strings.Contains(current, ".") {
		storedParts := strings.Split(stored, ".")
		currentParts := strings.Split(current, ".")

		if len(storedParts) == 4 && len(currentParts) == 4 {
			// Match first 3 octets (subnet)
			return storedParts[0] == currentParts[0] &&
				storedParts[1] == currentParts[1] &&
				storedParts[2] == currentParts[2]
		}
	}

	return false
}
