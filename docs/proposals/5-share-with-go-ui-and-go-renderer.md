# 5. What MUSA should hand to Go UI and Go Renderer

> Status: approved in part · Author's approval: proposal ◐ plan ☐ · Lives in: `go-renderer` (`draw2d`, the
> painters), `go-ui` (controls, motion, `uitest`), MUSA's `look` and every screen package
> Depends on: proposals 1 to 4 (MUSA as it is on `feature/go-ui`)

## Why

MUSA was built screen by screen against the author's design. Along the way it grew pieces that
are not about an artist's stock at all: a search that opens from a button, a tab whose pill
slides, a confirmation sheet, a page that rises into view, a box rounded on one corner only.
Each was written in MUSA's `look` package, or copied inside a screen, because Go UI did not
have it yet.

Two things show it is time to move them. First, counting: the same few lines appear again and
again (table below). Second, testing: pressing every control of every page (the new
`TestEveryControlOnEveryPageCanBePressed`) found bugs that live in those copies, not in MUSA's
domain. A bug fixed once in a library is fixed for every program; a bug fixed in a copy waits
in the next copy.

| Pattern in MUSA today | Copies | What it costs |
|---|---|---|
| `Clickable(…)` followed by `Describe(ButtonRole, same label)` so a screen reader names it | 26 | forget it and the control is read as «−» or by its icon; the stand's «Quitar uno» was |
| `ui.Looks{Normal: x, Hovered: x, Pressed: x}`: the same look three times | 31 | leave `Hovered` out and the card loses its fill under the pointer |
| a `tell` function, a note, its `noteUntil` time and a `RequestFrame` to hide it | 6, five different signatures | each screen reports results its own way |
| a sheet asking «¿Eliminar…?» with Cancelar and a red button | 2, and 2 missing | deleting a sale or a project happens on one tap, deleting a product asks |
| a search field opened by a button | 3 | the stand's «Buscar» did nothing; the others opened without the keyboard |
| `context.Appear` + `look.Arriving` + `LayerOptions{Hidden, Leave}` for a dialog | 4 | each dialog wires its own entrance and exit |
| a painter kept in a map «so drawing it allocates nothing» (`Corners`, `TopRounded`) | 3 | a cache per screen to work around a missing shape |
| `Fade` drawn as one rectangle per pixel of height | 1 | hundreds of rectangles every frame for one gradient |

## What

Three homes, by what a piece is about:

- **Go Renderer** draws: shapes and colours, nothing that knows a control.
- **Go UI** behaves: controls, focus, motion, what a screen reader and a test see. It takes its
  sizes and colours from `ui.Theme`, so Bubble (or any design system) dresses it.
- **MUSA** keeps what is about the artist: catalogue, fairs, stand, orders, projects, and the
  Bubble 4 sizes of the author's design.

### Go Renderer

| Id | Requirement | Check |
|---|---|---|
| R1 | A rectangle can round each corner by its own radius (0 for square), filled and bordered, with no allocation per frame | test `a rectangle rounded on one corner leaves the others square`; `testing.AllocsPerRun` = 0 |
| R2 | A rectangle can be filled with a linear gradient between two or more colour stops, top to bottom or left to right, in one command | test `a vertical gradient is clear at the top and solid at its last stop`; one command in the list |
| R3 | The Metal window waits for the interface to settle before a capture, as OpenGL does (`frameloop.Settle`), proved on a Mac | run on a Mac: a capture taken during an animation shows it finished |

With R1 and R2, MUSA's `look.Corners`, `look.TopRounded` and `look.Fade` disappear, and so do the
maps that keep them.

### Go UI

