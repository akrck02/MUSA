# 6. Tags that explain themselves, and an event page like the rest

> Status: proposed · Author's approval: proposal ☐ plan ☐ · Lives in: `catalogue` (the product
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

*To fill after the author approves the proposal.*

## Tasks

*To fill with the design.*

## Result

*To fill at the end.*
