# Inventory
---

The inventory module keeps the catalogue: what the artist sells, in which versions, at what price and how much is left.

## Products and variants

A product is either **physical** or **digital**, and always has at least one variant.

- A **physical product** has a material, and its variants carry SKU, stock, minimum stock, weight and width, height and depth.
- A **digital product** has a delivery format and an optional turnaround in days. Its variants are levels or formats, such as "sketch" or "full colour".

Each variant has its own price and cost. Money is stored as whole cents to avoid rounding errors.
The margin is price minus cost, also shown as a percentage of the price. Under the price and
the cost, the editor says what each piece earns as you type, and warns in red when the cost is
above the price; it still lets you save, for a piece you sell at a loss on purpose.

## Rules on save

A draft is checked before it is saved. It is rejected when:

- The product name is blank, or it has no variants.
- There are two or more variants and one has no name, or two share a name (case and accents ignored).
- A price, cost, stock, minimum stock or turnaround is negative.
- A weight or size is zero or negative.
- A SKU repeats, inside the product or in another product.

## Low stock

A physical variant is low on stock when it is sold out, or when its stock is at or below its minimum.
A product is low on stock when any of its variants is.

## Mis productos

The list has two halves, **Físico** and **Digital**, each a tab with how many products it holds.
Under them, a tab per tag narrows the list (with **Archivados** last, when something is archived),
and the magnifier opens a search over names, descriptions and SKUs that ignores case and accents.
Results are sorted by name.

On the physical side, a note says how many versions need restocking and names the first few
(«A3 de Lámina «Gato»»); «Ver cuáles» shows only those products.

Each card is a coloured picture (the first tag's colour and the work's first letter, until products
have photos), the name, the prices from cheapest to dearest and one pill per version with what is
left: warm when it runs low, red when it is sold out. Tapping a pill changes that stock.

## The product's page

The picture, the name and the tags sit on the left; the kind is shown with a lock, because it never
changes. On the right, **Versiones** is a table: each version's name, price, what it costs the
artist and its stock, with − and + and a «Pocos» warning. Everything set once (description,
status, material or delivery, each version's SKU, warning level, weight and size, and the margin)
is folded under **Más detalles**.

## Duplicating

Duplicating a product creates a new draft with new ids. On physical products the stock is reset
to zero and SKUs are cleared, because a SKU must be unique.
