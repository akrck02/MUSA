package docs

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// module is the import path every package of MUSA starts with.
const module = "github.com/Itros97/MUSA/"

// allowedImports is the code map's table of dependencies (8-contributing/1-the-code.md): which of
// MUSA's packages each one may import. A change here is a change to the map, in the same commit.
var allowedImports = map[string][]string{
	"money":     nil,
	"calendar":  nil,
	"storage":   nil,
	"look":      nil,
	"cash":      {"money"},
	"customers": {"storage"},
	"catalogue": {"look", "money", "storage"},
	"sales":     {"calendar", "catalogue", "look", "money", "storage"},
	"events":    {"calendar", "catalogue", "look", "money", "sales", "storage"},
	"projects":  {"calendar", "catalogue", "customers", "look", "money", "sales", "storage"},
	"stand":     {"calendar", "cash", "catalogue", "events", "look", "money", "sales"},
	"demo":      {"calendar", "catalogue", "customers", "events", "money", "projects", "sales", "storage"},
	"app": {"calendar", "catalogue", "customers", "demo", "events", "look", "money", "projects", "sales",
		"stand", "storage"},
	"cmd/musa":       {"app", "demo", "storage"},
	"docs":           nil,
	"internal/style": nil,
}

// TestPackagesImportOnlyWhatTheMapAllows keeps dependencies pointing one way: a base idea that
// imported what is built on it would tie the two together and, sooner or later, into a cycle.
func TestPackagesImportOnlyWhatTheMapAllows(test *testing.T) {
	for _, folder := range packageFolders(test) {
		allowed, mapped := allowedImports[folder]
		if !mapped {
			test.Errorf("package %s is not in the code map", folder)
			continue
		}
		for _, imported := range moduleImports(test, folder) {
			if !slices.Contains(allowed, imported) {
				test.Errorf("%s imports %s, which the code map does not allow", folder, imported)
			}
		}
	}
}

// packageFolders are the module's folders that hold Go code, relative to its root.
func packageFolders(test *testing.T) []string {
	var folders []string
	err := filepath.WalkDir("..", func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		folder := filepath.ToSlash(filepath.Dir(strings.TrimPrefix(path, "../")))
		if !slices.Contains(folders, folder) {
			folders = append(folders, folder)
		}
		return nil
	})
	if err != nil {
		test.Fatal(err)
	}
	return folders
}

// moduleImports are the packages of MUSA that a folder's code imports; its tests may import more.
func moduleImports(test *testing.T, folder string) []string {
	files, err := filepath.Glob(filepath.Join("..", folder, "*.go"))
	if err != nil {
		test.Fatal(err)
	}
	var imports []string
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			test.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			path, _ := strconv.Unquote(spec.Path.Value) // the parser only accepts quoted paths
			if strings.HasPrefix(path, module) {
				imports = append(imports, strings.TrimPrefix(path, module))
			}
		}
	}
	return imports
}
