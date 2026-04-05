// Package arch enforces DDD bounded context architecture rules at test time.
// These tests are the single source of truth for cross-BC isolation, DDD
// layering, and platform/contracts integrity. Run with `go test ./test/arch/...`.
//
// Architecture:
//
//	internal/
//	├── app/          — composition root (may import anything)
//	├── contexts/
//	│   ├── iam/{user,authz,session,audit,usersetting}/
//	│   ├── content/{file,notification,announcement,translation}/
//	│   ├── admin/{errorcode,featureflag,sitesetting,integration,dataexport}/
//	│   └── ops/{ratelimit,iprule,metric,systemerror,dashboard}/
//	├── contracts/    — Published Language (events/) + ACL ports/
//	└── platform/     — infrastructure & shared kernel (no BC deps allowed)
package arch_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	modulePath  = "gct"
	internalDir = "internal"
	// Top-level non-BC areas under internal/.
	appDir       = "app"
	platformDir  = "platform"
	contractsDir = "contract"
	contextsDir  = "context"
)

// subdomains groups BCs by Evans' subdomain classification.
var subdomains = []string{"iam", "content", "admin", "ops"}

// layer represents a DDD layer within a bounded context.
type layer string

const (
	layerDomain         layer = "domain"
	layerApplication    layer = "application"
	layerInfrastructure layer = "infrastructure"
	layerInterfaces     layer = "interfaces"
)

// layerRules defines which layers each layer may depend on WITHIN the same BC.
// Domain is pure; application depends only on domain; infrastructure and
// interfaces may depend on domain + application.
var layerRules = map[layer]map[layer]bool{
	layerDomain:         {},
	layerApplication:    {layerDomain: true},
	layerInfrastructure: {layerDomain: true, layerApplication: true},
	layerInterfaces:     {layerDomain: true, layerApplication: true},
}

// forbiddenInDomain lists external packages that must never appear in a
// domain layer. Domain must remain pure business logic.
var forbiddenInDomain = []string{
	"github.com/labstack/echo",
	"github.com/gin-gonic/gin",
	"gorm.io/gorm",
	"github.com/jackc/pgx",
	"github.com/redis/go-redis",
	"github.com/minio/minio-go",
	"github.com/golang-jwt/jwt",
	"net/http",
	"database/sql",
}

// repoRoot walks up from cwd to find go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	for dir := wd; dir != "/"; dir = filepath.Dir(dir) {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
	}
	t.Fatal("repo root not found")
	return ""
}

// bcPath represents a bounded context as "<subdomain>/<name>", e.g. "iam/user".
type bcPath struct {
	subdomain string
	name      string
}

func (b bcPath) importPrefix() string {
	return modulePath + "/" + internalDir + "/" + contextsDir + "/" + b.subdomain + "/" + b.name
}

func (b bcPath) dir(root string) string {
	return filepath.Join(root, internalDir, contextsDir, b.subdomain, b.name)
}

// boundedContexts enumerates every BC under contexts/<subdomain>/<bc>/.
func boundedContexts(t *testing.T, root string) []bcPath {
	t.Helper()
	var bcs []bcPath
	for _, sd := range subdomains {
		sdDir := filepath.Join(root, internalDir, contextsDir, sd)
		entries, err := os.ReadDir(sdDir)
		if err != nil {
			continue // subdomain may be empty
		}
		for _, e := range entries {
			if e.IsDir() {
				bcs = append(bcs, bcPath{subdomain: sd, name: e.Name()})
			}
		}
	}
	return bcs
}

// parseImports returns import paths of a Go file.
func parseImports(t *testing.T, path string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	imports := make([]string, 0, len(f.Imports))
	for _, imp := range f.Imports {
		imports = append(imports, strings.Trim(imp.Path.Value, `"`))
	}
	return imports
}

