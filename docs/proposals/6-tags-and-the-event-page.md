# 6. Tags that explain themselves, and an event page like the rest

> Status: built, pending review · Author's approval: proposal ☑ plan ☑ (2026-10-24, from the mock-ups) · Lives in: `catalogue` (the product
> editor's tags), `events` (the event's page), `look`
> Depends on: proposals 1 to 5 (MUSA as it is on `feature/go-ui`)

## Why

Two screens were reported as confusing or out of place.

**Tags.** Adding a tag takes four different-looking things: the tags as pills, a «+ Añadir» pill
that opens a field, a ✓ button beside the field, and suggestions as more «+ …» pills under it.
Nothing says that tapping a tag removes it, the field and its ✓ appear out of nowhere, and the
suggestions look exactly like the tags already chosen.

**The event page.** It was ported as it was in the first version and never designed. It has a grey
tab bar unlike any other page, three tabs where one is a copy of the stand (a «Vender 1» per
product), plain «Quedan 3 de 3 · 0 vendidas» rows with bars, and the event's accounts squeezed
into a side panel. Selling has a better home already: the stand. What only this page does is
**take stock to the event**, **record its expenses**, **close it**, and **show its accounts**.

## What

### Tags

One field that holds its tags, as tag fields do in most apps:

```
Etiquetas
┌────────────────────────────────────────────────┐
│ [Pegatina ×] [Gatos ×]  Añade una etiqueta…    │
└────────────────────────────────────────────────┘
  Sugerencias: Lámina · Llavero · Otoño
```

| Id | Requirement | Check |
|---|---|---|
| R1 | The product's tags show inside one field, each as a chip with its own × to remove it | test `a tag is removed by its ×, not by tapping its name` |
| R2 | Typing in the field and pressing Enter (or a comma) adds the tag; there is no separate ✓ button and no «+ Añadir» step | test `typing a tag and pressing Enter adds it` |
| R3 | Under the field, quiet text links offer the tags used on other products that match what is typed, at most five; tapping one adds it | test `a suggestion adds its tag`; suggestions never look like chosen tags |
| R4 | A tag already on the product is never suggested again, nor added twice in other capitals | test `a tag is not added twice` (exists: `addTag`) |

### The event page

One page, built like Inicio and Pedidos: header, figures, then cards. No tabs.

```
Mercado de Otoño                          [Abrir el puesto]  [⋯]
24 oct 2026 · Paseo del Río · Abierto

[ Vendido 25 € ] [ Coste 6,40 € ] [ Gastos 35 € ] [ Beneficio −16,40 € ]

┌ Lo que llevas ─────────── [+ Llevar] ┐ ┌ Ventas ─────────────────────┐
│ ● Cactus        3 de 3                │ │ 16:51  2 × Gato astronauta  │
│ ● Gato astr. A5 2 de 4  ██░░          │ │ 16:51  1 × Luna Mate        │
│ …                                     │ └─────────────────────────────┘
└───────────────────────────────────────┘ ┌ Gastos ───────── [+ Añadir] ┐
                                          │ Stand                 35 €  │
                                          └─────────────────────────────┘
```

| Id | Requirement | Check |
|---|---|---|
| R5 | The page has the header of every other page: name, when and where, its state as a pill, «Abrir el puesto» as the main button while it is open, and «⋯» with Editar, Cerrar evento and Eliminar (in red) | test `an open event opens its stand from its page` |
| R6 | Its accounts are four figures across the top, like the home page's: sold, cost of what was sold, expenses and profit (red when below zero) | test `the event's profit is sold minus costs` (exists: `Summarise`) |
| R7 | «Lo que llevas» is one card listing each product taken, with how many are left of how many and a thin bar; tapping a row offers «Llevar más» and «Devolver al catálogo». There is no «Vender 1» here: selling is the stand's | test `stock is taken and returned from the list` |
| R8 | «Ventas» lists every sale of the event, newest first, each with its undo while the event is open | test `a sale is undone from the event's page` |
| R9 | «Gastos» lists the expenses with their own red ×, and «Añadir» opens a small form with concept and amount | test `an expense is added and removed` (exists) |
| R10 | A closed event shows the same page read-only: no «Abrir el puesto», no «Llevar», no undo, no ×; its state pill says «Cerrado» | test `a closed event changes nothing` |
| R11 | On a narrow window the cards stack in one column, in this order: lo que llevas, ventas, gastos | test at a narrow width |

## Out of scope

- The stand itself, which already follows the design.
- Changing how stock moves between the catalogue and an event, which stays as `events.Store` has it.
- Pictures of products in the lists: they come with photos, later.

## Design

- **Tags.** Go UI gains `EntryField`, a text field that tells the step Enter is pressed: generic,
  for any line sent with Enter. The tag field is a frame drawn like a field (the field's line, the
  focus colour while typing), holding a chip per tag with its own round × and a plain `EntryField`
  that grows into the rest of the row. Suggestions are `look.QuietLink`s, five at most.
- **The event page.** `events/detail_screen.go` is rewritten as one page: `look.PageHeader`,
  `look.Figures` with four figures, then two columns (`look.Columns`): the stock card on the left,
  sales and expenses cards on the right. The take, return, close and delete sheets stay as they
  are; «Abrir el puesto» hands the event to the stand through a callback the app gives.

## Tasks

- [x] T1 `EntryField` in Go UI, with its test and docs (R2)
- [x] T2 The tag field, its chips and suggestions; tests for Enter, comma, suggestion and × (R1–R4)
- [x] T3 The event page's header, figures and «Abrir el puesto» (R5, R6, R10)
- [x] T4 «Lo que llevas» with «Llevar más» and «Devolver» (R7)
- [x] T5 «Ventas» and «Gastos» cards (R8, R9)
- [x] T6 Narrow layout and a closed event (R10, R11)

## Result

- **Tags**: one field with a chip and × per tag; Enter or a comma adds; up to five quiet suggestions.
  Go UI gained `EntryField`.
- **The event page**: header with state pill, «Abrir el puesto» and «⋯»; four figures; «Lo que
  llevas» with «Llevar», and «Llevar más» / «Devolver» on a tapped row; «Ventas» with undo;
  «Gastos» with «Añadir gasto». The old tabs, «Vender 1» and register panel are gone.
- Checked by tests (`app/catalogue_flow_test.go`, `app/fairs_flow_test.go`, `app/app_test.go`),
  including a closed event and a narrow window, and by a capture compared with the mock-up.
- Still for a person: try it on the iPad; the take sheet keeps its old look.
