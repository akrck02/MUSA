# 🖼️ Screens
---

This page is for whoever adds or changes a screen: by the end you know how a page is drawn, where
its state lives, how it opens another page, and what makes it testable.

## Drawn every frame

Go UI is immediate mode. Nothing on screen is an object you keep: every frame, a function walks
through what should be visible and draws it. The call that draws a control also answers whether
it was used in this frame.

This is how the empty state that every list shows is used:

```go
if look.Empty(context, "package", "Aún no hay productos", "Crea el primero para empezar a vender.", "Crear producto") {
	screens.newProduct()
}
```

`look/example_test.go` runs this exact code in a test window and clicks the button, so the example
cannot go stale.

Picture a whiteboard you wipe and redraw sixty times a second, rather than a corkboard of cards
you pin and move. Nothing on it can fall out of step with the data, because nothing on it
survives a frame.

## Where the state lives

What must survive between frames lives in the package's `Screens`: which page or sheet is open,
what is being typed, and what was read from the store. In `catalogue/screens.go`:

```go
type Screens struct {
	store    *Store
	products []Product // read from the store, drawn every frame
	query    Query     // the tab and the search
	stockOf  string    // the product whose quick stock sheet is open
	editor   *editor   // nil while the list shows
	// …
}
```

Three habits keep it simple:

1. **Read the store only after a change,** never in every frame: a frame must be cheap.
2. **A page or sheet is open when its field is set** (`editor != nil`, `stockOf != ""`); closing it
   is setting the field back.
3. **Keep only what you cannot read again.** If it can come from the store, read it there.

## One package, its pages

`Screens.Draw(context)` is the package's only door: `app` calls it for the place that is open.
Inside, `Draw` chooses the page:

```go
func (screens *Screens) Draw(context *ui.Context) {
	if screens.editor != nil {
		screens.drawEditor(context)
	} else {
		screens.drawList(context)
	}
	// …the note of the last action, as a toast
}
```

Each page or sheet has a file of its own (`list_screen.go`, `editor_screen.go`, `tags_sheet.go`),
with its measures as named constants at the top.

## Moving between places

A package never imports `app`. When `events` needs to open the stand, it asks for a function, and
`app` gives it one when it builds the screens (`app/app.go`, `buildScreens`):

```go
program.fairs.OpenStandWith(func(event events.Event) {
	program.Open(Stand)
	program.stand.Choose(stand.Fair{Event: event, Events: fairs, Products: products})
})
```

To add a new place: a `Place` in `app/app.go`, a case in `page`, a button in the frame
(`app/frame.go`), and, for captures and tests, a name in `Show`.

## The look

The look is Bubble 4's, through Go UI's theme and the pieces in `look`:

| Use | For |
|---|---|
| `look.PageHeader` | A page's title, subtitle and its buttons on the right |
| `look.Section`, `look.Card` | A part of a form or a detail page |
| `look.Field` | A labelled text field, with the problem under it |
| `look.Empty` | A page with nothing yet: what is missing and the action that fixes it |
| `look.Segmented`, `look.Search` | Tabs, and a search a button opens |
| `context.Sheet` | A sheet over the page: a question, a small form |

Colours, sizes and radii come from `context.Theme` (`theme.Primary`, `theme.Gap`), never written
by hand, so the dark theme comes for free. A piece that two packages need goes into `look`; a piece
that any Go UI program could use belongs in Go UI itself, through a proposal there.

## Labels: for tests and for people

Every control has a label: the text of a button, or one given with `context.Describe`. Tests find
controls by these labels, and a screen reader says them. When two controls read the same («Quitar
uno» on every tile), make the label unique (`"Quitar uno##"+id`: what follows `##` is not drawn)
or describe it: `context.Describe(ui.Description{Role: ui.ButtonRole, Label: "Quitar uno de Luna"})`.

## Seeing it

```bash
go run ./cmd/musa -demo                                 # the app with toy data
go run ./cmd/musa -demo -place 4                        # open on the stand
go run ./cmd/musa -demo -show product -capture out.png  # one frame of a page, to a picture
go run ./cmd/musa -demo -dark -capture out.png          # the same in the dark theme
```

On Linux, add `-tags=x11` after `run`. `-show` takes the pages listed in `app.Show`.
