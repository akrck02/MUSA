package style

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"
)

// declaredNames are the names a file declares: variables, constants, types, functions, their
// parameters and results, fields, and import names.
func declaredNames(syntax *ast.File) []*ast.Ident {
	var names []*ast.Ident
	ast.Inspect(syntax, func(node ast.Node) bool {
		switch declared := node.(type) {
		case *ast.ValueSpec:
			names = append(names, declared.Names...)
		case *ast.TypeSpec:
			names = append(names, declared.Name)
		case *ast.FuncDecl:
			names = append(names, declared.Name)
		case *ast.Field:
			names = append(names, declared.Names...)
		case *ast.ImportSpec:
			if declared.Name != nil {
				names = append(names, declared.Name)
			}
		case *ast.AssignStmt:
			if declared.Tok == token.DEFINE {
				names = append(names, identifiers(declared.Lhs)...)
			}
		case *ast.RangeStmt:
			if declared.Tok == token.DEFINE {
				names = append(names, identifiers([]ast.Expr{declared.Key, declared.Value})...)
			}
		}
		return true
	})
	return names
}

// identifiers are the plain names among expressions.
func identifiers(expressions []ast.Expr) []*ast.Ident {
	var names []*ast.Ident
	for _, expression := range expressions {
		if name, isName := expression.(*ast.Ident); isName {
			names = append(names, name)
		}
	}
	return names
}

// camelCaseWords are the words of a camelCase or snake_case name, in lower case: "basketQty" is
// "basket" and "qty"; a run of capitals is one word, so "HTTPReq" is "http" and "req".
func camelCaseWords(name string) []string {
	var words []string
	start := 0
	letters := []rune(name)
	for index := 1; index < len(letters); index++ {
		if letters[index] == '_' {
			words = append(words, strings.ToLower(string(letters[start:index])))
			start = index + 1
			continue
		}
		if index > start && startsWord(letters, index) {
			words = append(words, strings.ToLower(string(letters[start:index])))
			start = index
		}
	}
	return append(words, strings.ToLower(string(letters[start:])))
}

// startsWord tells whether the capital at index begins a new word: after a small letter, or as the
// last capital of a run that a small letter follows ("Req" in "HTTPReq").
func startsWord(letters []rune, index int) bool {
	if !unicode.IsUpper(letters[index]) {
		return false
	}
	if !unicode.IsUpper(letters[index-1]) {
		return true
	}
	return index+1 < len(letters) && unicode.IsLower(letters[index+1])
}

// publicDeclaration is an exported name declared at the top of a file, and its comment's text.
type publicDeclaration struct {
	name    *ast.Ident
	comment string
}

// publicDeclarations are a file's exported functions, methods, types, constants and variables. A
// name in a group (const ( … )) may borrow the group's comment, as go doc shows it.
func publicDeclarations(syntax *ast.File) []publicDeclaration {
	var found []publicDeclaration
	for _, declaration := range syntax.Decls {
		switch declared := declaration.(type) {
		case *ast.FuncDecl:
			if declared.Name.IsExported() && receiverIsPublic(declared) {
				found = append(found, publicDeclaration{declared.Name, declared.Doc.Text()})
			}
		case *ast.GenDecl:
			found = append(found, groupDeclarations(declared)...)
		}
	}
	return found
}

// groupDeclarations are the exported names of a type, const or var declaration.
func groupDeclarations(group *ast.GenDecl) []publicDeclaration {
	var found []publicDeclaration
	for _, spec := range group.Specs {
		switch declared := spec.(type) {
		case *ast.TypeSpec:
			if declared.Name.IsExported() {
				found = append(found, publicDeclaration{declared.Name, commentOf(declared.Name, declared.Doc, group)})
			}
		case *ast.ValueSpec:
			for _, name := range declared.Names {
				if name.IsExported() {
					found = append(found, publicDeclaration{name, commentOf(name, declared.Doc, group)})
				}
			}
		}
	}
	return found
}

// commentOf is a declaration's own comment. Inside a group (const ( … )) with a comment of its
// own, a name without one is described by the group's, as go doc shows it, and passes.
func commentOf(name *ast.Ident, own *ast.CommentGroup, group *ast.GenDecl) string {
	if own != nil {
		return own.Text()
	}
	if group.Lparen.IsValid() && group.Doc != nil {
		return name.Name + " is described by its group's comment"
	}
	return group.Doc.Text()
}

// receiverIsPublic tells whether a method belongs to an exported type; a plain function counts.
func receiverIsPublic(function *ast.FuncDecl) bool {
	if function.Recv == nil || len(function.Recv.List) == 0 {
		return true
	}
	receiver := function.Recv.List[0].Type
	if pointer, isPointer := receiver.(*ast.StarExpr); isPointer {
		receiver = pointer.X
	}
	name, isName := receiver.(*ast.Ident)
	return isName && name.IsExported()
}
