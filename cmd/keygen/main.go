// keygen generates cryptographic material required by the multi-integration
// JWT architecture:
//
//   - one RSA key pair per integration, stored as PEM files on disk via the
//     keyring package (<dir>/<name>_private.pem, <name>_public.pem);
//   - a base64-encoded refresh-token pepper (48 raw bytes) for HMAC;
//   - a base64-encoded API-key pepper (48 raw bytes) for HMAC;
//   - a 32-byte random API key per integration, printed as
//     JWT_<UPPER_SNAKE_NAME>_API_KEY=<base64url>.
//
// Usage:
//
//	go run ./cmd/keygen                          # default: all integrations
//	go run ./cmd/keygen -out-dir config/keys
//	go run ./cmd/keygen -integrations a,b,c
//	go run ./cmd/keygen -force                   # rotate existing PEMs
//	go run ./cmd/keygen -pepper-only             # only print peppers
//	go run ./cmd/keygen -api-keys-only           # only print API keys + peppers
//	go run ./cmd/keygen -quiet                   # suppress stderr status
package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gct/internal/kernel/infrastructure/security/keyring"
)

const (
	defaultOutDir       = "config/keys"
	defaultBits         = 4096
	defaultIntegrations = "gct-admin,gct-client,gct-mobile"
	pepperSize          = 48 // 48 raw bytes -> 64 base64 chars
	apiKeySize          = 32 // 32 raw bytes -> 43 base64url chars
)

func main() {
	outDir := flag.String("out-dir", defaultOutDir, "directory to write key files into")
	bits := flag.Int("bits", defaultBits, "RSA key size in bits (2048 | 3072 | 4096)")
	integrationsFlag := flag.String("integrations", defaultIntegrations, "comma-separated list of integration names")
	force := flag.Bool("force", false, "overwrite existing PEM files")
	pepperOnly := flag.Bool("pepper-only", false, "only generate and print fresh refresh + api-key peppers")
	apiKeysOnly := flag.Bool("api-keys-only", false, "only print API keys + peppers, no PEM generation")
	quiet := flag.Bool("quiet", false, "suppress stderr status messages")
	flag.Parse()

	integrations := parseIntegrations(*integrationsFlag)

	if *pepperOnly {
		if err := generatePeppers(); err != nil {
			fatal(err)
		}
		return
	}

	if *apiKeysOnly {
		if err := generatePeppers(); err != nil {
			fatal(err)
		}
		if err := generateAPIKeys(integrations); err != nil {
			fatal(err)
		}
		return
	}

	if err := validateBits(*bits); err != nil {
		fatal(err)
	}
	if len(integrations) == 0 {
		fatal(errors.New("no integrations specified"))
	}

	if err := os.MkdirAll(*outDir, 0o750); err != nil {
		fatal(fmt.Errorf("create out-dir: %w", err))
	}

	if err := generatePEMsForIntegrations(*outDir, *bits, integrations, *force, *quiet); err != nil {
		fatal(err)
	}

	if err := generatePeppers(); err != nil {
		fatal(err)
	}
	if err := generateAPIKeys(integrations); err != nil {
		fatal(err)
	}
}

// parseIntegrations splits a comma-separated list, trimming whitespace and
// dropping empty entries.
func parseIntegrations(s string) []string {
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		name := strings.TrimSpace(r)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

// generatePEMsForIntegrations calls EnsureAndLoad for every integration,
// honouring -force by deleting existing PEM files beforehand.
func generatePEMsForIntegrations(dir string, bits int, integrations []string, force, quiet bool) error {
	kr, err := keyring.New(dir, bits)
	if err != nil {
		return fmt.Errorf("init keyring: %w", err)
	}

	for _, name := range integrations {
		privPath := filepath.Join(dir, name+"_private.pem")
		pubPath := filepath.Join(dir, name+"_public.pem")

		if force {
			if err := removeIfExists(privPath); err != nil {
				return err
			}
			if err := removeIfExists(pubPath); err != nil {
				return err
			}
		}

		// Detect whether this call will generate fresh files.
		willGenerate := false
		if _, err := os.Stat(privPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				willGenerate = true
			} else {
				return fmt.Errorf("stat %q: %w", privPath, err)
			}
		}

		if !quiet {
			if willGenerate {
				fmt.Fprintf(os.Stderr, "generating RSA-%d key pair for %s...\n", bits, name)
			} else {
				fmt.Fprintf(os.Stderr, "loading existing key pair for %s...\n", name)
			}
		}

		if _, err := kr.EnsureAndLoad(name, ""); err != nil {
			return fmt.Errorf("ensure+load %s: %w", name, err)
		}

		if !quiet && willGenerate {
			fmt.Fprintf(os.Stderr, "wrote %s (mode 0600)\n", privPath)
			fmt.Fprintf(os.Stderr, "wrote %s (mode 0644)\n", pubPath)
		}
	}
	return nil
}

// generatePeppers emits the refresh pepper + api-key pepper to stdout.
func generatePeppers() error {
	refresh, err := randomB64std(pepperSize)
	if err != nil {
		return fmt.Errorf("refresh pepper: %w", err)
	}
	apiKeyPepper, err := randomB64std(pepperSize)
	if err != nil {
		return fmt.Errorf("api-key pepper: %w", err)
	}
	fmt.Fprintf(os.Stdout, "JWT_REFRESH_PEPPER=%s\n", refresh)
	fmt.Fprintf(os.Stdout, "JWT_API_KEY_PEPPER=%s\n", apiKeyPepper)
	return nil
}

// generateAPIKeys emits one API key per integration to stdout.
func generateAPIKeys(integrations []string) error {
	for _, name := range integrations {
		key, err := randomB64url(apiKeySize)
		if err != nil {
			return fmt.Errorf("api key %s: %w", name, err)
		}
		fmt.Fprintf(os.Stdout, "JWT_%s_API_KEY=%s\n", envVarName(name), key)
	}
	return nil
}

// envVarName converts an integration identifier like "gct-admin" into the
// environment-variable fragment "GCT_ADMIN".
func envVarName(integration string) string {
	return strings.ToUpper(strings.ReplaceAll(integration, "-", "_"))
}

// randomB64url returns n random bytes encoded using base64 URL (no padding).
func randomB64url(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("rand read: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// randomB64std returns n random bytes encoded using base64 std (no padding).
func randomB64std(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("rand read: %w", err)
	}
	return base64.RawStdEncoding.EncodeToString(buf), nil
}

// removeIfExists deletes a file but swallows os.ErrNotExist.
func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove %q: %w", path, err)
	}
	return nil
}

func validateBits(bits int) error {
	switch bits {
	case 2048, 3072, 4096:
		return nil
	default:
		return fmt.Errorf("invalid -bits %d: allowed values are 2048, 3072, 4096", bits)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "keygen: %v\n", err)
	os.Exit(1)
}
