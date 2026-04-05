# JWT Signing Keys

This directory holds the RSA key pair used by the JWT package to sign and
verify access tokens. **The contents are gitignored** (`*.pem` in `.gitignore`).

## Files

| File              | Mode | Purpose                                 |
|-------------------|------|-----------------------------------------|
| `jwt_private.pem` | 0600 | RS256 signing key — **never commit**    |
| `jwt_public.pem`  | 0644 | Verification key — safe to distribute   |

## Generating a key pair

```bash
# RSA-4096 (default) → config/keys/
go run ./cmd/keygen

# Rotate existing keys
go run ./cmd/keygen -force

# Only generate a new refresh pepper (prints to stdout)
go run ./cmd/keygen -pepper-only
```

## Configuration

Point the app at the generated files (YAML or env vars):

```yaml
jwt:
  private_key_path: "config/keys/jwt_private.pem"
  public_key_path:  "config/keys/jwt_public.pem"
```

Or:

```bash
export JWT_PRIVATE_KEY_PATH=/run/secrets/jwt_private.pem
export JWT_PUBLIC_KEY_PATH=/run/secrets/jwt_public.pem
export JWT_REFRESH_PEPPER="<output from keygen>"
```

## Rotation

Rotating `jwt_private.pem` invalidates all outstanding access tokens (their
signatures stop verifying). Refresh tokens continue to work because their
hashing depends on `JWT_REFRESH_PEPPER`, not on the RSA key.

Rotating `JWT_REFRESH_PEPPER` invalidates all outstanding **refresh** tokens.
Use this after a suspected DB dump.

For zero-downtime rotation, use the `key_id` (`kid`) JWT header and publish
multiple public keys server-side. (Not yet wired up — future work.)
