# ADR-0006: RSA-Based JWT Authentication with Per-Integration Key Pairs

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

The GCT platform serves three client integrations -- `gct-admin`, `gct-client`, and
`gct-mobile` -- each with different trust profiles and token lifetimes. A shared
symmetric secret means compromising one integration compromises all. We also need a
JWKS endpoint so external services can verify tokens without sharing secrets.

## Decision

Use RSA-256 asymmetric JWT with a dedicated key pair per integration, managed by a
keyring service (`internal/kernel/infrastructure/security/keyring/`):

- **Key generation**: `bootstrap.go` generates or loads RSA key pairs at startup,
  one per integration ID.
- **Key rotation**: `rotate_task.go` rotates keys on a configurable schedule; old
  keys remain valid for verification until all outstanding tokens expire.
- **Token lifetimes**: access tokens 15 minutes, refresh tokens 720 hours (30 days).
- **JWKS endpoint**: `internal/kernel/infrastructure/security/jwks/` exposes public
  keys per integration so clients and API gateways can verify tokens independently.

Access tokens (`jwt/access_token.go`) carry integration ID, user ID, and role claims.
Refresh tokens are opaque and stored server-side in the session BC.

## Consequences

### Positive
- Compromising one integration's private key does not affect others.
- Public-key verification enables zero-trust token validation at edge proxies.
- Key rotation limits the blast radius of a leaked key.

### Negative
- RSA operations are slower than HMAC; mitigated by the 15-minute access token cache.
- Per-integration key pairs increase operational surface (more keys to manage).
- Refresh token stored server-side requires a database round-trip on refresh.

## Alternatives Considered

- **Symmetric HMAC JWT** -- simpler but forces secret sharing; no JWKS support.
- **Delegated OAuth2 server** (Keycloak, Hydra) -- full-featured but adds a
  separate service to deploy and operate; premature for current team size.
- **Single RSA key pair for all integrations** -- simpler key management but a single
  point of compromise; cannot set per-integration policies.
