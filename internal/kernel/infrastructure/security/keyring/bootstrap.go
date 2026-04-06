package keyring

import (
	"context"
	"fmt"

	"gct/internal/kernel/infrastructure/logger"
)

// BootstrapIntegration carries the minimal info needed to decide whether an
// integration's RSA key pair must be generated.
type BootstrapIntegration struct {
	Name  string
	KeyID string // current kid from DB; empty means generation is needed.
}

// UpdateKeyFn is called after a key is generated or rotated to persist the
// public key and kid back to the database. The caller provides a closure that
// wraps the repository call.
type UpdateKeyFn func(ctx context.Context, name, publicKeyPEM, keyID string) error

// Bootstrap iterates active integrations and ensures each has an RSA key pair
// on disk. For any integration missing files it generates new keys via the
// Keyring and calls updateFn to persist the public key + kid back to the DB.
//
// Errors for individual integrations are logged but do not abort the loop
// (partial success is better than total failure).
func Bootstrap(ctx context.Context, kr *Keyring, integrations []BootstrapIntegration, updateFn UpdateKeyFn, l logger.Log) error {
	var firstErr error
	for _, integ := range integrations {
		kp, err := kr.EnsureAndLoad(integ.Name, integ.KeyID)
		if err != nil {
			l.Errorc(ctx, "keyring bootstrap: failed to ensure key pair",
				"integration", integ.Name,
				"error", err,
			)
			if firstErr == nil {
				firstErr = fmt.Errorf("keyring bootstrap %s: %w", integ.Name, err)
			}
			continue
		}

		// If the returned KeyID differs from the input, a new pair was just
		// generated and the DB row needs updating.
		if kp.KeyID != integ.KeyID {
			l.Infoc(ctx, "keyring bootstrap: generated new key pair",
				"integration", integ.Name,
				"kid", kp.KeyID,
			)
			if err := updateFn(ctx, integ.Name, string(kp.PublicKeyPEM), kp.KeyID); err != nil {
				l.Errorc(ctx, "keyring bootstrap: failed to persist new key to DB",
					"integration", integ.Name,
					"kid", kp.KeyID,
					"error", err,
				)
				if firstErr == nil {
					firstErr = fmt.Errorf("keyring bootstrap persist %s: %w", integ.Name, err)
				}
				continue
			}
		} else {
			l.Infoc(ctx, "keyring bootstrap: key pair already exists",
				"integration", integ.Name,
				"kid", kp.KeyID,
			)
		}
	}
	return firstErr
}
