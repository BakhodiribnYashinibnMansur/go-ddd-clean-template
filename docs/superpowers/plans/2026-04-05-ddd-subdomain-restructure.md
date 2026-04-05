# DDD Subdomain Restructure (Core / Supporting / Generic)

> **Status:** DRAFT — user reviews first, then we adjust before execution.
> **For agentic workers:** Use superpowers:executing-plans once approved. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Reorganize `internal/context/` from current flat-by-area layout (`admin/`, `iam/`, `ops/`, `content/`) into DDD strategic classification (`core/`, `supporting/`, `generic/`) so that strategic importance of each BC is explicit in the folder structure.

**Architecture:** Pure mechanical restructure — no behavior changes. Move directories, update Go import paths (`gct/internal/context/<old>` → `gct/internal/context/<new>`), verify `go build ./...` and `go test ./...` after each move. One BC per commit for easy rollback.

**Tech Stack:** Go 1.25, module `gct`. ~601 Go files under `internal/context/`, ~1260 files importing from it across the codebase.

---

## Decision Point (BEFORE Task 1)

User must confirm the target layout. Current draft:

```
internal/context/
├── core/                           # EMPTY — reserved for future core domain
│   └── .gitkeep
│
├── supporting/                     # Custom biz helpers (no off-the-shelf alt)
│   ├── audit/                      ← from iam/audit
│   ├── iprule/                     ← from ops/iprule
│   ├── announcement/               ← from content/announcement
│   ├── statistics/                 ← from admin/statistics
│   ├── integration/                ← from admin/integration
│   ├── sitesetting/                ← from admin/sitesetting
│   ├── dataexport/                 ← from admin/dataexport
│   └── errorcode/                  ← from admin/errorcode
│
└── generic/                        # Replaceable w/ SaaS / off-the-shelf
    ├── identity/
    │   ├── user/                   ← from iam/user
    │   ├── session/                ← from iam/session
    │   └── usersetting/            ← from iam/usersetting
    ├── authorization/              ← from iam/authz
    ├── messaging/                  ← from content/notification
    ├── storage/                    ← from content/file
    ├── i18n/                       ← from content/translation
    ├── observability/
    │   ├── metric/                 ← from ops/metric
    │   └── systemerror/            ← from ops/systemerror
    ├── throttling/                 ← from ops/ratelimit
    └── toggles/                    ← from admin/featureflag
```

**Alternative schemes to consider:**
- **Option A (above):** full flattening under core/supporting/generic + semantic grouping
- **Option B:** keep `iam/ops/content/admin` under each strategic bucket (hybrid)
- **Option C:** only add strategic prefix, keep current sub-areas (e.g. `generic-iam/`, `supporting-admin/`)

**Questions for user:**
1. Which option (A/B/C)?
2. Is `identity/{user,session,usersetting}` nesting OK, or flatten to `generic/{user,session,usersetting,...}`?
3. `authz` → rename to `authorization` or keep short name?
4. Any BC the user wants to reclassify (e.g. move `announcement` from supporting to generic)?

**STOP here until user answers.** Rest of plan assumes Option A.

---

## Migration Strategy

- **One BC per task, one commit per BC.** If anything breaks, `git revert` isolates damage.
- **Order:** leaves first (no internal dependencies), trunk last. Start with BCs that other BCs don't import.
- **Per-BC procedure (the same 4 steps every time):**
  1. `git mv` the directory.
  2. Global replace import path: `gct/internal/context/<old>` → `gct/internal/context/<new>`.
  3. `go build ./...` — must pass.
  4. `go test ./...` — must pass.
  5. Commit.

**Tooling:** Use `grep -rl` + `sed -i ''` (macOS) for import rewrites. Go's `goimports` doesn't rename paths, so sed is the right tool.

**Safety rail:** Before starting, run full test suite + `go build ./...` to get a green baseline. If baseline is red, stop and fix first.

---

## Task 0: Baseline Verification

**Files:** none

- [ ] **Step 1:** Check git is clean (or stash uncommitted work)

Run: `git status`
Expected: clean working tree, or known-good WIP stashed.

- [ ] **Step 2:** Full build

