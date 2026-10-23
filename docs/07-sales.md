# Sales
---

The sales module records every sale, whatever the channel, with its products, shipping and margin, and sums them up by period and channel.

## A sale

A sale is an order. It has:

- A **channel**: Evento, Tienda online, Redes y mensajes, Marketplace, Venta local or Otro.
- A date, and an optional customer and note (an order number, the shop it came from…).
- One or more **lines**. Each line is a product variant with a quantity, a unit price and a unit cost.
- What shipping was **charged** to the customer and what it **cost**.

Names, prices and costs are copied when the sale is made, so editing or deleting a product later does not change past sales.

## How far it has travelled

An order that has to be sent has a **status**, and the status only moves forward:

| Status | Means |
|--------|-------|
| Por preparar | the order came in; the parcel is still to be made |
| Enviado | the parcel left |
| Entregado | the customer has it |

A **local sale** (Venta local, or a sale at an event) hands the goods over on the spot, so it is
delivered the moment it is made. Sales recorded before statuses existed are taken as delivered.

## Figures

| Figure | How it is computed |
|--------|--------------------|
| Total | Sum of quantity × price, plus shipping charged |
| Cost | Sum of quantity × cost, plus shipping cost |
| **Margin** | Total − cost |

The margin is negative when the sale lost money.

## Recording a sale

**New sale** opens an editor: pick the channel, the date (today by default), the customer and a note, then add products from the catalogue. Each line starts at one unit and the variant's price; both can be changed for discounts or bundles. Totals update while typing.

Physical variants leave the catalogue when the sale is saved. A sale cannot take more units than the catalogue has, counting every line of the same variant. Digital variants, such as commissions, have no stock.

## Deleting a sale

Deleting a sale gives its physical units back to the catalogue.

## Event sales

Sales made at an event are orders of one line with the Evento channel. They show in the sales list too, but they are undone from their event: their units belong to the event's stock, not to the catalogue.

## The sales list

The list shows the sales of a period (this month, last month, this year or all) and, optionally, of one channel, newest first. On top, a summary of what is shown: number of sales, units, revenue, cost and margin. Tapping a sale opens its lines, shipping and totals.

## Shipping

This module only keeps the amounts. Carriers, tracking and the state of a parcel will come with the shipping module.
