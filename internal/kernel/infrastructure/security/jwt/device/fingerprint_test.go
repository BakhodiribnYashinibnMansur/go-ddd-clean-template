package device

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name      string
		headers   map[string]string
		remoteAddr string
	}{
		{
			name: "chrome_desktop",
			headers: map[string]string{
				"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
				"Accept-Language": "en-US,en;q=0.9",
			},
			remoteAddr: "192.168.1.100:12345",
		},
		{
			name: "mobile_safari",
			headers: map[string]string{
				"User-Agent":      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
				"Accept-Language": "en-GB",
			},
			remoteAddr: "10.0.0.5:8080",
		},
		{
			name:       "empty_headers",
			headers:    map[string]string{},
			remoteAddr: "127.0.0.1:9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://example.com", nil)
			require.NoError(t, err)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tt.remoteAddr

			fp := Generate(req)
			assert.NotEmpty(t, fp.Hash)
			assert.NotEmpty(t, fp.IP)
		})
	}
}

func TestFingerprint_Verify(t *testing.T) {
	req1, _ := http.NewRequest("GET", "http://example.com", nil)
	req1.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0.0.0")
	req1.RemoteAddr = "192.168.1.100:12345"

	fp1 := Generate(req1)

	t.Run("same_request_matches", func(t *testing.T) {
		req2, _ := http.NewRequest("GET", "http://example.com", nil)
		req2.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0.0.0")
		req2.RemoteAddr = "192.168.1.100:12345"
		fp2 := Generate(req2)

		assert.True(t, fp1.Verify(fp2))
	})

	t.Run("same_subnet_matches", func(t *testing.T) {
		req2, _ := http.NewRequest("GET", "http://example.com", nil)
		req2.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0.0.0")
		req2.RemoteAddr = "192.168.1.200:12345" // same subnet, different host
		fp2 := Generate(req2)

		assert.True(t, fp1.Verify(fp2))
	})

	t.Run("different_browser_fails", func(t *testing.T) {
		req2, _ := http.NewRequest("GET", "http://example.com", nil)
		req2.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/604.1")
		req2.RemoteAddr = "192.168.1.100:12345"
		fp2 := Generate(req2)

		assert.False(t, fp1.Verify(fp2))
	})
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "x_forwarded_for",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50, 70.41.3.18"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.50",
		},
		{
			name:       "x_real_ip",
			headers:    map[string]string{"X-Real-IP": "203.0.113.50"},
			remoteAddr: "127.0.0.1:8080",
			expected:   "203.0.113.50",
		},
		{
			name:       "fallback_remote_addr",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1:9999",
			expected:   "10.0.0.1:9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tt.remoteAddr

			ip := getClientIP(req)
			assert.Equal(t, tt.expected, ip)
		})
	}
}

func TestVerifyIP(t *testing.T) {
	tests := []struct {
		name    string
		stored  string
		current string
		want    bool
	}{
		{
			name:    "exact_match",
			stored:  "192.168.1.100",
			current: "192.168.1.100",
			want:    true,
		},
		{
			name:    "same_subnet",
			stored:  "192.168.1.100",
			current: "192.168.1.200",
			want:    true,
		},
		{
			name:    "different_subnet",
			stored:  "192.168.1.100",
			current: "10.0.0.100",
			want:    false,
		},
		{
			name:    "ipv6_exact",
			stored:  "::1",
			current: "::1",
			want:    true,
		},
		{
			name:    "ipv6_different",
			stored:  "::1",
			current: "::2",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifyIP(tt.stored, tt.current)
			assert.Equal(t, tt.want, result)
		})
	}
}
