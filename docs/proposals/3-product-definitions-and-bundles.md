# 3. Product definitions, bundles and orders that know their fair

> Status: proposed · Author's approval: proposal ☐ plan ☐ · Lives in: `catalogue`, `orders` (today `sales`), `events`, `demo`, `storage`

> Note (2026-10-24): bundles («packs») moved to proposal 9; R4, R5 and R8's bundles are decided there.

## Why

Today a `product` holds both the artwork and its kind, and a `product_variant` is what is sold.
The author's model names the two levels plainly — a **product definition** (the artwork, which
says whether it is digital or physical) and a **product** (a size or finish that is sold and
counted) — and adds **bundles**, the packs that analysis gap H9 asked for. Orders and fairs keep
working as today, with the links made explicit.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A product definition says once whether it is digital or physical; its products inherit it | test `a product takes its kind from its definition` |
| R2 | A product of a physical definition has physical data (stock, minimum, SKU, measures); one of a digital definition has digital data (format, turnaround); never both, never the wrong one, never none | test `a product whose data does not match its kind is not saved` |
| R3 | The kind of a definition cannot change once it has products | test `the kind of a definition with products cannot change` |
| R4 | A bundle has its own name and price and holds several products, each with a quantity; a product can be in several bundles | test `a bundle holds products with quantities` |
| R5 | Selling a bundle takes the stock of each of its physical products; deleting the order gives it back | test `selling a bundle takes the stock of its pieces` |
| R6 | An order line points to one product or one bundle and keeps a copy of its name, price and cost at the time | test `changing a price leaves past orders as they were` |
| R7 | An order may belong to a fair; the stand's remaining stock and the fair's profit read from it | existing fair tests |
| R8 | A tag has a name and a colour and labels definitions and bundles, so the stand and the catalogue group both by tag | test `a bundle shows under its tags` |
| R9 | An existing database is migrated without losing a product, a sale or a fair | test `a version 1 database opens as version 2` |
| R10 | Every new problem and failure has a code and a short sentence the artist understands (table below) | test that each constructor returns its own code; test that every problem code has words |
| R11 | A definition says what its products differ in (`variant_label`: «Tamaño», «Acabado», «Color»…); a definition with one product needs no product name | test `a single product needs no name` |
| R12 | A definition may have one photo, shown in the catalogue, the stand and the orders | test `a definition keeps its photo` |

### New error codes

Each code is a constant in its package's `errors.go`; programs compare codes, never text. The
artist's sentence lives apart, in `problemWords` (`words.go`), so it can change without breaking
anything. A **problem** is something the artist can fix in the form: `Save` returns it and the
editor shows it beside its field, in a few words. A **failure** (`*Error`: code, detail, cause)
should not happen from the screens; it is shown as a toast.

| Code | Kind | When | What the artist reads |
|---|---|---|---|
| `catalogue.empty-bundle` | problem | a bundle with no products | «Añade al menos un producto al pack» |
| `catalogue.bundle-quantity` | problem | a bundle item under 1 unit | «Mínimo 1 unidad» |
| `catalogue.kind-is-fixed` | problem | changing digital ↔ physical on a definition with products | «Ya tiene productos: no se puede cambiar el tipo» |
| `catalogue.data-does-not-match-kind` | failure | a product has the other kind's data, both or none | «Este producto no encaja con su tipo» |
| `catalogue.bundle-not-found` | failure | an order line points to a missing bundle | «Ese pack ya no existe» |
| `orders.line-without-item` | failure | an order line points to neither a product nor a bundle, or both | «Cada línea tiene que ser un producto o un pack» |

Tests: `TestEachErrorConstructorCarriesItsOwnCode` grows with the new failures; one test per
rule; and a new test, `every problem code has words for the artist`, so no raw code ever reaches
the screen.

## Out of scope

- Free-text lines at a fair (H2).
- Editing past orders (H6), currency and tax (H10).
- Order channel and status, payments, customers and projects: proposal 4.
- Showing bundles on the fair stand: they are sold from "Nueva venta" first.

## Design

```
TAG                                             name, colour
PRODUCT_DEFINITION                              name, kind (fixed), variant_label, image_path
PRODUCT_DEFINITION 1:N PRODUCT                  name (empty when it is the only one), price, cost
PRODUCT            1:1 PHYSICAL_PRODUCT_DATA    only under a physical definition
PRODUCT            1:1 DIGITAL_PRODUCT_DATA     only under a digital definition
PRODUCT_DEFINITION N:M TAG                      through DEFINITION_TAG
BUNDLE             N:M TAG                      through BUNDLE_TAG
BUNDLE             N:M PRODUCT                  through BUNDLE_ITEM (quantity); bundle has name, price
EVENT              1:N EVENT_STOCK  N:1 PRODUCT
EVENT              1:N EVENT_EXPENSE
EVENT              1:N ORDER                    event_id empty outside a fair
ORDER              1:N ORDER_LINE   N:1 PRODUCT or BUNDLE   (+ name, price, cost copied)
```

SQLite cannot check R2 across tables, so `catalogue.Store` checks it when saving, in the same
transaction.

**Sales become orders**, as in the author's model: the package `sales` becomes `orders`, `Sale`
becomes `Order` and `SaleLine` becomes `OrderLine`; error codes move from `sales.*` to `orders.*`.
The migration renames the tables `sale` and `sale_line` to `orders` and `order_line` (`ORDER` is a
reserved word in SQL, so the table takes the plural). The artist's screens keep saying «ventas»:
it is the word she uses.

_Tasks to be written once the proposal is approved._

## Result

_To be filled at the end._
