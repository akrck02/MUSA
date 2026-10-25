# 1. Port MUSA to Go with go-ui and Bubble 4

> Status: done · Author's approval: proposal ☑ plan ☑

## Why

MUSA is written in Kotlin Multiplatform with Compose. The author's own stack for apps is now Go:
[go-ui](https://akrck02.org/go-ui) for the interface, [go-ui-android](https://akrck02.org/go-ui-android)
for the phone and [Bubble 4](https://github.com/akrck02/bubble) for the look. Porting MUSA puts it
on that stack, so it shares its fixes, its accessibility and its design with the other projects.

## What

The Go version does everything the Kotlin version does today, in one delivery.

| Id | Requirement | Check |
|---|---|---|
| R1 | Products: digital and physical, variants, stock, cost and sale price; create, edit, duplicate, delete | tests `a product draft is validated before saving`, `a duplicated product gets new ids` |
| R2 | Tags: suggestions, search and filters on the product list | tests ported from `TagSuggesterTest`, `ProductFilterTest` |
| R3 | Events: stock moved to a fair and back on close, sales at the variant price or an edited one, undo, expenses, summary (gross, cost of goods, profit) | tests ported from `EventValidatorTest` and the event use case tests |
| R4 | Sales: orders with lines, channel, customer, note and shipping; stock taken on sale and given back on delete; list filtered by period and channel with revenue, cost and margin | tests ported from `SaleTest`, `SalesFilterTest` |
| R5 | Data lives in a local SQLite file; ids are UUIDs made on the device | test `a product survives closing and reopening the database` |
| R6 | A demo mode fills a throwaway database with toy data | test `demo data fills every module` |
| R7 | Runs on desktop (Linux, macOS, Windows) and Android | `go run ./cmd/musa`; APK built with go-ui-android and opened in the emulator |
| R8 | Looks like Bubble 4: its theme through `bubble/goui`, Lucide icons | screenshots of every screen beside the Kotlin ones |
| R9 | Money is exact (integer cents), never floating point | tests ported from `MoneyTest` |

## Out of scope

- iOS: go-ui has no iOS backend yet. The Kotlin `iosApp` goes away with the rest of the Kotlin code;
  git history keeps it.
- Reading databases made by the Kotlin version (the schema has no migrations yet, as the README warns).
- The roadmap modules not built yet (backups, accounts, marketplace, shipping, income tax).
- New Bubble components: if a screen needs one, the open Bubble session is asked for it, not invented here.

## Design

**Decisions taken with the proposal (2026-10-23):** the Kotlin code leaves in this branch (git
history keeps it); `CLAUDE.md` is rewritten for Go; SQLite comes from `modernc.org/sqlite`, a
driver in pure Go, so the phone build needs no C library of its own.

**One Go module, `github.com/Itros97/MUSA`, with folders by feature.** Each feature keeps its
logic, its storage, its screens and its tests together:

| Folder | Holds | Depends on |
|---|---|---|
| `money/` | `Money` in whole cents, parsing and formatting euros | nothing |
| `catalogue/` | products, variants, tags, validation, filters, tag suggestions, their tables, their screens | `money`, `storage` |
| `events/` | events, stock taken, expenses, summary, their screens | `catalogue` |
| `sales/` | orders, lines, channels, filters by period, summary, their screens | `catalogue`, `events` |
| `storage/` | opening the database file, the schema, transactions | the SQLite driver |
| `demo/` | the toy data and the throwaway database | every feature |
| `app/` | home, navigation between screens, the Bubble theme | every feature |
| `cmd/musa/` | `main`: flags (`-demo`, `-database`), the window | `app` |

Logic is plain functions over plain structs, tested without a window. Clocks and id generators
are passed in, as in the Kotlin version, so tests use fixed values. Every error carries a code
(`catalogue.sku-taken`), declared in an `errors.go` per package.

**Screens** are go-ui's immediate-mode controls, themed by `bubble/goui` (Bubble 4, Paper, light
or dark with the system). Icons are go-ui's Lucide names. On a phone the same program is packaged
with `go-ui-apk` from Go UI for Android; the layout switches from a side index to a bottom bar
when the window is narrow. If a screen needs a piece neither go-ui nor Bubble has, the Bubble
session is asked for it.

**The schema** is the same tables as the SQLDelight files, written once in `storage/schema.sql`
and run on open. Soft deletes, UUID ids and integer cents stay, ready for the sync planned later.

**Not taken:** keeping both versions side by side (two codebases drift); go-ui-android's Material
pieces as the look (the request is Bubble 4; Material stays for what Bubble lacks); the cgo
driver `mattn/go-sqlite3` (needs a C compiler for every target; kept as the fallback if the pure
Go driver fails on Android).

## Tasks

- [x] T1 — Go module, `money/`, CI running `gofmt`, `go vet` and the tests · serves R9 · tests ported from `MoneyTest`
- [x] T2 — `storage/`: open the file, schema, transactions · serves R5 · test `a product survives closing and reopening the database`
- [x] T3 — `catalogue/` logic: models, validation, duplicate, tag normalising and suggestions, filters · R1, R2 · tests ported from `ProductValidator`, `TagSuggesterTest`, `ProductFilterTest`, `TextAndTagTest`
- [x] T4 — `catalogue/` storage: save, read, delete, stock baseline kept · R1, R5 · tests ported from the save and duplicate use cases
- [x] T5 — `events/` logic and storage: open, take and return stock, sell, undo, expenses, close, summary · R3 · tests ported from `EventValidatorTest` and the event use cases
- [x] T6 — `sales/` logic and storage: orders, stock taken and given back, filters, summary · R4 · tests ported from `SaleTest`, `SalesFilterTest`, `SaleTimeTest`
- [x] T7 — `demo/`: toy catalogue, events and sales, wipe · R6 · test `demo data fills every module`
- [x] T8 — `app/` and `cmd/musa`: theme, home, navigation, narrow and wide layouts · R7, R8 · capture of the home screen
- [x] T9 — catalogue screens: list with search, tags and filters; product editor · R1, R2 · go-ui `uitest` of creating a product
- [x] T10 — event screens: list, editor, detail with stock, sales and expenses · R3 · `uitest` of selling at a fair
- [x] T11 — sales screens: list with period and channel, order editor, detail · R4 · `uitest` of recording an order
- [x] T12 — Android: APK with `go-ui-apk`, opened in the emulator · R7 · screenshot from the emulator
- [x] T13 — remove Kotlin, rewrite `README.md`, `CLAUDE.md` and `docs/`, screenshots beside the Kotlin ones · R8 · `check_repository.py`

## Result

_To be filled at the end._
