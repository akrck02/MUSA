package style

import (
	"go/ast"
	"path"
	"slices"
	"strings"
	"testing"
)

// allowedShortNames are the short names the guide allows: what Go and everyone call them ("id" is
// a word in English too, and Go writes it ID).
var allowedShortNames = []string{"err", "id", "x", "y", "z", "ui", "_"}

// shortWords are English words of one or two letters: words, not abbreviations.
var shortWords = []string{"at", "by", "to", "up", "is", "of", "in", "on", "or", "as", "be", "do", "go", "if",
	"it", "no", "an"}

// abbreviations are shortenings the guide forbids, as a name or as one word of a name.
var abbreviations = []string{
	"pos", "ctx", "msg", "btn", "idx", "cfg", "conf", "tmp", "temp", "val", "num", "str", "ptr", "img",
	"buf", "src", "dst", "req", "res", "resp", "cnt", "arr", "obj", "elem", "attr", "param", "args",
	"calc", "util", "mgr", "impl", "prev", "curr", "cur", "fn", "cb", "db", "qty",
}

// longestFunction is the most lines a function may take, its signature and braces counted: the
// guide's "about 40", with some room.
const longestFunction = 50

// TestNamesAreWholeWords: a name a newcomer has to decode costs every reader, every time.
func TestNamesAreWholeWords(test *testing.T) {
	for _, file := range repositoryFiles(test) {
		for _, name := range declaredNames(file.syntax) {
			if problem := nameProblem(name.Name); problem != "" {
				test.Errorf("%s: %q %s", file.where(name), name.Name, problem)
			}
		}
	}
}

// nameProblem says what is wrong with a name, or "" when nothing is.
func nameProblem(name string) string {
	lower := strings.ToLower(name)
	if slices.Contains(allowedShortNames, lower) || slices.Contains(shortWords, lower) {
		return ""
	}
	if len(name) < 3 {
		return "is too short to say what it is"
	}
	for _, word := range camelCaseWords(name) {
		if slices.Contains(abbreviations, word) {
			return "holds the abbreviation " + word
		}
	}
	return ""
}

// TestPublicNamesSayWhatTheyAre: every exported name has a comment that starts with the name, so
// go doc and the reference explain the whole public surface.
func TestPublicNamesSayWhatTheyAre(test *testing.T) {
	for _, file := range repositoryFiles(test) {
		if file.isTest() || file.syntax.Name.Name == "main" {
			continue
		}
		for _, declared := range publicDeclarations(file.syntax) {
			if !strings.HasPrefix(declared.comment, declared.name.Name+" ") {
				test.Errorf("%s: %s has no comment that starts with its name", file.where(declared.name), declared.name.Name)
			}
		}
	}
}

// TestFunctionsAreShort: a long function is usually doing two jobs.
func TestFunctionsAreShort(test *testing.T) {
	for _, file := range repositoryFiles(test) {
		if file.isTest() {
			continue
		}
		for _, declaration := range file.syntax.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			lines := file.files.Position(function.End()).Line - file.files.Position(function.Pos()).Line + 1
			if lines > longestFunction {
				test.Errorf("%s: %s takes %d lines; split it (at most %d)", file.where(function), function.Name.Name, lines, longestFunction)
			}
		}
	}
}

// TestNoElseAfterAReturn: return early, and the rest of the function reads straight down.
func TestNoElseAfterAReturn(test *testing.T) {
	for _, file := range repositoryFiles(test) {
		ast.Inspect(file.syntax, func(node ast.Node) bool {
			condition, isIf := node.(*ast.IfStmt)
			if !isIf || condition.Else == nil || len(condition.Body.List) == 0 {
				return true
			}
			if _, returns := condition.Body.List[len(condition.Body.List)-1].(*ast.ReturnStmt); returns {
				test.Errorf("%s: an else after a return: return early instead", file.where(condition.Else))
			}
			return true
		})
	}
}

// TestErrorsAreMadeInErrorsFiles: errors carry codes, declared in each package's errors.go; an
// error made anywhere else would carry only text, which no program should check.
func TestErrorsAreMadeInErrorsFiles(test *testing.T) {
	for _, file := range repositoryFiles(test) {
		if file.isTest() || path.Base(file.path) == "errors.go" || file.syntax.Name.Name == "main" {
			continue
		}
		ast.Inspect(file.syntax, func(node ast.Node) bool {
			if called := calledFunction(node); called == "errors.New" || called == "fmt.Errorf" {
				test.Errorf("%s: %s outside errors.go makes an error without a code", file.where(node), called)
			}
			return true
		})
	}
}

// TestNoMapOfBoolAsAChoice: map[bool]string{true: …, false: …} is an if written as a puzzle.
func TestNoMapOfBoolAsAChoice(test *testing.T) {
	for _, file := range repositoryFiles(test) {
		ast.Inspect(file.syntax, func(node ast.Node) bool {
			mapType, isMap := node.(*ast.MapType)
			if !isMap {
				return true
			}
			if key, isName := mapType.Key.(*ast.Ident); isName && key.Name == "bool" {
				test.Errorf("%s: a map keyed by bool: write an if", file.where(node))
			}
			return true
		})
	}
}

// calledFunction is "package.Function" when node calls a function of an imported package.
func calledFunction(node ast.Node) string {
	call, isCall := node.(*ast.CallExpr)
	if !isCall {
		return ""
	}
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return ""
	}
	owner, isName := selector.X.(*ast.Ident)
	if !isName {
		return ""
	}
	return owner.Name + "." + selector.Sel.Name
}
