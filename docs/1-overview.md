# Overview
---

MUSA is a local-first inventory and business app for digital artists who also sell physical stock.
It is written in Go with Go UI and Bubble 4, and runs on the desktop and on Android; the iPad is the device it is designed for.

## What it does today

The first module is **inventory**. It keeps the artist's catalogue in a local SQLite database:

- **Products** of two kinds: physical stock (prints, stickers, merch) and digital work made to order (commissions).
- **Variants** for each product, each with its own price and cost. Physical variants also carry SKU, stock, minimum stock, weight and size.
- **Tags** with a colour from a fixed palette, used to group and filter the catalogue.
- **Statuses**: draft, active and archived.

The second module is **events**: fairs and markets where stock is taken from the catalogue, sold on the spot, and summed up into gross, cost, expenses and profit.

The third module is **sales**: orders from every channel (online shop, social media, marketplaces, in person and events), with their products, shipping amounts and margin.

## What it does not do

- It has no server and no account. One device belongs to one artist.
- It does not sync between devices yet. See the [roadmap](6-roadmap.md).

## Read next

- [Get started](2-get-started.md) to build and run it.
- [Core ideas](3-core-ideas.md) and [contributing](8-contributing/1-the-code.md) to change the code.
- [Inventory](4-guides/1-inventory.md) for the rules the inventory module follows.
- [Events](4-guides/2-events.md) for how stock, sales and profit work at a fair.
- [Sales](4-guides/3-sales.md) for orders, channels, shipping and margin.
- [The stand](4-guides/4-stand.md) for selling at a fair or from the whole stock.
- [Projects](4-guides/5-projects.md) for commissions, customers and their networks.