// walkGoFiles invokes fn for every .go file under dir (tests included).
func walkGoFiles(t *testing.T, dir string, fn func(path string)) {
	t.Helper()
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		fn(path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
}

// detectLayer extracts BC + layer from a path like
// internal/contexts/iam/user/domain/... → ("iam/user", "domain").
func detectLayer(rel string) (bc string, l layer) {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	// internal / contexts / <subdomain> / <bc> / <layer> / ...
	if len(parts) < 6 || parts[0] != internalDir || parts[1] != contextsDir {
		return "", ""
	}
	return parts[2] + "/" + parts[3], layer(parts[4])
}

// TestBoundedContextIsolation asserts no BC imports another BC. All cross-BC
// communication must flow through gct/internal/contracts.
func TestBoundedContextIsolation(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	bcs := boundedContexts(t, root)

	// Build a set of all BC import prefixes for quick lookup.
	bcPrefixes := make(map[string]bcPath, len(bcs))
	for _, b := range bcs {
		bcPrefixes[b.importPrefix()] = b
	}

	for _, bc := range bcs {
		self := bc.importPrefix()
		walkGoFiles(t, bc.dir(root), func(path string) {
			rel, _ := filepath.Rel(root, path)
			for _, imp := range parseImports(t, path) {
				// Does this import belong to a different BC?
				for prefix, other := range bcPrefixes {
					if prefix == self {
						continue
					}
					if imp == prefix || strings.HasPrefix(imp, prefix+"/") {
						t.Errorf("cross-BC import forbidden:\n  file: %s\n  BC:   %s\n  imp:  %s\n  fix:  route through gct/internal/contracts (events or ports)",
							rel, bc.importPrefix(), imp)
						_ = other
					}
				}
			}
		})
	}
}

// TestLayerDependencies verifies intra-BC DDD layer discipline.
func TestLayerDependencies(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	bcs := boundedContexts(t, root)

	for _, bc := range bcs {
		walkGoFiles(t, bc.dir(root), func(path string) {
			rel, _ := filepath.Rel(root, path)
			_, currentLayer := detectLayer(rel)
			allowed, known := layerRules[currentLayer]
			if !known {
				return
			}
			selfPrefix := bc.importPrefix() + "/"
			for _, imp := range parseImports(t, path) {
				if !strings.HasPrefix(imp, selfPrefix) {
					continue
				}
				tail := strings.TrimPrefix(imp, selfPrefix)
				impLayer := layer(strings.SplitN(tail, "/", 2)[0])
				if impLayer == currentLayer {
					continue
				}
				if !allowed[impLayer] {
					t.Errorf("DDD layer violation:\n  file: %s\n  from: %s layer\n  to:   %s layer\n  imp:  %s",
						rel, currentLayer, impLayer, imp)
				}
			}
		})
	}
}

// TestDomainPurity verifies domain layer has no forbidden external deps.
func TestDomainPurity(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	bcs := boundedContexts(t, root)

	for _, bc := range bcs {
		domainDir := filepath.Join(bc.dir(root), string(layerDomain))
		if _, err := os.Stat(domainDir); err != nil {
			continue
		}
		walkGoFiles(t, domainDir, func(path string) {
			if strings.HasSuffix(path, "_test.go") {
				return
			}
			rel, _ := filepath.Rel(root, path)
			for _, imp := range parseImports(t, path) {
				for _, forbidden := range forbiddenInDomain {
					if strings.HasPrefix(imp, forbidden) {
						t.Errorf("domain purity violation:\n  file: %s\n  imp:  %s\n  rule: domain must not depend on %q",
							rel, imp, forbidden)
					}
				}
			}
		})
	}
}

// TestPlatformHasNoBCDeps verifies the platform layer (infrastructure / shared
// kernel) does not reach into any bounded context.
func TestPlatformHasNoBCDeps(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	platformPath := filepath.Join(root, internalDir, platformDir)

	bcImportPrefix := modulePath + "/" + internalDir + "/" + contextsDir + "/"
	walkGoFiles(t, platformPath, func(path string) {
		rel, _ := filepath.Rel(root, path)
		for _, imp := range parseImports(t, path) {
			if strings.HasPrefix(imp, bcImportPrefix) {
				t.Errorf("platform must not depend on a bounded context:\n  file: %s\n  imp:  %s",
					rel, imp)
			}
		}
	})
}

// TestContractsHaveNoBCDeps verifies contracts/ depends only on platform/.
// Contracts are the stable Published Language + ACL surface; they must not
// couple back to any BC implementation.
func TestContractsHaveNoBCDeps(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	contractsPath := filepath.Join(root, internalDir, contractsDir)

	bcImportPrefix := modulePath + "/" + internalDir + "/" + contextsDir + "/"
	walkGoFiles(t, contractsPath, func(path string) {
		rel, _ := filepath.Rel(root, path)
		for _, imp := range parseImports(t, path) {
			if strings.HasPrefix(imp, bcImportPrefix) {
				t.Errorf("contracts must not depend on a bounded context:\n  file: %s\n  imp:  %s",
					rel, imp)
			}
		}
	})
}

// TestDomainHasTypedIDs enforces that every BC which has a domain/ layer also
// declares at least one typed ID in domain/id.go (e.g. `type UserID uuid.UUID`).
// After the typed ID migration, every aggregate-owning BC publishes typed IDs
// to prevent identifier mix-ups at call sites; this test stops new BCs from
// being added without them. BCs without a domain/ directory (e.g. query-only
// BCs) are skipped.
func TestDomainHasTypedIDs(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	bcs := boundedContexts(t, root)

	var missing []string
	for _, bc := range bcs {
		domainDir := filepath.Join(bc.dir(root), string(layerDomain))
		if _, err := os.Stat(domainDir); err != nil {
			continue // skip BCs without a domain/ directory
		}
		idPath := filepath.Join(domainDir, "id.go")
		if _, err := os.Stat(idPath); err != nil {
			missing = append(missing, bc.subdomain+"/"+bc.name+" (missing domain/id.go)")
			continue
		}
		if !hasTypedIDDecl(t, idPath) {
			missing = append(missing, bc.subdomain+"/"+bc.name+" (domain/id.go has no typed ID declaration)")
		}
	}
	if len(missing) > 0 {
		t.Errorf("BCs missing typed IDs in domain/id.go:\n  %s\n  fix: add `type XxxID uuid.UUID` in domain/id.go",
			strings.Join(missing, "\n  "))
	}
}

// hasTypedIDDecl reports whether path declares at least one type whose name
// ends in "ID" and whose underlying type references uuid.UUID.
func hasTypedIDDecl(t *testing.T, path string) bool {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if !strings.HasSuffix(ts.Name.Name, "ID") {
				continue
			}
			sel, ok := ts.Type.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok {
				continue
			}
			if pkg.Name == "uuid" && sel.Sel.Name == "UUID" {
				return true
			}
		}
	}
	return false
}

