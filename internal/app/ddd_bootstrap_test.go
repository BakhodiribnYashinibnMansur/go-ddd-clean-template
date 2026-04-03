package app

import (
	"testing"

	"gct/internal/announcement"
	"gct/internal/audit"
	"gct/internal/authz"
	"gct/internal/dashboard"
	"gct/internal/dataexport"
	"gct/internal/errorcode"
	"gct/internal/featureflag"
	"gct/internal/file"
	"gct/internal/integration"
	"gct/internal/iprule"
	"gct/internal/metric"
	"gct/internal/notification"
	"gct/internal/ratelimit"
	"gct/internal/session"
	"gct/internal/sitesetting"
	"gct/internal/systemerror"
	"gct/internal/translation"
	"gct/internal/user"
	"gct/internal/usersetting"
)

func TestDDDBoundedContexts_ZeroValue(t *testing.T) {
	var bc DDDBoundedContexts

	if bc.User != nil {
		t.Error("expected User to be nil")
	}
	if bc.Authz != nil {
		t.Error("expected Authz to be nil")
	}
	if bc.Session != nil {
		t.Error("expected Session to be nil")
	}
	if bc.Audit != nil {
		t.Error("expected Audit to be nil")
	}
	if bc.Dashboard != nil {
		t.Error("expected Dashboard to be nil")
	}
	if bc.SystemError != nil {
		t.Error("expected SystemError to be nil")
	}
	if bc.Metric != nil {
		t.Error("expected Metric to be nil")
	}
	if bc.FeatureFlag != nil {
		t.Error("expected FeatureFlag to be nil")
	}
	if bc.Integration != nil {
		t.Error("expected Integration to be nil")
	}
	if bc.Notification != nil {
		t.Error("expected Notification to be nil")
	}
	if bc.Announcement != nil {
		t.Error("expected Announcement to be nil")
	}
	if bc.Translation != nil {
		t.Error("expected Translation to be nil")
	}
	if bc.SiteSetting != nil {
		t.Error("expected SiteSetting to be nil")
	}
	if bc.RateLimit != nil {
		t.Error("expected RateLimit to be nil")
	}
	if bc.IPRule != nil {
		t.Error("expected IPRule to be nil")
	}
	if bc.DataExport != nil {
		t.Error("expected DataExport to be nil")
	}
	if bc.File != nil {
		t.Error("expected File to be nil")
	}
	if bc.UserSetting != nil {
		t.Error("expected UserSetting to be nil")
	}
	if bc.ErrorCode != nil {
		t.Error("expected ErrorCode to be nil")
	}
}

func TestDDDBoundedContexts_AllFieldsAssignable(t *testing.T) {
	userBC := &user.BoundedContext{}
	authzBC := &authz.BoundedContext{}
	sessionBC := &session.BoundedContext{}
	auditBC := &audit.BoundedContext{}
	dashboardBC := &dashboard.BoundedContext{}
	systemErrorBC := &systemerror.BoundedContext{}
	metricBC := &metric.BoundedContext{}
	featureFlagBC := &featureflag.BoundedContext{}
	integrationBC := &integration.BoundedContext{}
	notificationBC := &notification.BoundedContext{}
	announcementBC := &announcement.BoundedContext{}
	translationBC := &translation.BoundedContext{}
	siteSettingBC := &sitesetting.BoundedContext{}
	rateLimitBC := &ratelimit.BoundedContext{}
	ipRuleBC := &iprule.BoundedContext{}
	dataExportBC := &dataexport.BoundedContext{}
	fileBC := &file.BoundedContext{}
	userSettingBC := &usersetting.BoundedContext{}
	errorCodeBC := &errorcode.BoundedContext{}

	bc := DDDBoundedContexts{
		User:         userBC,
		Authz:        authzBC,
		Session:      sessionBC,
		Audit:        auditBC,
		Dashboard:    dashboardBC,
		SystemError:  systemErrorBC,
		Metric:       metricBC,
		FeatureFlag:  featureFlagBC,
		Integration:  integrationBC,
		Notification: notificationBC,
		Announcement: announcementBC,
		Translation:  translationBC,
		SiteSetting:  siteSettingBC,
		RateLimit:    rateLimitBC,
		IPRule:       ipRuleBC,
		DataExport:   dataExportBC,
		File:         fileBC,
		UserSetting:  userSettingBC,
		ErrorCode:    errorCodeBC,
	}

	if bc.User != userBC {
		t.Error("User field mismatch")
	}
	if bc.Authz != authzBC {
		t.Error("Authz field mismatch")
	}
	if bc.Session != sessionBC {
		t.Error("Session field mismatch")
	}
	if bc.Audit != auditBC {
		t.Error("Audit field mismatch")
	}
	if bc.Dashboard != dashboardBC {
		t.Error("Dashboard field mismatch")
	}
	if bc.SystemError != systemErrorBC {
		t.Error("SystemError field mismatch")
	}
	if bc.Metric != metricBC {
		t.Error("Metric field mismatch")
	}
	if bc.FeatureFlag != featureFlagBC {
		t.Error("FeatureFlag field mismatch")
	}
	if bc.Integration != integrationBC {
		t.Error("Integration field mismatch")
	}
	if bc.Notification != notificationBC {
		t.Error("Notification field mismatch")
	}
	if bc.Announcement != announcementBC {
		t.Error("Announcement field mismatch")
	}
	if bc.Translation != translationBC {
		t.Error("Translation field mismatch")
	}
	if bc.SiteSetting != siteSettingBC {
		t.Error("SiteSetting field mismatch")
	}
	if bc.RateLimit != rateLimitBC {
		t.Error("RateLimit field mismatch")
	}
	if bc.IPRule != ipRuleBC {
		t.Error("IPRule field mismatch")
	}
	if bc.DataExport != dataExportBC {
		t.Error("DataExport field mismatch")
	}
	if bc.File != fileBC {
		t.Error("File field mismatch")
	}
	if bc.UserSetting != userSettingBC {
		t.Error("UserSetting field mismatch")
	}
	if bc.ErrorCode != errorCodeBC {
		t.Error("ErrorCode field mismatch")
	}
}
