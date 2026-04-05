# Context Map

> **Purpose:** Document the strategic classification of every Bounded Context (BC) in this codebase per Eric Evans' DDD (Blue Book, Part IV).
>
> **Important:** Strategic classification is a _strategic decision tool_ (where to invest, build vs. buy) — not a code organization mandate. The folder structure under `internal/context/` is organized by **domain area** (iam/ops/content/admin), NOT by subdomain tier. This document is the single source of truth for classification.
>
> **How to use this document:**
> - Before starting work on a BC, check its tier here.
> - **Core:** apply rigorous DDD tactical patterns (aggregates, domain events, invariants). Best engineers. In-house.
> - **Supporting:** pragmatic DDD. Mid-tier effort. Custom because no off-the-shelf fit.
> - **Generic:** minimal DDD ceremony. Could be replaced by SaaS/library. Keep API stable, internals simple.

---

## Subdomain Tiers

### 🔴 Core Domain

_Strategic competitive advantage. Where the business differentiates. In-house, top talent, heavy investment._

**Current state: EMPTY.**

This repository is a **backend template (boilerplate)**. Core domain belongs to the _product_ that consumes this template, not the template itself. When forking this template for a specific product, the product's core BCs should be added under `internal/context/` (e.g. `internal/context/commerce/`, `internal/context/learning/`, `internal/context/finance/` — depending on the product).

| BC | Location | Justification |
|----|----------|---------------|
| _(none yet)_ | — | — |

---

### 🟡 Supporting Subdomains

_Custom business logic that serves the Core. No direct off-the-shelf equivalent. Mid-tier investment._

| BC | Location | Why Supporting (not Generic) |
|----|----------|------------------------------|
| **audit** | `iam/audit` | Compliance (GDPR, SOC2) — audit trail semantics are business-specific (what to log, retention, who can read) |
| **iprule** | `ops/iprule` | Security policy — allow/deny rules encode business risk appetite |
| **announcement** | `content/announcement` | Product-specific broadcast messaging (scheduling, targeting, localization) — not a generic notification channel |
| **statistics** | `admin/statistics` | Business KPIs — each product defines its own aggregations (user stats, content stats, error stats, etc.) |
| **integration** | `admin/integration` | Config registry for outbound external integrations — knowledge of which third parties matters per product |
| **sitesetting** | `admin/sitesetting` | Platform-wide configuration values — product-defined keys and semantics |
| **dataexport** | `admin/dataexport` | GDPR "right to data portability" — jurisdiction/compliance-specific rules |
| **errorcode** | `admin/errorcode` | API error catalog — part of the product's public contract with clients |

---

### 🔵 Generic Subdomains

_Problems every SaaS/web app solves the same way. Off-the-shelf alternatives exist. Low investment — keep simple and stable._

| BC | Location | Off-the-Shelf Alternative | Notes |
|----|----------|---------------------------|-------|
| **user** | `iam/user` | Keycloak, Auth0, Firebase Auth, Ory Kratos | Custom implementation for template simplicity |
| **session** | `iam/session` | Keycloak sessions, Redis session store | — |
| **usersetting** | `iam/usersetting` | — (trivial CRUD) | Per-user preferences storage |
| **authz** | `iam/authz` | Casbin, OpenFGA, Keycloak RBAC, Oso | Roles, permissions, policies, scopes |
| **notification** | `content/notification` | SendGrid, Twilio, Novu, Postmark | Multi-channel notification dispatch |
| **file** | `content/file` | S3 SDK directly, UploadThing, Uploadcare | Upload/download/metadata |
| **translation** | `content/translation` | go-i18n, Crowdin, Lokalise, Phrase | i18n key-value store |
| **metric** | `ops/metric` | Prometheus, StatsD, Datadog | Application metric collection |
| **systemerror** | `ops/systemerror` | Sentry, Rollbar, Bugsnag | Error capture and resolution workflow |
| **ratelimit** | `ops/ratelimit` | Nginx limit_req, Envoy, Redis limiter | Per-endpoint request throttling |
| **featureflag** | `admin/featureflag` | LaunchDarkly, Unleash, Flagsmith, PostHog | Feature toggles and rule groups |

---

## Classification Change Procedure

If a BC's strategic tier changes (e.g. `notification` evolves into a core capability because the product becomes a notification engine):

1. Update this document's table.
2. Update the BC's `doc.go` comment (`// Subdomain: ...` marker).
3. Update the team's investment plan (where to allocate effort).
4. **Do NOT move the directory** — folder structure reflects domain area, not strategy.

---

## BC Communication Rules

Regardless of tier, all BCs follow the same isolation rules (see `internal/context/doc.go`):

- No BC imports another BC directly.
- Cross-BC communication flows through:
  - `gct/internal/contract/events` — Published Language (domain events)
  - `gct/internal/contract/ports` — Anti-Corruption Layer interfaces

---

## Template Consumer Guidance

When you fork this template for a real product:

1. **Identify your Core domain.** Ask: "What does my business do that competitors cannot copy?"
2. **Add your Core BCs** under `internal/context/<your-area>/`.
3. **Record them** in the Core Domain table above.
4. **Invest accordingly:**
   - Core: best engineers, rigorous DDD, highest test coverage.
   - Supporting: pragmatic approach, "good enough" quality.
   - Generic: consider replacing with SaaS if team scales — don't over-engineer.
