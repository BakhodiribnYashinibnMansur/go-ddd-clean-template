package admin

import (
	"html/template"
	"os"
	"path/filepath"
	"testing"
)

func TestTemplatesParse(t *testing.T) {
	// Change to project root so template paths resolve
	if err := os.Chdir(findProjectRoot(t)); err != nil {
		t.Fatalf("failed to chdir to project root: %v", err)
	}

	baseFiles := []string{
		"internal/web/admin/templates/layout/base.html",
		"internal/web/admin/templates/layout/header.html",
		"internal/web/admin/templates/layout/sidebar.html",
		"internal/web/admin/templates/layout/pagination.html",
	}

	// Collect partials
	partials, _ := filepath.Glob("internal/web/admin/templates/partials/*.html")
	nestedPartials, _ := filepath.Glob("internal/web/admin/templates/partials/*/*.html")
	partials = append(partials, nestedPartials...)

	// All page templates to test
	pages, err := filepath.Glob("internal/web/admin/templates/pages/*.html")
	if err != nil {
		t.Fatalf("failed to glob pages: %v", err)
	}
	nestedPages, _ := filepath.Glob("internal/web/admin/templates/pages/*/*.html")
	pages = append(pages, nestedPages...)

	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			files := make([]string, 0, len(baseFiles)+len(partials)+1)
			files = append(files, baseFiles...)
			files = append(files, partials...)
			files = append(files, page)

			_, err := template.New("base").Funcs(templateFuncs).ParseFiles(files...)
			if err != nil {
				t.Errorf("template parse failed for %s: %v", page, err)
			}
		})
	}
}

func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod)")
		}
		dir = parent
	}
}
