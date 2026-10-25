package style

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// sourceFile is a parsed Go file of the repository.
type sourceFile struct {
	path   string // from the repository's root, with forward slashes
	syntax *ast.File
	files  *token.FileSet
}

// repositoryFiles parses every Go file of the repository, its tests included.
func repositoryFiles(test *testing.T) []sourceFile {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		test.Fatal(err)
	}
	var parsed []sourceFile
	files := token.NewFileSet()
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		syntax, err := parser.ParseFile(files, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		parsed = append(parsed, sourceFile{path: filepath.ToSlash(relative), syntax: syntax, files: files})
		return err
	})
	if walkErr != nil {
		test.Fatal(walkErr)
	}
	return parsed
}

// where is a node's file and line, as an editor opens them: "catalogue/store.go:42".
func (file sourceFile) where(node ast.Node) string {
	return file.path + ":" + strconv.Itoa(file.files.Position(node.Pos()).Line)
}

// isTest tells whether the file holds tests.
func (file sourceFile) isTest() bool { return strings.HasSuffix(file.path, "_test.go") }
