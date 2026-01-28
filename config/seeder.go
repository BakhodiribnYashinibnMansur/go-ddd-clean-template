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
