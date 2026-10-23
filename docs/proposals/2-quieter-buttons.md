# 2. Quieter buttons in Bubble 4 Paper

> Status: proposed · Author's approval: proposal ☐ plan ☐ · Lives in: Bubble (`bubble/goui`), asked for by MUSA

## Why

MUSA's screens are now flat and calm: one card per idea, soft titles, a slate accent. The buttons
are what still shout. Every button has a bold label; the main one stands on a dark edge, like a
key you press; the outlined ones draw a dark, heavy line that is louder than the cards around it.
On the fair stand a row of identical outlined buttons turns a grid of products into a grid of
buttons. The look comes from Bubble's theme, so the fix belongs in Bubble, for every app on it.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | The main button is flat: a fill of the accent, no darker edge under it | capture of a page with `PrimaryButton`; `PrimaryEdge` unused or `EdgeWidth` 0 in Paper |
| R2 | The outlined button's line is the soft `Line` of cards, not `ControlLine`; it darkens on hover and focus | capture of `IconButton` at rest and hovered |
| R3 | Button labels are medium weight, not bold; titles stay bold | `goui.Theme(...).ButtonWeight == draw2d.Medium`; test beside the theme |
| R4 | A new **tonal** button: a pale tint of the accent (`PrimaryTint`) with accent ink and no line, for actions repeated in a list or grid | Bubble component `TonalButton`/`TonalIconButton`; capture of the fair stand |
| R5 | Contrast stays at Bubble's bar: label to fill 4.5:1, control edge and focus ring 3:1, in light and dark | Bubble's contrast tests extended to the tonal pair and the new edge |
| R6 | Every touch target stays at least 44 × 44 px | existing size tests |

## Out of scope

- The Material style: it keeps its own buttons.
- Glass: it keeps its translucent panes; R1–R4 apply to Paper first.
- New shapes or radii: today's are right.

## Design

The four changes in `goui/theme.go`, plus one component:

| Today | Proposed |
|---|---|
| `ButtonWeight = draw2d.Bold` | `draw2d.Medium` (go-ui's own default already is) |
| `PrimaryEdge = AccentEdge`, `EdgeWidth = shape.EdgeWidth` | `EdgeWidth = 0` in Paper: the fill alone marks the main action |
| outline drawn in `ControlLine` | outline in `Line`; `ControlLine` only on hover and focus |
| — | `TonalButton`: fill `PrimaryTint`, ink `Primary`, no line, same size and focus ring |

Use in MUSA, once Bubble has it: one `PrimaryButton` per screen (Guardar, Abrir el puesto);
outlined buttons for the rest; tonal buttons where an action repeats per item ("Vender 1 · 8 €"
on the stand, "Añadir" on tags).

**Not taken:** styling buttons inside MUSA (it would drift from Bubble and from every other app);
dropping the outline altogether, as a text-only button (on Paper it no longer looks pressable, and
a finger needs to see where to tap).

## Tasks

- [ ] T1 — Bubble: `ButtonWeight` medium, no primary edge in Paper · R1, R3 · theme test, captures
- [ ] T2 — Bubble: soft outline, darker on hover and focus · R2, R5 · contrast test, captures
- [ ] T3 — Bubble: `TonalButton` and `TonalIconButton` in go-ui or `bubble/goui` · R4, R5, R6 · contrast and size tests
- [ ] T4 — MUSA: tonal buttons on the stand and the tag field; one primary per screen · R4 · `uitest` of selling at a fair, captures

## Result

_To be filled at the end._
