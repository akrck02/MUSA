# 📐 Style guide
---

This page is for everyone who writes code in MUSA: by the end you know how the code is written,
and why, so three people's code reads like one person's.

The rules come from the author's guidelines for every project, applied to MUSA. The test package
`internal/style` checks the ones a program can check; the table at the end says which.

## 1. English in the code, Spanish on the screen

Code, comments, names, error codes, commit messages and docs are in English. What the artist reads
is Spanish. A field's label and hint sit beside the field they name; sentences, notes, status
words and the words of each problem live in the package's `words.go`.

```go
// No: Spanish names, and a one-letter receiver
func (s *Screens) guardar() { s.nota = "Guardado" }

// Yes (catalogue/screens.go has tell)
func (screens *Screens) save() { screens.tell(screens.store.Save(screens.draft), "Guardado") }
```

## 2. Names say what things are

- **Whole words.** `position`, not `pos`; `index`, not `i`; `context`, not `ctx`; `quantity`,
  not `qty`. Short English words (`at`, `to`, `up`) are words, not abbreviations.
- **The few exceptions** are what everyone calls them: `err`, `id` (written `ID` in a longer
  name, as Go does), the axes `x`, `y`, `z`, and `ui` for Go UI's package.
- **Tests** call their `*testing.T` `test`; a helper that builds what a test needs is a `fixture`.
- **Say the unit:** `pictureHeight` in a block titled "in pixels", `noteTime` as a
  `time.Duration`, prices as `money.Money`, never a bare `int` of cents.
- **Booleans read as a question:** `archived`, `found`, `readable`. Not `ok`: it does not say
  what is fine.
- **Functions are verbs, or what they give back:** `Save`, `MoveStock`, `PriceWords`.
- **Don't repeat the package:** `catalogue.Store`, not `catalogue.CatalogueStore`.

## 3. No magic numbers

A number that means something gets a named constant, with its unit. Screens group their measures
at the top of the file, as the design gives them:

```go
// Measures of a product tile, in pixels, as in the author's design.
const (
	tileRadius    = 18
	pictureHeight = 132
	lowStock      = 3 // from this many left down, the tile warns
)
```

## 4. One function, one job

- A function does one thing, and its name says which. About 40 lines at most.
- If it needs comments to split it into parts, split it into functions instead.
- A function that finds something does not also change it.
- A screen that grows long is split by the parts a person sees: `drawEditor` calls `whereAndWho`,
  `lineSection` and the summary.

## 5. Early returns, no else after a return

Handle the errors and the special cases first, and return. The normal path runs down the left
margin.

```go
// No
if days == 1 {
	return "Mañana"
} else {
	return fmt.Sprintf("Dentro de %d días", days)
}

// Yes
if days == 1 {
	return "Mañana"
}
return fmt.Sprintf("Dentro de %d días", days)
```

## 6. An if is an if

Choosing between two values is written as an `if`, in a function with a name, never as a map
keyed by a boolean:

```go
// No
context.Text(map[bool]string{true: "Restaurar", false: "Archivar"}[archived], style)

// Yes, in words.go
func archiveWords(archived bool) string {
	if archived {
		return "Restaurar"
	}
	return "Archivar"
}
```

The map builds a table every frame to throw it away, and hides a simple choice behind a puzzle.

## 7. Errors have codes

Every error a package returns carries a stable code, so programs and tests check the code, never
the text.

- **Each package declares its codes once,** at the top of its `errors.go`, named `package.what`:
  `catalogue.blank-name`, `demo.not-empty`.
- **Problems** are what the artist can fix (a blank name): the store returns their `Code`, and
  `words.go` says it in Spanish. **Failures** are what the artist cannot fix (the disk is full):
  an `*Error` with the code, a detail and the cause underneath.
- **Errors are made only in `errors.go`,** by small constructors (`productNotFound(id)`). There is
  no `errors.New` or `fmt.Errorf` anywhere else, except in `cmd/musa`'s `main`, which reports to
  a person and exits.
- **Check codes** with `errors.As` or the package's helper (`demo.CodeOf(err)`, which uses it), never
  with `err.Error()`.
- **Never drop an error silently.** When ignoring one is right, say why on that line:

  ```go
  event, _, _ := fairs.Find(id) // a fair that cannot be read is named by its channel instead
  ```

## 8. Comments say why

- Every public name has a comment that starts with the name: `// Store keeps the catalogue…`.
- A package's first file starts with what the package is for.
- A comment says why, or what the code cannot show; it does not repeat the code.
- When you change code, check that its comments still tell the truth.
- Unfinished work says so: `// TODO(name): what is missing`.

## 9. Screens

- **Everything clickable has a label** a test can find (`context.Describe` when two would read
  the same), and that label is what a screen reader says. A sign alone, such as «−», says nothing.
- **Colours come from the theme** (`theme.Primary`, `theme.InkSoft`), never written by hand, so
  the dark theme works without a second look.
- **SQL lives only in the store's files** (`store.go`, and the files beside it that read rows);
  a screen asks its store.
- **A frame allocates nothing it can avoid:** keep what does not change between frames (the
  tiles' shapes, read data) in the package's `Screens`, and read the store only after a change.

## 10. Packages and files

- A package is an idea the artist works with, and its name is a good prefix for what it holds:
  `stand.Basket`, `events.StockLine`. No `util`, `common`, `helpers`, `models`, `types`.
- Dependencies point one way: the [code map](1-the-code.md) draws them and a test keeps them.
- Divide a package into files by part (`tile.go`, `basket.go`, `errors.go`), not into more
  packages.
- Tests sit beside the code they test.

## 11. Tests

- Every change in behaviour comes with its test.
- A test's name is a sentence about what must be true:
  `TestASaleOpenedFromTheHistoryCanBeDeleted`.
- A test does not depend on today's date or on the order of a map: give stores a fixed `now`,
  and compute dates from the same clock as the data.
- [Testing](4-testing.md) shows each kind of test.

## 12. Format and checks

Before every commit:

```bash
gofmt -l .        # prints nothing
go vet ./...
go test ./...
```

## What the tests check

| Rule | Test in `internal/style` (or `docs`) |
|---|---|
| Whole-word names | `TestNamesAreWholeWords` |
| A comment on every public name | `TestPublicNamesSayWhatTheyAre` |
| About 40 lines per function (50 with some room) | `TestFunctionsAreShort` |
| No else after a return | `TestNoElseAfterAReturn` |
| Errors only in `errors.go` | `TestErrorsAreMadeInErrorsFiles` |
| No map keyed by a boolean | `TestNoMapOfBoolAsAChoice` |
| Dependencies point one way | `TestPackagesImportOnlyWhatTheMapAllows` (in `docs`) |
| No broken link in the docs | `TestEveryDocsLinkLeadsToAFile` (in `docs`) |

The rest (units, why-comments, one job per function) is for the reviewer: a program cannot judge
them.
