package config

// Seeder - data seeding configuration.
type Seeder struct {
	Enabled bool `env:"SEEDER_ENABLED" envDefault:"false"`
	// Number of fake users to create
	UsersCount int `env:"SEEDER_USERS_COUNT" envDefault:"50"`
	// Number of fake roles to create
	RolesCount int `env:"SEEDER_ROLES_COUNT" envDefault:"10"`
	// Number of fake permissions to create
	PermissionsCount int `env:"SEEDER_PERMISSIONS_COUNT" envDefault:"20"`
	// Number of fake policies to create
	PoliciesCount int `env:"SEEDER_POLICIES_COUNT" envDefault:"20"`
	// Number of fake announcements to create
	AnnouncementsCount int `env:"SEEDER_ANNOUNCEMENTS_COUNT" envDefault:"10"`
	// Number of fake notifications to create
	NotificationsCount int `env:"SEEDER_NOTIFICATIONS_COUNT" envDefault:"30"`
	// Number of fake feature flags to create
	FeatureFlagsCount int `env:"SEEDER_FEATURE_FLAGS_COUNT" envDefault:"15"`
	// Number of fake integrations to create
	IntegrationsCount int `env:"SEEDER_INTEGRATIONS_COUNT" envDefault:"5"`
	// Number of fake translations to create
	TranslationsCount int `env:"SEEDER_TRANSLATIONS_COUNT" envDefault:"50"`
	// Number of fake file metadata records to create
	FileMetadataCount int `env:"SEEDER_FILE_METADATA_COUNT" envDefault:"20"`
	// Number of fake site settings to create
	SiteSettingsCount int `env:"SEEDER_SITE_SETTINGS_COUNT" envDefault:"15"`
	// Number of fake error codes to create
	ErrorCodesCount int `env:"SEEDER_ERROR_CODES_COUNT" envDefault:"20"`
	// Number of fake IP rules to create
	IPRulesCount int `env:"SEEDER_IP_RULES_COUNT" envDefault:"10"`
	// Number of fake rate limits to create
	RateLimitsCount int `env:"SEEDER_RATE_LIMITS_COUNT" envDefault:"8"`
	// Number of fake audit logs to create
	AuditLogsCount int `env:"SEEDER_AUDIT_LOGS_COUNT" envDefault:"50"`
	// Number of fake function metrics to create
	FunctionMetricsCount int `env:"SEEDER_FUNCTION_METRICS_COUNT" envDefault:"30"`
	// Seed value for reproducible random data (0 = random seed)
	Seed int64 `env:"SEEDER_SEED" envDefault:"0"`
	// Clear existing data before seeding
	ClearData bool `env:"SEEDER_CLEAR_DATA" envDefault:"false"`
}

// IsEnabled returns true if seeder is enabled.
func (s *Seeder) IsEnabled() bool {
	return s.Enabled
}

// ShouldClearData returns true if existing data should be cleared before seeding.
func (s *Seeder) ShouldClearData() bool {
	return s.ClearData
}