Run: `go build ./...`
Expected: exit 0, no output.

- [ ] **Step 3:** Full test suite

Run: `go test ./...`
Expected: all packages PASS.

- [ ] **Step 4:** Record baseline commit

Run: `git rev-parse HEAD` → save this SHA as rollback point.

---

## Task 1: Create Target Directory Skeleton

**Files:**
- Create: `internal/context/core/.gitkeep`
- Create: `internal/context/supporting/.gitkeep`
- Create: `internal/context/generic/.gitkeep`
- Create: `internal/context/generic/identity/.gitkeep`
- Create: `internal/context/generic/observability/.gitkeep`

- [ ] **Step 1:** Create directories

```bash
mkdir -p internal/context/core \
         internal/context/supporting \
         internal/context/generic/identity \
         internal/context/generic/observability
touch internal/context/core/.gitkeep \
      internal/context/supporting/.gitkeep \
      internal/context/generic/.gitkeep \
      internal/context/generic/identity/.gitkeep \
      internal/context/generic/observability/.gitkeep
```

- [ ] **Step 2:** Commit

```bash
git add internal/context/core internal/context/supporting internal/context/generic
git commit -m "refactor(ddd): add core/supporting/generic skeleton directories"
```

---

## Tasks 2–12: Move Generic BCs (11 BCs)

> Template per BC. Apply to each BC in the list below.

### Template

**For BC `<old_path>` → `<new_path>`:**

- [ ] **Step 1:** Move directory

```bash
git mv internal/context/<old_path> internal/context/<new_path>
```

- [ ] **Step 2:** Update import paths across the codebase

```bash
grep -rl "gct/internal/context/<old_path>" --include="*.go" . | \
  xargs sed -i '' 's|gct/internal/context/<old_path>|gct/internal/context/<new_path>|g'
```

- [ ] **Step 3:** Verify build

Run: `go build ./...`
Expected: exit 0.

- [ ] **Step 4:** Verify tests

Run: `go test ./...`
Expected: all PASS.

- [ ] **Step 5:** Commit

```bash
git add -A
git commit -m "refactor(ddd): move <bc-name> to generic/<path>"
```

### BC Move List (Generic)

| # | Old path | New path |
|---|----------|----------|
| 2 | `iam/user` | `generic/identity/user` |
| 3 | `iam/session` | `generic/identity/session` |
| 4 | `iam/usersetting` | `generic/identity/usersetting` |
| 5 | `iam/authz` | `generic/authorization` |
| 6 | `content/notification` | `generic/messaging` |
| 7 | `content/file` | `generic/storage` |
| 8 | `content/translation` | `generic/i18n` |
| 9 | `ops/metric` | `generic/observability/metric` |
| 10 | `ops/systemerror` | `generic/observability/systemerror` |
| 11 | `ops/ratelimit` | `generic/throttling` |
| 12 | `admin/featureflag` | `generic/toggles` |

> ⚠️ Tasks 5, 6, 7, 9, 11, 12 rename the BC directory itself (e.g. `authz` → `authorization`). This means the Go **package name** inside may also need renaming. Check: after `git mv`, open a file in the moved directory; if `package authz` appears, decide whether to rename package to `authorization`. **Recommendation:** keep package names short (`authz`, `notification`, etc.) even if the directory name is longer — Go imports use the last path segment by default, but explicit `package` declaration controls the import identifier. Test this on Task 5 first; if it causes churn, keep original package names.

---

## Tasks 13–20: Move Supporting BCs (8 BCs)

Same template as above.

| # | Old path | New path |
|---|----------|----------|
| 13 | `iam/audit` | `supporting/audit` |
| 14 | `ops/iprule` | `supporting/iprule` |
| 15 | `content/announcement` | `supporting/announcement` |
| 16 | `admin/statistics` | `supporting/statistics` |
| 17 | `admin/integration` | `supporting/integration` |
| 18 | `admin/sitesetting` | `supporting/sitesetting` |
| 19 | `admin/dataexport` | `supporting/dataexport` |
| 20 | `admin/errorcode` | `supporting/errorcode` |

---

## Task 21: Remove Empty Old Parent Directories

After tasks 2–20, the old `iam/`, `ops/`, `content/`, `admin/` folders should be empty.

