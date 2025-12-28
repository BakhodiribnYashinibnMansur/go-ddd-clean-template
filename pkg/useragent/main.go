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
	deviceType := DeviceTypeUnknown
	switch {
	case ua.Mobile:
		deviceType = DeviceTypeMobile
	case ua.Tablet:
		deviceType = DeviceTypeTablet
	case ua.Desktop:
		deviceType = DeviceTypeDesktop
	case ua.Bot:
		deviceType = DeviceTypeBot
	default:
		deviceType = DeviceTypeUnknown
	}
	return deviceType
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