| Id | Requirement | Check |
|---|---|---|
| R4 | A `Clickable` is read by a screen reader by its label's shown words when it has some; `Describe` still overrides. Cards that want their whole content read say so | test `a clickable showing a sign is read by its label`; MUSA's 26 `Describe` calls that repeat the label go |
| R5 | In `ui.Looks`, a `Hovered` or `Pressed` left empty takes `Normal` (a touch screen has no hover) | test `a clickable with only a normal look keeps it under the pointer` |
| R6 | `Context.Notify(text, warning)` shows a toast for a theme-given time and hides it by itself, asking for the frames it needs | test with the test clock: `a note hides itself after its time` |
| R7 | `Context.Confirm(name, question, explanation, confirmWords, destructive) (confirmed, cancelled bool)` is a sheet with «Cancel» and a confirm button, red when destructive | test `a destructive question confirms only on its button` |
| R8 | A search box: a button that opens a field ready to type in, and closes it forgetting the text (`ui.Search`, as `look.Search` today) | test `a search opens with the keyboard in its field`; `closing a search forgets its text` |
| R9 | `Tabs` draws its chosen segment as one pill that slides to the newly chosen tab and stays on its row while the page scrolls; pill and segmented shapes come from the theme | the two tests of `look/slide_test.go`, moved |
| R10 | `Modal` and `Sheet` rise in and sink out by themselves, with the theme's motion; `LayerOptions.Still` turns it off | test `a modal is half shown half way through its motion` |
| R11 | `Context.Page(name, content)`: when `name` changes, the content rises in; a dialog opening over it does not change it | test `a dialog over a page leaves the page still` (MUSA's test, moved) |
| R12 | `Context.Breakpoint()` names the window's width class (narrow, regular, wide) from theme widths; `Columns(wanted)` gives fewer columns on narrow windows | test per class; MUSA's `Narrow`, `Wide`, `Columns`, `Across` go |
| R13 | `uitest.Window.PressEverything(open func(), check func(label string))` opens a fresh state, presses each control the screen reader sees, one at a time, and runs `check` after each | MUSA's explorer test becomes a few lines; a test in `uitest` with a control that panics |
| R14 | `uitest` finds a long-labelled control by the start of its label (`ElementStarting`), so a card read whole is found by its title | test `a card is found by its title` |

### MUSA

| Id | Requirement | Check |
|---|---|---|
| R15 | Every destructive action asks first through `Confirm`: delete a sale, a project, a product, a fair, empty the database | one test per action: `nothing is deleted before the answer` |
| R16 | Every screen reports what happened through `Notify`; the six `tell` functions go | `grep "func (screens \*Screens) tell"` finds nothing |
| R17 | Problems and refusals share one shape across packages: a code and the words for it (`catalogue.Problem`, `events.Refusal`, `projects.Code`, `sales.Problem` today) | test per package: `each code has its words`, as now, on one type |
| R18 | `look` holds only MUSA's design: the Bubble 4 sizes of headers, cards, pills, figures and the tab sizes of the author's mock-ups | `look` has no `Paint` methods and no motion code |
| R19 | `map[bool]string{…}[condition]` gives way to a plain `if`, which a newcomer reads without stopping | `grep "map\[bool\]"` finds nothing |

### The author's answers (2026-10-26)

| Requirements | Answer |
|---|---|
| R1–R3 (Go Renderer) | Approved |
| R4 (a control is read by its words) | Approved |
| R6, R7 (notes and confirmations) | Approved; the colour of a destructive button comes from the theme and can be changed, red by default |
| R9 (the sliding pill) | Approved, with its motion configurable: duration and easing from the theme, and a way to turn it off |
| R12 (width classes) | Approved |
| R13, R14 (`uitest`) | Approved; the test viewer also gains a replay at the host's speed (done in Go UI, `4288f2c`) |
| R5, R8, R10, R11, R15 | Not answered yet |

## Found while testing

Pressing every control of 23 pages (508 checks) and walking each flow found these; the ones
marked ✓ are fixed on `feature/go-ui` already:

1. ✓ Archiving a product crashed the editor: it closed itself while still drawing.
2. ✓ The stand's «Buscar» changed a value nothing read.
3. ✓ Opening a search did not give the field the keyboard: two taps to type.
4. ✓ The stand's «−» was read as «−» by a screen reader, and could not be found by a script.
5. ✓ Icon-only buttons labelled `"##Name"` were read by their icon («ellipsis», «undo 2») — fixed
   in Go UI for every program, and drag sources are read by their name.
6. ✓ Two controls were both «Añadir etiqueta» (the pill and its ✓).
7. ✓ Version messages said «variante» where MUSA says «versión».
8. ✓ The chosen tab's pill chased its row while the page scrolled.
9. Deleting a sale or a project happens on one tap, with no question (R15).
10. A project card is read twice by a screen reader: once as the thing to drag, once as the
    button that opens it (R4 decides how it is read once).
11. On the author's Mac, icons, tints and button words were missing: an old checkout of Go UI and
    Go Renderer beside MUSA. MUSA's README should say which branches it builds against (R3's
    Mac run is the occasion to check it).

## Out of scope

- Moving MUSA's domain (catalogue, fairs, sales, projects) anywhere: it is the app.
- A Kanban board control. MUSA's two boards use `DragSource` and `DropTarget`, which are generic
  already; a board becomes a control when a second program needs one.
- The month calendar as a Go UI control. It needs dates in many languages; MUSA's is in Spanish
  only. It moves when a second program wants it.
- Bubble 4's own adapter (`bubble/goui`): it keeps mapping Bubble's tokens onto `ui.Theme`, and
  gains the new theme fields R6, R9, R10 and R12 read.

## Design

*To fill after the author approves the proposal.* The order will follow dependencies: Go Renderer
first (R1–R3), then Go UI one requirement per commit with its test and docs, then MUSA takes each
piece and deletes its copy in the same commit.

## Tasks

*To fill with the design.*

## Result

*To fill at the end.*
