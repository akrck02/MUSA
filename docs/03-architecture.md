# Architecture
---

This page is for whoever wants to change MUSA: by the end you know where each thing lives and why.

## One package per idea

MUSA is one Go module. Each package is a thing the artist works with, and holds all of it: the
model, the rules, the store that reads and writes it, and its screens.

| Package     | Holds |
|-------------|-------|
| `catalogue` | Products, variants, stock and tags. |
| `events`    | Fairs: the stock taken to them, their sales and expenses. |
| `sales`     | Orders from every channel, their status, and the sales history. |
| `stand`     | The stand: the basket, «Cobrar» and where it sells from. |
| `cash`      | Paying in cash: what was handed over and the change. |
| `calendar`, `money` | Days without time zones, and euros as whole cents. |
| `storage`   | The SQLite connection, the schema and its migrations, and the artist's settings (the theme they chose). |
| `look`      | The pieces every screen shares, on top of Bubble 4. |
| `app`       | The frame, the navigation and the home page. |
| `cmd/musa`  | The program: reads its flags, opens the database, opens the window. |

Dependencies point one way: `money` and `calendar` know nothing else; `catalogue` builds on them;
`sales` on the catalogue; `events` on sales; `stand` on all three; `app` on everything.

## Screens

Screens are drawn with Go UI every frame, from the state the package keeps (immediate mode). There
are no view models: a screen reads its store, keeps what it shows, and asks the store again when
something changes.

## Errors

Every error has a code (`sales.already-delivered`) declared in its package's `errors.go`. Programs
check codes, never text. What the artist reads lives apart, in `words.go`, in Spanish.

## Ids

Every row has a UUID made on the device, so a later sync never has to renumber anything.
