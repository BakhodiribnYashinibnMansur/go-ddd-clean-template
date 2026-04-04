// Package arch enforces DDD bounded context architecture rules at test time.
// These tests complement the depguard linter rules in .golangci.yml and
// provide a single source of truth for cross-BC isolation and layering.
package arch_test

import (
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
	sharedBC    = "shared"
	// appBC is the composition root — it wires all BCs together via DI.
	// By definition it is allowed to depend on every other BC.
	appBC = "app"
)

// layer represents a DDD layer within a bounded context.
type layer string

const (
	layerDomain         layer = "domain"
	layerApplication    layer = "application"
	layerInfrastructure layer = "infrastructure"
	layerInterfaces     layer = "interfaces"
)

// layerRules defines which layers each layer is allowed to depend on
// WITHIN the same bounded context. Cross-BC imports are handled separately.
var layerRules = map[layer]map[layer]bool{
	layerDomain: {
		// domain depends on nothing inside the BC
	},
	layerApplication: {
		layerDomain: true,
	},
	layerInfrastructure: {
		layerDomain:      true,
		layerApplication: true,
	},
	layerInterfaces: {
		layerDomain:      true,
		layerApplication: true,
	},
}

// knownViolations lists existing architecture violations that are tolerated
// until fixed. Each entry is "source_file -> imported_package". DO NOT add
// new entries without team approval. Remove entries as they are fixed.
var knownViolations = map[string]string{
	// TODO: move audit job handlers out of shared infrastructure into audit BC
	"internal/shared/infrastructure/asynq/handlers.go": "gct/internal/audit",
	// TODO: move cache reads behind a domain-defined port
	"internal/featureflag/application/query/evaluate.go": "gct/internal/featureflag/infrastructure/cache",
	// TODO: replace direct user query with an authz-owned port + event sync
	"internal/authz/interfaces/http/middleware/authz.go": "gct/internal/user/application/query",
}

// isKnownViolation reports whether a file->import pair is an accepted exception.
func isKnownViolation(relFile, imp string) bool {
	prefix, ok := knownViolations[filepath.ToSlash(relFile)]
	return ok && strings.HasPrefix(imp, prefix)
}

// forbiddenInDomain lists external packages that MUST NOT appear in domain layer.
// Domain must remain pure business logic.
var forbiddenInDomain = []string{
	"github.com/labstack/echo",
	"gorm.io/gorm",
	"github.com/jackc/pgx",
	"github.com/redis/go-redis",
	"github.com/minio/minio-go",
	"github.com/golang-jwt/jwt",
	"net/http",
	"database/sql",
}

// repoRoot finds the repository root by walking up from the test file.
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

// boundedContexts lists all BCs under internal/, excluding shared.
func boundedContexts(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, internalDir))
	if err != nil {
		t.Fatalf("read internal dir: %v", err)
	}
	var bcs []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != sharedBC {
			bcs = append(bcs, e.Name())
		}
	}
	return bcs
}

// parseImports returns the import paths of a Go file.
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

// walkGoFiles invokes fn for every non-test .go file under dir.
func walkGoFiles(t *testing.T, dir string, fn func(path string)) {
	t.Helper()
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fn(path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
}

// detectLayer extracts the layer name from a path under internal/<bc>/<layer>/...
// Returns empty string if path is not within a layered BC.
func detectLayer(rel string) (bc string, l layer) {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) < 3 || parts[0] != internalDir {
		return "", ""
	}
	return parts[1], layer(parts[2])
}

// TestBoundedContextIsolation verifies no BC imports from another BC except shared.
func TestBoundedContextIsolation(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	bcs := boundedContexts(t, root)
	bcSet := make(map[string]struct{}, len(bcs))
	for _, b := range bcs {
		bcSet[b] = struct{}{}
	}

	for _, bc := range bcs {
		if bc == appBC {
			continue // composition root — allowed to import all BCs for DI wiring
		}
		bcDir := filepath.Join(root, internalDir, bc)
		walkGoFiles(t, bcDir, func(path string) {
			rel, _ := filepath.Rel(root, path)
			for _, imp := range parseImports(t, path) {
				if !strings.HasPrefix(imp, modulePath+"/"+internalDir+"/") {
					continue
				}
				// extract the BC of the import
				tail := strings.TrimPrefix(imp, modulePath+"/"+internalDir+"/")
				impBC := strings.SplitN(tail, "/", 2)[0]
				if impBC == sharedBC || impBC == bc {
					continue
				}
				if _, isBC := bcSet[impBC]; isBC {
					if isKnownViolation(rel, imp) {
						continue
					}
					t.Errorf("cross-BC import forbidden:\n  file: %s\n  BC:   %s\n  imp:  %s\n  fix:  route through gct/internal/shared or use integration events",
						rel, bc, imp)
				}
			}
		})
	}
}

// TestLayerDependencies verifies DDD layering rules within each BC.
func TestLayerDependencies(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	bcs := boundedContexts(t, root)

	for _, bc := range bcs {
		bcDir := filepath.Join(root, internalDir, bc)
		walkGoFiles(t, bcDir, func(path string) {
			rel, _ := filepath.Rel(root, path)
			_, currentLayer := detectLayer(rel)
			allowed, known := layerRules[currentLayer]
			if !known {
				return
			}
			for _, imp := range parseImports(t, path) {
				prefix := modulePath + "/" + internalDir + "/" + bc + "/"
				if !strings.HasPrefix(imp, prefix) {
					continue
				}
				tail := strings.TrimPrefix(imp, prefix)
				impLayer := layer(strings.SplitN(tail, "/", 2)[0])
				if impLayer == currentLayer {
					continue
				}
				if !allowed[impLayer] {
					if isKnownViolation(rel, imp) {
						continue
					}
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
		domainDir := filepath.Join(root, internalDir, bc, string(layerDomain))
		if _, err := os.Stat(domainDir); err != nil {
			continue
		}
		walkGoFiles(t, domainDir, func(path string) {
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

// TestSharedDomainHasNoBCDeps verifies shared kernel does not depend on any BC.
func TestSharedDomainHasNoBCDeps(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	sharedDir := filepath.Join(root, internalDir, sharedBC)

	walkGoFiles(t, sharedDir, func(path string) {
		rel, _ := filepath.Rel(root, path)
		for _, imp := range parseImports(t, path) {
			if !strings.HasPrefix(imp, modulePath+"/"+internalDir+"/") {
				continue
			}
			tail := strings.TrimPrefix(imp, modulePath+"/"+internalDir+"/")
			impBC := strings.SplitN(tail, "/", 2)[0]
			if impBC != sharedBC {
				if isKnownViolation(rel, imp) {
					continue
				}
				t.Errorf("shared kernel must not depend on BC:\n  file: %s\n  imp:  %s",
					rel, imp)
			}
		}
	})
}
