# ADR-0010: Distroless Container Image

**Status:** ACCEPTED  
**Date:** 2026-04-07  

## Context

The GCT backend is a statically-linked Go binary with no runtime dependencies on
libc or system utilities. The Docker image used in production should minimise attack
surface, image size, and CVE exposure from unused OS packages.

## Decision

Use `gcr.io/distroless/static-debian12` as the final stage in the multi-stage
Dockerfile:

```dockerfile
# Step 4: Final
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/config /config
COPY --from=builder /app/migrations /migrations
COPY --from=builder /bin/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app"]
```

The build pipeline has four stages:
1. `modules` -- downloads Go module cache (`golang:1.26.1-alpine3.21`)
2. `proto-generator` -- compiles protobuf definitions
3. `builder` -- compiles the binary with `CGO_ENABLED=0`
4. `final` -- distroless image containing only the binary, config, migrations, and
   CA certificates

## Consequences

### Positive
- No shell, no package manager, no unused libraries -- dramatically reduced attack
  surface.
- Image size is ~15-25 MB compared to ~150 MB for Alpine-based images.
- Fewer CVEs to triage in vulnerability scans.

### Negative
- No shell access for debugging; must use ephemeral debug containers
  (`kubectl debug`) or sidecar tools.
- Cannot install runtime utilities (curl, wget) for ad-hoc troubleshooting.
- CA certificates must be explicitly copied from the builder stage.

## Alternatives Considered

- **Alpine** (`golang:alpine`) -- includes musl libc, apk, and a shell; useful for
  debugging but adds unnecessary packages and a larger attack surface.
- **scratch** -- even smaller than distroless but lacks CA certificates, timezone
  data, and `/etc/passwd`; requires more manual setup.
- **Ubuntu/Debian slim** -- familiar but carries hundreds of unnecessary packages;
  larger image and more CVE noise.
