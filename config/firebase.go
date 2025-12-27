package config

// Firebase configuration -.
type Firebase struct {
	Mobile FirebaseConf `envPrefix:"MOBILE_"`
	Web    FirebaseConf `envPrefix:"WEB_"`
}

// FirebaseConf -.
type FirebaseConf struct {
	Type                    string `env:"TYPE" json:"type"`
	ProjectID               string `env:"PROJECT_ID" json:"project_id"`
	PrivateKeyID            string `env:"PRIVATE_KEY_ID" json:"private_key_id"`
	PrivateKey              string `env:"PRIVATE_KEY" json:"private_key"`
	ClientEmail             string `env:"CLIENT_EMAIL" json:"client_email"`
	ClientID                string `env:"ClientID" json:"client_id"`
	AuthURI                 string `env:"AUTH_URI" json:"auth_uri"`
	TokenURI                string `env:"TOKEN_URI" json:"token_uri"`
	AuthProviderX509CertURL string `env:"AUTH_PROVIDER_X509_CERT_URL" json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `env:"CLIENT_X509_CERT_URL" json:"client_x509_cert_url"`
	UniverseDomain          string `env:"UNIVERSE_DOMAIN" json:"universe_domain"`
}