// TestBCsDependOnlyOnPlatformAndContracts asserts that any non-self
// gct/internal import a BC makes resolves to platform/ or contracts/.
func TestBCsDependOnlyOnPlatformAndContracts(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	bcs := boundedContexts(t, root)

	allowedInternalPrefixes := []string{
		modulePath + "/" + internalDir + "/" + platformDir + "/",
		modulePath + "/" + internalDir + "/" + platformDir + `"`,
		modulePath + "/" + internalDir + "/" + contractsDir + "/",
		modulePath + "/" + internalDir + "/" + contractsDir + `"`,
	}

	for _, bc := range bcs {
		self := bc.importPrefix()
		walkGoFiles(t, bc.dir(root), func(path string) {
			rel, _ := filepath.Rel(root, path)
			for _, imp := range parseImports(t, path) {
				if !strings.HasPrefix(imp, modulePath+"/"+internalDir+"/") {
					continue
				}
				// Own BC is fine.
				if imp == self || strings.HasPrefix(imp, self+"/") {
					continue
				}
				// platform/ and contracts/ are allowed.
				ok := false
				for _, p := range allowedInternalPrefixes {
					needle := strings.TrimSuffix(strings.TrimSuffix(p, "/"), `"`)
					if imp == needle || strings.HasPrefix(imp, needle+"/") {
						ok = true
						break
					}
				}
				if !ok {
					t.Errorf("BC may only depend on platform/ or contracts/:\n  file: %s\n  imp:  %s",
						rel, imp)
				}
			}
		})
	}
}
