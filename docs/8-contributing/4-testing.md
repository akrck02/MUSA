# 🧪 Testing
---

This page is for whoever changes behaviour in MUSA: by the end you know which kind of test to
write, where it goes, and how to run and watch it.

Every change in behaviour comes with a test. A test exists to prove something a person relies
on, so its name is a sentence about what must be true: `TestPayingInCashAtTheStandRecordsALocalSaleAndTakesTheStock`.

## Run them

```bash
gofmt -l .                       # prints nothing when everything is formatted
go vet ./...
go test ./...                    # everything; no display needed
go test ./stand -run Basket      # the tests of one package whose name matches
```

No test needs a screen: the screen tests draw into memory, with a fake text measurer.
On Linux, add `-tags=x11` after `run`, `vet` and `test` (see
[troubleshooting](../7-troubleshooting.md)).

## Four kinds of test

| Kind | Proves | Lives in | Example |
|---|---|---|---|
| Rules | A calculation or a rule, with no database | beside the code | `money/money_test.go`, `stand/basket_test.go` |
| Store | What is saved, read and refused | beside the store | `events/events_test.go` |
| Flow | What a person can do, through the whole app | `app/*_flow_test.go` | `app/stand_flow_test.go` |
| Docs and style | The docs' links, the code map, the style guide | `docs/`, `internal/style` | `internal/style/style_test.go` |

### Rules

Plain Go tests of plain functions. When several cases cover one behaviour, write a table:

```go
func TestLeftWordsWarnWhenFewAreLeft(test *testing.T) {
	cases := []struct {
		left  int
		words string
	}{{0, "Agotado"}, {2, "Quedan 2"}, {9, "9"}}
	for _, each := range cases {
		if words := leftWords(each.left); words != each.words {
			test.Errorf("%d left: %q, want %q", each.left, words, each.words)
		}
	}
}
```

### Store

A store test opens a database in memory, with a clock and ids that never change, so every run
sees the same data:

```go
database, err := storage.Open(storage.InMemory)
if err != nil {
	test.Fatal(err)
}
test.Cleanup(func() { database.Close() })
counter := 0
ids := func() string { counter++; return fmt.Sprint("id", counter) }
now := func() time.Time { return time.UnixMilli(1_700_000_000_000) }
products := catalogue.NewStore(database, now, ids)
```

Packages that need several stores build them once in a `newBench(test)` helper and call the
result a `fixture`. Check problems and failures by their code, never by their text.

### Flow

A flow test opens the whole app, with or without the demo's data, in a window that lives in
memory (Go UI's `uitest`), and uses it as a person would: by the labels on screen.

```go
func TestCreatingAProductPutsItInTheCatalogue(test *testing.T) {
	program, window := openApp(test, false) // false: an empty database
	program.Open(Catalogue)
	window.Frame()
	steps(test, window,
		click(window, "Crear tu primer producto"),
		click(window, "Producto físico"),
		typeInto(window, "Nombre del producto", "Lámina de prueba"),
		click(window, "Guardar"),
	)
	products, _ := program.stores.Products.All()
	// …check what was saved, then what the screen shows:
	if _, err := window.Element("Lámina de prueba"); err != nil {
		test.Error("the new product is not in the list:", err)
	}
}
```

The helpers live in `app/app_test.go`: `openApp`, `steps`, `click`, `typeInto`, and `settle`,
which lets animations finish. A few things to know:

- **Find by label.** If a label is not found, the error lists what the frame showed.
- **Two controls with one label** make the test ambiguous: give them unique labels (see
  [screens](3-screens.md#labels-for-tests-and-for-people)).
- **The demo dates its data from today.** Compute dates in the test from the same clock
  (`calendar.Of(time.Now())`), never write "23 oct".
- **Check the store too,** not only the screen: what was saved is what matters.

`app/explore_test.go` opens every page of the app and presses what can be pressed, looking for
crashes, and `app/phone_test.go` checks that every page fits a phone: add a new page to their
list.

## Watch them

Go UI's results viewer runs the tests with recording on, then shows each one in a window: which
passed, what it wrote, and a frame-by-frame replay with the pointer and the element each step
found.

```bash
go run github.com/akrck02/go-ui/cmd/uitest ./...
go run github.com/akrck02/go-ui/cmd/uitest -dark ./app -run TestStand
```

It needs a display (and `-tags=x11` after `run` on Linux). A test is recorded when its window is
made with `uitest.Test(test)`, as `openApp` does.

## What still needs a person

Tests draw into memory, so they cannot see:

- how the real window looks on each system (fonts, the dark theme, Retina screens);
- resizing the window, and moving it between screens;
- the touch of a real finger on the iPad or Android.

After a visual change, open the app, or take a picture of the page with `-capture` (see
[screens](3-screens.md#seeing-it)), and say in the pull request what you looked at.
