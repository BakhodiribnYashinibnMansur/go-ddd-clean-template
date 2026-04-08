# Deployment Runbook

> Go Clean Template (GCT) Backend -- operational procedures for build, deploy, verify, and rollback.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Environment Configuration](#2-environment-configuration)
3. [Database Migration](#3-database-migration)
4. [Build and Deploy](#4-build-and-deploy)
5. [Health Checks](#5-health-checks)
6. [Post-Deploy Verification](#6-post-deploy-verification)
7. [Monitoring](#7-monitoring)
8. [Rollback Procedure](#8-rollback-procedure)
9. [Troubleshooting](#9-troubleshooting)
10. [Security Checklist](#10-security-checklist)

---

## 1. Prerequisites

### Required Tools

| Tool | Minimum Version | Install |
|------|----------------|---------|
| Go | 1.26+ | https://go.dev/dl/ |
| Docker | 24+ | https://docs.docker.com/get-docker/ |
| Docker Compose | v2+ (plugin) | bundled with Docker Desktop |
| goose | v3 | `go install github.com/pressly/goose/v3/cmd/goose@latest` |
| swag | latest | `go install github.com/swaggo/swag/cmd/swag@latest` |
| air | latest | `go install github.com/air-verse/air@latest` |
| golangci-lint | latest | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| sqlc | latest | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| k6 | latest | https://k6.io/docs/get-started/installation/ |
| schemathesis | latest | installed via `make test-schemathesis-install` (Python venv) |

Install all Go tools at once:

```bash
make bin-deps
```

### Required Secrets / Credentials

Before any deployment you must have:

- PostgreSQL credentials (`PG_USER`, `PG_PASSWORD`)
- Redis password (`REDIS_PASSWORD`)
- MinIO access/secret keys (`MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`)
- JWT RSA key pair (`JWT_PRIVATE_KEY`, `JWT_PUBLIC_KEY`) -- generate with `make keygen`
- CSRF secret (`CSRF_SECRET`) -- minimum 32 characters, generate with `openssl rand -base64 32`

---

## 2. Environment Configuration

### .env File

The application loads configuration in this order:

1. `.env` file in the project root (if present)
2. `.env.example` as fallback (with a warning)
3. Environment variables override both

Copy the example and fill in production values:

```bash
cp .env.example .env
```

### Key Configuration Sections

```dotenv
# ── Application ──────────────────────────────────────
APP_NAME=go-clean-template
APP_VERSION=1.0.0
HTTP_PORT=8080
HTTP_USE_PREFORK_MODE=false
LOG_LEVEL=debug          # debug | info | warn | error

# ── PostgreSQL ───────────────────────────────────────
PG_POOL_MAX=2
PG_HOST=localhost
PG_PORT=5432
PG_NAME=db
PG_USER=user
PG_PASSWORD=pass

# ── Redis ────────────────────────────────────────────
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_NAME=0
REDIS_USER=
REDIS_PASSWORD=

# ── MinIO (file storage) ────────────────────────────
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_REGION=us-east-1
MINIO_BUCKET=my-bucket

# ── JWT Authentication ──────────────────────────────
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h
JWT_ISSUER=auth-service
JWT_PRIVATE_KEY=""       # PEM-encoded RSA private key
JWT_PUBLIC_KEY=""        # PEM-encoded RSA public key

# ── Observability ───────────────────────────────────
METRICS_ENABLED=true
SWAGGER_ENABLED=true
TRACING_ENABLED=false
TRACING_SERVICE_NAME=go-clean-template
TRACING_ENDPOINT=http://localhost:14268/api/traces

# ── Rate Limiting ───────────────────────────────────
LIMITER_ENABLED=true
LIMITER_LIMIT=100
LIMITER_PERIOD=M         # S = second, M = minute, H = hour

# ── CSRF ────────────────────────────────────────────
CSRF_SECRET=<openssl rand -base64 32>

# ── Asynq (background jobs) ────────────────────────
ASYNQ_ADDR=              # defaults to Redis addr if empty
ASYNQ_PASSWORD=
ASYNQ_DB=0
ASYNQ_CONCURRENCY=10
ASYNQ_MAX_RETRY=3
ASYNQ_WORKER_ENABLED=true

# ── Feature Flags ──────────────────────────────────
FEATURE_FLAG_ENABLED=true
FEATURE_FLAG_CONFIG_PATH=./config/flags.yaml
FEATURE_FLAG_USE_FILE=true
FEATURE_FLAG_USE_REDIS=false

# ── Seeder (dev/staging only) ──────────────────────
SEEDER_ENABLED=false
SEEDER_USERS_COUNT=50
```

### Docker Compose Overrides

When running via `docker-compose.yml`, environment variables are defined in the `x-backend-env` anchor. Key differences from local:

- `PG_URL` uses the Docker service name `db` as host
- `TRACING_ENDPOINT` points to `http://jaeger:14268/api/traces`
- `TRACING_ENABLED=true` by default
- `RMQ_URL` uses the `rabbitmq` service name
- `GOMAXPROCS=1` (adjust per container resource limits)

---

## 3. Database Migration

Migration files live in `migrations/postgres/` (50 files, goose sequential format).

The Makefile exports `GOOSE_MIGRATION_DIR`, `GOOSE_DRIVER`, and `GOOSE_DBSTRING` from your `.env` automatically.

### Commands

| Command | Description |
|---------|-------------|
| `make migration-status` | Show applied/pending status for every migration |
| `make migration-up` | Apply all pending migrations |
| `make migration-down` | Roll back the most recent migration (one step) |
| `make migration-redo` | Roll back then re-apply the latest migration |
| `make migration-validate` | Check migration files for errors without executing |
| `make migration-reset` | Roll back ALL migrations (destructive) |
| `make migration-fix` | Apply sequential ordering to migration files |
| `make migration-create` | Create a new migration file (interactive prompt) |

### Deployment Workflow

```bash
# 1. Validate migrations before applying
make migration-validate

# 2. Check current status
make migration-status

# 3. Apply pending migrations
make migration-up

# 4. Verify
make migration-status
```

### Important Notes

- Always run `migration-validate` before `migration-up` in CI/CD.
- `migration-reset` drops everything -- never use in production.
- The Docker image copies migrations into `/migrations` so goose can run them inside the container.
- Goose uses a `goose_db_version` table to track state.

---

## 4. Build and Deploy

### Local Development

```bash
# Full run: generates swagger docs, builds, and starts the server
make run

# Hot reload (watches for file changes, auto-restarts)
make air

# Initialize air config if first time
make air-init
```

`make run` internally:
1. Kills any process on `HTTP_PORT` (default 8080)
2. Runs `swag init` to regenerate Swagger docs
3. Builds and runs `./cmd/app` with `CGO_ENABLED=0`

### Docker Build

```bash
# Build the backend image
docker build -t gct-backend .
```

The Dockerfile uses a 4-stage multi-stage build:

1. **modules** -- downloads Go dependencies
2. **proto-generator** -- compiles `.proto` files
3. **builder** -- compiles the Go binary (`CGO_ENABLED=0 GOOS=linux GOARCH=amd64`)
4. **final** -- `gcr.io/distroless/static-debian12` with only the binary, config, migrations, and CA certs

### Docker Compose

```bash
# Infrastructure only (PostgreSQL, Redis, MinIO, RabbitMQ, NATS,
# Elasticsearch, ClickHouse, Cassandra, MongoDB, MySQL,
# Jaeger, Prometheus, Grafana, Loki, Alertmanager, etc.)
make compose-up

# Full stack (infra + backend + admin panel + nginx reverse proxy)
make compose-up-all

# Tear down everything (removes orphan containers)
make compose-down

# Remove persistent PostgreSQL volume (destructive)
make docker-rm-volume
```

### Service Topology (docker-compose.yml)

| Service | Image | Ports |
|---------|-------|-------|
| db | postgres:18.1-alpine | 5432 |
| redis | redis:7.4-alpine | 6379 |
| minio | minio/minio:latest | 9000 (API), 9001 (console) |
| rabbitmq | rabbitmq:4.2.1-management | 5672, 15672 |
| nats | nats:2.12-alpine | 4222, 8222 |
| mongodb | mongo:8.0 | 27017 |
| mysql | mysql:9.1 | 3306 |
| elasticsearch | elasticsearch:8.17.0 | 9200 |
| clickhouse | clickhouse-server:24.12-alpine | 8123 |
| cassandra | cassandra:5.0 | 9042 |
| jaeger | jaegertracing/all-in-one:1.64 | 16686, 14268, 4317, 4318 |
| prometheus | prom/prometheus:v3.4.0 | 9090 |
| grafana | grafana/grafana:11.6.0 | 3001 |
| loki | grafana/loki:3.4.2 | 3100 |
| alertmanager | prom/alertmanager:v0.28.1 | 9093 |
| redis-exporter | redis_exporter:v1.66.0 | 9121 |
| promtail | grafana/promtail:3.4.2 | -- |
| gca (backend) | built from ./GCA | 8080, 8081 |
| admin-panel | built from ./AdminPanel | 3000 |
| nginx | nginx:1.29.4-alpine | 80 |

### Resource Limits (backend container)

- CPU: 1.00 (reserved 0.60)
- Memory: 2 GB (reserved 1 GB)

---

## 5. Health Checks

### Endpoints

| Endpoint | Purpose | Expected Response |
|----------|---------|-------------------|
| `GET /health/live` | Liveness probe -- process is running | `200 OK` |
| `GET /health/ready` | Readiness probe -- DB and dependencies reachable | `200 OK` |
| `GET /healthz` | Kubernetes-style alias | `200 OK` |

### Docker Healthcheck

The `gca` service in docker-compose uses:

```yaml
healthcheck:
  test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
  interval: 15s
  timeout: 5s
  retries: 5
```

Dependent services (admin-panel, nginx) wait for `condition: service_healthy`.

### Manual Check

```bash
curl -sf http://localhost:8080/health/live   && echo "LIVE"
curl -sf http://localhost:8080/health/ready  && echo "READY"
```

---

## 6. Post-Deploy Verification

Run these checks after every deployment:

### Smoke Test (k6)

```bash
make test-k6-smoke
```

Runs a fast sanity check against `http://localhost:8080` covering basic endpoint reachability.

### Full k6 Suite

```bash
make test-k6-auth      # Auth flow load test
make test-k6-crud      # CRUD operations
make test-k6-files     # File upload
make test-k6-mixed     # Mixed workload
make test-k6-all       # All of the above
```

### Schemathesis (API contract testing)

```bash
# Quick (10 examples, fail-fast)
make test-schemathesis-quick

# Standard (50 examples, 4 workers)
make test-schemathesis

# Stateful (link-based workflow testing)
make test-schemathesis-stateful

# Full suite
make test-schemathesis-full
```

Schemathesis tests against the Swagger spec at `docs/swagger/swagger.yaml`.

### Swagger UI

If `SWAGGER_ENABLED=true`, open:

```
http://localhost:8080/swagger/index.html
```

### Metrics Endpoint

```bash
curl -s http://localhost:8080/metrics | head -20
```

### Recommended Post-Deploy Sequence

```bash
# 1. Health
curl -sf http://localhost:8080/health/ready

# 2. Smoke test
make test-k6-smoke

# 3. Schema contract
make test-schemathesis-quick

# 4. Verify metrics are flowing
curl -sf http://localhost:9090/api/v1/targets | grep -c '"health":"up"'
```

---

## 7. Monitoring

### Dashboards and UIs

| Service | URL | Credentials |
|---------|-----|-------------|
| Prometheus | http://localhost:9090 | none |
| Grafana | http://localhost:3001 | admin / admin |
| Jaeger (traces) | http://localhost:16686 | none |
| Alertmanager | http://localhost:9093 | none |
| Loki (logs) | via Grafana data source | -- |
| RabbitMQ Management | http://localhost:15672 | guest / guest |
| MinIO Console | http://localhost:9001 | minioUser / minioPassword123 |
| NATS Monitoring | http://localhost:8222 | -- |

### Prometheus

- Config: `monitoring/prometheus/prometheus.yml`
- Alert rules: `monitoring/prometheus/alerts.yml`
- Retention: 15 days
- Lifecycle API enabled (`--web.enable-lifecycle`)
- Reload config without restart: `curl -X POST http://localhost:9090/-/reload`

### Grafana

- Provisioning: `monitoring/grafana/provisioning/`
- Dashboards: `monitoring/grafana/dashboards/`
- Data sources are auto-provisioned (Prometheus, Loki, Jaeger)

### Loki (Logs)

- Config: `monitoring/loki/loki-config.yml`
- Promtail config: `monitoring/promtail/promtail-config.yml`
- Promtail scrapes Docker container logs from `/var/lib/docker/containers`
- Query logs in Grafana Explore with LogQL

### Jaeger (Tracing)

- Requires `TRACING_ENABLED=true`
- Collector endpoint: `http://jaeger:14268/api/traces` (inside Docker network)
- OTLP gRPC: port 4317, OTLP HTTP: port 4318

### Alertmanager

- Config: `monitoring/alertmanager/alertmanager.yml`
- Receives alerts from Prometheus alert rules
- Configure notification channels (email, Slack, PagerDuty) in the alertmanager config

### Redis Exporter

- Exposes Redis metrics at `:9121/metrics`
- Scraped by Prometheus automatically

---

## 8. Rollback Procedure

### Standard Rollback

```bash
# 1. Stop the current deployment
make compose-down
# or: docker stop gca

# 2. Roll back the last migration (if the new release added one)
make migration-down
# Repeat if multiple migrations were added in the release

# 3. Verify migration state
make migration-status

# 4. Deploy the previous version
git checkout <previous-tag-or-commit>
make compose-up-all
# or: docker build -t gct-backend . && docker run ...

# 5. Verify rollback
curl -sf http://localhost:8080/health/ready
make test-k6-smoke
```

### Emergency Rollback (Docker)

```bash
# If you tagged images before deploy:
docker stop gca
docker run -d --name gca gct-backend:<previous-tag>
```

### Migration Rollback Notes

- `make migration-down` rolls back ONE migration at a time.
- Always check `make migration-status` to confirm you are at the correct version.
- Some migrations may not be safely reversible (data migrations, column drops). Review the down-migration SQL before running.
- Never run `make migration-reset` in production.

---

## 9. Troubleshooting

### Port Conflict

**Symptom:** `bind: address already in use` on startup.

```bash
# Kill the process occupying HTTP_PORT
make kill-port

# Or manually:
lsof -ti tcp:8080 | xargs kill -9
```

### Database Connection Failure

**Symptom:** Application fails to start, logs show `connection refused` or `role does not exist`.

1. Verify PostgreSQL is running:
   ```bash
   docker ps | grep db
   pg_isready -h localhost -p 5432 -U user
   ```
2. Check `.env` values match docker-compose environment.
3. Verify the database exists:
   ```bash
   psql -h localhost -p 5432 -U user -l
   ```
4. If using Docker, ensure the service is healthy:
   ```bash
   docker inspect db --format='{{.State.Health.Status}}'
   ```

### Redis Connection Failure

**Symptom:** `NOAUTH Authentication required` or `connection refused`.

1. Verify Redis is running:
   ```bash
   docker ps | grep redis
   redis-cli -a <password> ping
   ```
2. Check `REDIS_PASSWORD` in `.env` matches the Docker compose password (default: `redisPassword123`).
3. For Asynq issues, verify `ASYNQ_ADDR` and `ASYNQ_PASSWORD` are set correctly.

### JWT / Authentication Errors

**Symptom:** `401 Unauthorized`, `token signature is invalid`, `key not found`.

1. Verify RSA keys are generated:
   ```bash
   make keygen
   ```
2. Confirm `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY` are set in `.env` (PEM-encoded, in quotes).
3. Per-integration keys: each integration has its own RSA key pair stored in the database. Verify the `integrations` table has the correct keys seeded (migration `20260406001000_seed_jwt_integrations.sql`).
4. Check token expiry: `JWT_ACCESS_TTL=15m` means tokens expire quickly in dev.

### Migration Failures

**Symptom:** `goose: migration failed` or schema mismatch errors.

1. Check current status:
   ```bash
   make migration-status
   ```
2. Validate files:
   ```bash
   make migration-validate
   ```
3. If a migration partially applied, check the `goose_db_version` table manually:
   ```sql
   SELECT * FROM goose_db_version ORDER BY id DESC LIMIT 5;
   ```
4. Fix and redo:
   ```bash
   make migration-redo
   ```

### Container Startup Order

**Symptom:** Backend starts before DB is ready.

The docker-compose file uses `depends_on` with `condition: service_healthy` for `db` and `redis`. If you still see timing issues:

```bash
# Restart just the backend after infra is up
docker compose restart gca
```

### High Memory / CPU

1. Check container stats:
   ```bash
   docker stats --no-stream
   ```
2. Backend is limited to 1 CPU / 2 GB RAM via docker-compose resource limits.
3. Check `PG_POOL_MAX` -- increase if you see connection pool exhaustion.
4. Review Prometheus metrics at `http://localhost:9090` for Go runtime stats (`go_memstats_*`, `go_goroutines`).

---

## 10. Security Checklist

Run through this checklist before every production deployment:

### Secrets

- [ ] `.env` is listed in `.gitignore` and never committed
- [ ] `CSRF_SECRET` is a unique, random value (minimum 32 characters)
- [ ] `PG_PASSWORD` is strong and unique per environment
- [ ] `REDIS_PASSWORD` is set (not empty)
- [ ] `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` are changed from defaults
- [ ] `JWT_PRIVATE_KEY` / `JWT_PUBLIC_KEY` are unique per environment (generate with `make keygen`)

### Authentication

- [ ] JWT RSA keys are per-environment (dev, staging, prod use different key pairs)
- [ ] Per-integration JWT keys are seeded correctly
- [ ] `JWT_ACCESS_TTL` is short (15m or less in production)
- [ ] Token revocation is enabled (security audit log at migration `20260406100000`)

### Container Security

- [ ] Final Docker image uses `gcr.io/distroless/static-debian12` (no shell, no package manager)
- [ ] Build uses `CGO_ENABLED=0` for static linking
- [ ] Only the binary, config, migrations, and CA certs are in the final image
- [ ] Container resource limits are set (CPU and memory)

### Network Security

- [ ] CORS is configured for allowed origins only (not `*` in production)
- [ ] Rate limiting is enabled (`LIMITER_ENABLED=true`)
- [ ] CSRF protection is enabled with a strong secret
- [ ] Swagger UI is disabled in production (`SWAGGER_ENABLED=false`)
- [ ] Tracing endpoints are not publicly accessible

### Monitoring

- [ ] Prometheus, Grafana, and Alertmanager are not exposed to the public internet
- [ ] Grafana default password is changed from `admin/admin`
- [ ] Alert rules are configured for critical conditions (see `monitoring/prometheus/alerts.yml`)

### Database

- [ ] `SEEDER_ENABLED=false` in production
- [ ] Database user has minimal required privileges
- [ ] SSL mode is appropriate for the environment (`PG_SSLMODE`)
- [ ] All migrations are validated before applying (`make migration-validate`)

---

## Quick Reference

```bash
# ── Development ──────────────────────────────
make run                     # Build + run locally
make air                     # Hot reload
make swag                    # Regenerate Swagger docs
make sqlc                    # Regenerate SQLC code
make keygen                  # Generate JWT RSA keys
make format                  # Format code (gofumpt + gci)
make pre-commit              # Full pre-commit suite

# ── Testing ──────────────────────────────────
make test                    # Unit tests with coverage
make test-e2e                # End-to-end tests
make integration-test        # Integration tests
make test-k6-smoke           # k6 smoke test
make test-schemathesis-quick # Quick API contract test
make test-api-all            # All tests combined
make arch-check              # DDD architecture rules

# ── Docker ───────────────────────────────────
make compose-up              # Infra only
make compose-up-all          # Full stack
make compose-down            # Tear down

# ── Database ─────────────────────────────────
make migration-status        # Current state
make migration-validate      # Validate files
make migration-up            # Apply pending
make migration-down          # Roll back one step

# ── Linting ──────────────────────────────────
make linter-golangci         # Go linter
make linter-hadolint         # Dockerfile linter
make linter-dotenv           # .env linter
make deps-audit              # Vulnerability check
```