- [ ] **Step 1:** Verify they're empty

```bash
find internal/context/iam internal/context/ops internal/context/content internal/context/admin -type f 2>/dev/null
```
Expected: no output.

- [ ] **Step 2:** Remove them

```bash
rm -rf internal/context/iam internal/context/ops internal/context/content internal/context/admin
```

- [ ] **Step 3:** Verify build & test

Run: `go build ./... && go test ./...`
Expected: all PASS.

- [ ] **Step 4:** Commit

```bash
git add -A
git commit -m "refactor(ddd): remove empty iam/ops/content/admin parent dirs"
```

---

## Task 22: Update Route Registration Comments & Helpers

The route-registration helpers in `internal/app/ddd_routes.go` group routes under old area names (`registerIAMRoutes`, `registerOpsRoutes`, etc.). These names now mismatch the new structure.

**Files:**
- Modify: `internal/app/ddd_routes.go`

**Decision for user:** Keep old function names (routes grouping is orthogonal to strategic classification — `registerIAMRoutes` still makes sense as "IAM-related HTTP routes" even if the BCs live under `generic/identity/`). **Recommendation: keep route grouping as-is.** The strategic classification is about *code organization*, not *URL organization*.

- [ ] **Step 1:** Decide with user whether to rename route helpers.

If YES: rename to `registerIdentityRoutes`, `registerObservabilityRoutes`, etc.
If NO: skip this task.

---

## Task 23: Update Architecture Docs

**Files:**
- Modify: `internal/context/doc.go` (if exists — verify)
- Create/Modify: `docs/architecture/ddd-subdomains.md` (if user wants docs)

- [ ] **Step 1:** Check what's in `internal/context/doc.go`

Run: `cat internal/context/doc.go`

- [ ] **Step 2:** Update doc comment to reflect new core/supporting/generic layout.

- [ ] **Step 3:** Commit

```bash
git add internal/context/doc.go
git commit -m "docs(ddd): update context doc.go to reflect strategic layout"
```

---

## Task 24: Final Verification

- [ ] **Step 1:** Full build + test

```bash
go build ./...
go test ./...
go vet ./...
```
Expected: all green.

- [ ] **Step 2:** Check no stale import paths remain

```bash
grep -r "gct/internal/context/iam/" --include="*.go" . || echo "OK: no old iam/ imports"
grep -r "gct/internal/context/ops/" --include="*.go" . || echo "OK: no old ops/ imports"
grep -r "gct/internal/context/content/" --include="*.go" . || echo "OK: no old content/ imports"
grep -r "gct/internal/context/admin/" --include="*.go" . || echo "OK: no old admin/ imports"
```
Expected: all four print "OK: ...".

- [ ] **Step 3:** Run the server locally (smoke test)

Run the app locally, hit a few endpoints (`/api/v1/users`, `/api/v1/sessions`), confirm they still work.

- [ ] **Step 4:** Final commit (if any lint/doc fixes surfaced)

---

## Rollback Plan

If any task breaks beyond repair:

```bash
git reset --hard <baseline-sha-from-Task-0>
```

Each task is a single commit, so `git revert <sha>` on the failing one works too.

---

## Risk Notes

1. **Package name vs directory name mismatch** — Go allows them to differ, but it's confusing. Decide policy in Task 5.
2. **CI/CD config files** may reference old paths (coverage reports, lint excludes). Grep for them:
   ```bash
   grep -r "internal/context/iam\|internal/context/ops\|internal/context/admin\|internal/context/content" --include="*.yml" --include="*.yaml" --include="*.toml" --include="Makefile" --include="*.mk" .
   ```
3. **Generated code / mocks** — if any mock generation directives (`//go:generate`) use old paths, they need updating too. Grep for `go:generate` lines in moved BCs.
4. **IDE / editor history** — goimports cache, GOPATH cache: `go clean -cache -modcache` may be needed if builds behave strangely.

---

## Estimated Commit Count

1 (skeleton) + 11 (generic) + 8 (supporting) + 1 (cleanup) + 0-2 (docs/routes) = **~21-23 commits**.

Each commit is small and independently revertable.
