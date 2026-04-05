package useragent

import (
	"github.com/mileusna/useragent"
)

const (
	DeviceTypeMobile  = "MOBILE"
	DeviceTypeTablet  = "TABLET"
	DeviceTypeDesktop = "DESKTOP"
	DeviceTypeBot     = "BOT"
	DeviceTypeUnknown = ""
)

type UserAgent struct {
	URL            string
	String         string
	Browser        string
	BrowserVersion string
	OS             string
	OSVersion      string
	Device         string
	DeviceType     string
}

func getDeviceType(ua *useragent.UserAgent) string {
	// Check in priority order
	if ua.Bot {
		return DeviceTypeBot
	}
	if ua.Mobile {
		return DeviceTypeMobile
	}
	if ua.Tablet {
		return DeviceTypeTablet
	}
	// Default to desktop for all other cases (browsers on PC/Mac/Linux)
	return DeviceTypeDesktop
}

func ParseUserAgent(userAgentTxt string) *UserAgent {
	userAgent := useragent.Parse(userAgentTxt)
	ua := &UserAgent{
		URL:            userAgent.URL,
		String:         userAgent.String,
		Browser:        userAgent.Name,
		BrowserVersion: userAgent.Version,
		OS:             userAgent.OS,
		OSVersion:      userAgent.OSVersion,
		Device:         getDevice(&userAgent),
		DeviceType:     getDeviceType(&userAgent),
	}

	return ua
}

func getDevice(ua *useragent.UserAgent) string {
	deviceName := ua.Device
	if deviceName == "" {
		deviceName = ua.OS + " " + ua.Name
	}
	return deviceName
}
