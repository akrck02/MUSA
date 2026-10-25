# 🗺️ The code
---

This page is for whoever is about to change MUSA: by the end you know where each thing lives,
which files a package usually has, and which packages may use which.

MUSA is one Go module, `github.com/Itros97/MUSA`. Read [core ideas](../3-core-ideas.md) first if
you have not: this page assumes them.

## The packages

| Package | What it holds |
|---|---|
| `money` | Euros as whole cents, and how they are written ("1.234,50 €"). |
| `calendar` | Days with no hour or time zone, in ISO text and in Spanish words. |
| `storage` | The SQLite file: opening it, the schema, its migrations, new ids and the artist's settings. |
| `look` | The pieces of Bubble 4 that every screen shares and Bubble does not give yet: cards, headers, tabs, search, empty states, motion. |
| `cash` | Paying in cash: what was handed over and the change. |
| `customers` | The people who order commissions, and their networks. |
| `catalogue` | Products, versions (variants), stock, tags and their icons. |
| `sales` | Orders from every channel, their status, and the sales history. |
| `events` | Fairs: the stock taken to them, their sales, expenses and profit. |
| `projects` | Commissions on a board of stages; delivering one records its sale. |
| `stand` | The stand: tiles, the basket, «Cobrar», and preparing an event's stock. |
| `demo` | A made-up studio to fill an empty database, through the same stores as the screens. |
| `app` | The frame, the navigation between places, the home page and its widgets. |
| `cmd/musa` | The program: reads its flags, opens the database, builds `app`, opens the window. |
| `docs` | No code: the tests that keep these pages true. |

## Which may use which

Dependencies point one way, from the top of this drawing down. A base idea never imports what is
built on it.

```
cmd/musa ─▶ app ─▶ stand ─▶ events ─▶ sales ─▶ catalogue ─▶ storage, money, look
                    │          projects ─▶ customers ─▶ storage
                    └▶ cash ─▶ money
            demo (fills every store)            calendar (used by most)
```

The exact table lives in `docs/dependencies_test.go`, and the test
`TestPackagesImportOnlyWhatTheMapAllows` fails when an import breaks it. If your feature needs a
new arrow, add it to that table in the same commit and say why in the pull request: a new arrow
is a design decision, not a detail.

Two arrows are missing on purpose:

- **No package imports `app`**: `app` connects the others. When `events` needs to open the stand,
  it asks for a function (`events.Screens.PrepareWith`) and `app` gives it one.
- **`look` imports nothing of MUSA**: it is the look of every screen, so it knows no screen.

## Inside a package

The packages that the artist works with (`catalogue`, `events`, `sales`, `projects`, `customers`)
share a shape. Taking `projects` as the example:

| File | What it holds |
|---|---|
| `project.go` | The package comment and the model: `Project`, its `Stage`, plain Go values. |
| `store.go` | `Store`: reading and writing the model in the database. It and the files beside it that read rows know SQL; nothing else does. |
| `errors.go` | The package's error codes (`projects.…`), declared once. |
| `words.go` | Everything the artist reads, in Spanish. |
| `screens.go` | `Screens`: the state the pages keep between frames, and the page that is open. |
| `board.go`, `editor.go` | One file per page or sheet, with the measures it uses as named constants. |
| `projects_test.go` | Tests of the rules and the store, on a database in memory. |

A file is a part of the package, not a kind of file. When a page grows past a few hundred lines,
split it into the parts a reader would name (`editor_layout.go`, `editor_screen.go`), not into new
packages.

## Where the screens are tested

The flow tests, which drive the whole app as a person would, live in `app/`
(`stand_flow_test.go`, `catalogue_flow_test.go`…), because only `app` has every package wired
together. Tests of one package's rules live beside its code. [Testing](4-testing.md) explains both.

## Where to start reading

1. `cmd/musa/main.go`: the flags and how the pieces are built.
2. `app/app.go`: the places a person moves between, and how each package is given what it needs.
3. A package you will work on: its model, then its store, then its screens.
