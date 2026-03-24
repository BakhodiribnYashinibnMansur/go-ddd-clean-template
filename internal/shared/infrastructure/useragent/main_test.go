package useragent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		name           string
		ua             string
		wantDeviceType string
		wantBrowser    string
		wantOSContains string
	}{
		{
			name:           "Chrome on Windows desktop",
			ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantDeviceType: DeviceTypeDesktop,
			wantBrowser:    "Chrome",
			wantOSContains: "Windows",
		},
		{
			name:           "Safari on iPhone mobile",
			ua:             "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			wantDeviceType: DeviceTypeMobile,
			wantBrowser:    "Safari",
			wantOSContains: "iOS",
		},
		{
			name:           "Chrome on iPad tablet",
			ua:             "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/120.0.6099.119 Mobile/15E148 Safari/604.1",
			wantDeviceType: DeviceTypeTablet,
			wantOSContains: "iOS",
		},
		{
			name:           "Googlebot",
			ua:             "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			wantDeviceType: DeviceTypeBot,
		},
		{
			name:           "empty string defaults to desktop",
			ua:             "",
			wantDeviceType: DeviceTypeDesktop,
		},
		{
			name:           "Firefox on Linux",
			ua:             "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0",
			wantDeviceType: DeviceTypeDesktop,
			wantBrowser:    "Firefox",
			wantOSContains: "Linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseUserAgent(tt.ua)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantDeviceType, result.DeviceType)
			if tt.wantBrowser != "" {
				assert.Equal(t, tt.wantBrowser, result.Browser)
			}
			if tt.wantOSContains != "" {
				assert.Contains(t, result.OS, tt.wantOSContains)
			}
		})
	}
}

func TestGetDeviceTypePriority(t *testing.T) {
	// Bot takes priority over mobile/tablet
	t.Run("bot beats mobile", func(t *testing.T) {
		ua := ParseUserAgent("Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.199 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
		require.NotNil(t, ua)
		assert.Equal(t, DeviceTypeBot, ua.DeviceType)
	})
}

func TestGetDevice(t *testing.T) {
	t.Run("known device name", func(t *testing.T) {
		ua := ParseUserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1")
		require.NotNil(t, ua)
		assert.NotEmpty(t, ua.Device)
	})

	t.Run("fallback to OS + browser", func(t *testing.T) {
		ua := ParseUserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0")
		require.NotNil(t, ua)
		assert.NotEmpty(t, ua.Device)
	})
}

func TestParseUserAgentFieldsPopulated(t *testing.T) {
	ua := ParseUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	require.NotNil(t, ua)
	assert.Equal(t, "Chrome", ua.Browser)
	assert.NotEmpty(t, ua.BrowserVersion)
	assert.Contains(t, ua.OS, "macOS")
	assert.NotEmpty(t, ua.OSVersion)
	assert.Equal(t, DeviceTypeDesktop, ua.DeviceType)
}
